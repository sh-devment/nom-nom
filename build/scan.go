package main

// scan.go — AI nutrition estimation. Both modes (photo, free-text description)
// return the same nutrition JSON, which the UI uses to prefill the meal sheet for
// the user to review and save.
//
// The provider is picked per deployment (each region has its own GitHub
// environment): AI_PROVIDER + AI_MODEL + AI_KEY. Only "gemini" exists for now;
// another provider is one more case in scanAI.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Set from env in main(). The model is never hardcoded, so any model can be
// A/B tested without a rebuild.
var (
	aiProvider string // AI_PROVIDER: gemini
	aiModel    string // AI_MODEL, e.g. gemini-2.5-flash
	aiKey      string // AI_KEY: the provider's API key
)

var errNoCredits = errors.New("no credits")

const scanSystemPrompt = `You are a nutrition expert. Estimate the nutritional content of the food for the visible/described portion.
Respond with ONLY a valid JSON object in this exact format, no other text:
{"name": "food name in Russian", "grams": estimated_grams, "calories": estimated_calories, "proteins": estimated_proteins_g, "fats": estimated_fats_g, "carbs": estimated_carbs_g}`

// ScanResult is what the frontend meal sheet consumes (mirrors MealInput + carbs).
type ScanResult struct {
	Name  string  `json:"name"`
	Grams int     `json:"grams"`
	Kcal  int     `json:"kcal"`
	Prot  float64 `json:"prot"`
	Fat   float64 `json:"fat"`
	Carb  float64 `json:"carb"`
}

// scanInput is the provider-neutral request: an image (base64) with a prompt,
// or a text description alone.
type scanInput struct {
	Image     string
	MediaType string
	Text      string
}

func photoInput(imageBase64, mediaType string) scanInput {
	if mediaType == "" {
		mediaType = "image/jpeg"
	}
	return scanInput{
		Image:     imageBase64,
		MediaType: mediaType,
		Text:      "What food is this? Estimate name, grams, calories and macros.",
	}
}

func textInput(description string) scanInput {
	return scanInput{Text: "Estimate the nutrition for this meal: " + description}
}

// scanTimeout bounds one AI call — without it a hung upstream request would
// freeze the scan forever.
const scanTimeout = 20 * time.Second

// scanAI asks the configured provider and parses the nutrition JSON out of the
// reply. The context ties the upstream call to the HTTP request: if the browser
// gives up and aborts, the AI call is canceled too instead of burning tokens.
func scanAI(ctx context.Context, in scanInput) (*ScanResult, error) {
	if aiModel == "" || aiKey == "" {
		return nil, fmt.Errorf("AI_MODEL / AI_KEY not set")
	}

	ctx, cancel := context.WithTimeout(ctx, scanTimeout)
	defer cancel()

	var text string
	var err error
	switch aiProvider {
	case "gemini":
		text, err = askGemini(ctx, in)
	default:
		return nil, fmt.Errorf("unknown AI_PROVIDER %q", aiProvider)
	}
	if err != nil {
		return nil, err
	}
	return parseScan(text)
}

const geminiAPI = "https://generativelanguage.googleapis.com/v1beta/models/"

// askGemini calls generateContent and returns the reply text.
func askGemini(ctx context.Context, in scanInput) (string, error) {
	var parts []map[string]any
	if in.Image != "" {
		parts = append(parts, map[string]any{
			"inlineData": map[string]any{"mimeType": in.MediaType, "data": in.Image},
		})
	}
	parts = append(parts, map[string]any{"text": in.Text})

	payload := map[string]any{
		"systemInstruction": map[string]any{
			"parts": []map[string]any{{"text": scanSystemPrompt}},
		},
		"contents": []map[string]any{
			{"role": "user", "parts": parts},
		},
		"generationConfig": map[string]any{
			"responseMimeType": "application/json",
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", geminiAPI+aiModel+":generateContent", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", aiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("gemini error status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(b)))
		// 429 RESOURCE_EXHAUSTED — the key's quota or billing ran out.
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusPaymentRequired {
			return "", errNoCredits
		}
		return "", fmt.Errorf("gemini api error %d", resp.StatusCode)
	}

	var gr struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text    string `json:"text"`
					Thought bool   `json:"thought"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return "", err
	}
	if len(gr.Candidates) == 0 {
		return "", fmt.Errorf("empty response from gemini")
	}
	var sb strings.Builder
	for _, p := range gr.Candidates[0].Content.Parts {
		if !p.Thought {
			sb.WriteString(p.Text)
		}
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("empty response from gemini finish=%s", gr.Candidates[0].FinishReason)
	}
	return sb.String(), nil
}

// parseScan pulls the JSON object out of the model's reply and maps the
// prompt's key names onto ScanResult.
func parseScan(reply string) (*ScanResult, error) {
	text := strings.TrimSpace(reply)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || end < start {
		return nil, fmt.Errorf("no JSON in response")
	}

	var raw struct {
		Name     string  `json:"name"`
		Grams    float64 `json:"grams"`
		Calories float64 `json:"calories"`
		Proteins float64 `json:"proteins"`
		Fats     float64 `json:"fats"`
		Carbs    float64 `json:"carbs"`
	}
	if err := json.Unmarshal([]byte(text[start:end+1]), &raw); err != nil {
		return nil, fmt.Errorf("parse result: %w", err)
	}
	return &ScanResult{
		Name:  raw.Name,
		Grams: int(raw.Grams),
		Kcal:  int(raw.Calories),
		Prot:  raw.Proteins,
		Fat:   raw.Fats,
		Carb:  raw.Carbs,
	}, nil
}
