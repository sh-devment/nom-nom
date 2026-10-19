package main

// region.go — what differs between the ru and com deployments apart from env
// config: the zone's domain, the language of every user-facing string, and the
// features the zone offers. The values live in region-ru.go and region-com.go.
// Technical text (logs, "bad request"-style errors) is English in both zones
// and stays out of here. Infrastructure URLs (auth-center) stay in env.

import (
	"fmt"
	"log"
	"os"
	"reflect"
)

type regionDef struct {
	Domain string // sh-development.<zone>: the login link and the menu redirect
	Lang   string // <html lang>
	Locale string // dates in script.js (toLocaleDateString)

	// AI scans (photo and text) are offered. Off: both source buttons open a
	// popup saying the feature is unavailable, and POST /api/scan refuses 403.
	AIScan bool

	// Language the AI names a scanned dish in (goes into the prompt, in English).
	AIFoodLang string

	Text   pageText
	Server serverText
}

// pageText holds the strings of web/index.html. Fields, not a map: a typo in
// the template fails at render instead of silently printing nothing.
type pageText struct {
	// login screen
	Login string // login button
	About string // text under the login card

	// header
	TabFood   string
	TabWeight string
	TabInfo   string
	Apps      string // profile menu
	Readme    string
	Logout    string

	// food panel
	KcalToday string // under the donut total
	ProtUnit  string // "<n> <ProtUnit>" under the donut total
	Manual    string // source buttons
	AIPhoto   string
	AIText    string
	Favorites string // source button and the picker title
	TodayHead string // meals table head
	Kcal      string // kcal unit: table heads, meal card

	// weight panel
	Week        string
	Month       string
	Year        string
	NoData      string // graph without points
	WeightToday string
	Kg          string
	Save        string
	Date        string // history table head

	// info panel
	Name       string
	YourID     string
	Status     string
	ScansToday string
	BuyPro     string
	ProText    string // the buy-PRO popup (no shop yet)

	// meal card
	MealName string // placeholder
	Grams    string
	Prot     string
	Fat      string
	Carb     string

	// daily goals popup
	GoalsTitle string
	GoalKcal   string
	GoalProt   string

	// AI popups
	DescribeHint string // AI text placeholder
	PointCamera  string // AI photo, camera unavailable
	Analyzing    string // full-screen loader

	// favorites picker
	FavSearch      string // placeholder
	FavSearchLabel string // aria-label

	OK string // popup button

	JS jsText // strings set by script.js
}

// jsText is rendered into the page as JSON (window.TEXT) and read by script.js.
type jsText struct {
	Error         string `json:"error"` // a failed request without a server message
	NewDay        string `json:"newDay"`
	NoMeals       string `json:"noMeals"`
	GoalKcalRange string `json:"goalKcalRange"`
	GoalProtRange string `json:"goalProtRange"`
	GoalsSaved    string `json:"goalsSaved"`
	EmailCopied   string `json:"emailCopied"`
	Today         string `json:"today"`
	Edit          string `json:"edit"`
	NoWeights     string `json:"noWeights"`
	WeightPrompt  string `json:"weightPrompt"` // %s = the date
	EnterWeight   string `json:"enterWeight"`
	WeightSaved   string `json:"weightSaved"`
	ScanSaved     string `json:"scanSaved"`
	DefaultMeal   string `json:"defaultMeal"` // a meal saved without a name (server uses it too)
	MealDeleted   string `json:"mealDeleted"`
	NoFavorites   string `json:"noFavorites"`
	NotFound      string `json:"notFound"`
	GramsUnit     string `json:"gramsUnit"`
	PickFirst     string `json:"pickFirst"`
	ScanFailed    string `json:"scanFailed"` // also the error popup's initial text
	NoCamera      string `json:"noCamera"`
	DescribeMeal  string `json:"describeMeal"`
	Unavailable   string `json:"unavailable"` // a feature the region does not offer
}

// serverText is what the Go handlers send to the user.
type serverText struct {
	DailyLimit   string // 429 on /api/scan
	NoFreeScans  string // 403 on /api/scan
	ServiceDown  string // 402 on /api/scan: AI quota/billing
	AITimeout    string // 504 on /api/scan
	StatusTester string // Info screen; %d = daily limit
	StatusFree   string // Info screen; %d = free scans left
}

var regions = map[string]*regionDef{
	"ru":  &regionRU,
	"com": &regionCOM,
}

// Active region, set once by initRegion before anything serves.
var (
	regionName string
	region     *regionDef
	domain     string // region.Domain; auth-server.go builds the menu redirect from it
)

// initRegion picks the region by REGION (ru | com, set by the deploy matrix).
// Empty means local dev and falls back to ru; anything else is a config error
// and stops the process rather than serving the wrong zone — the health check
// then fails and deploy.sh rolls back. A string forgotten in either region is
// fatal too, whichever region this instance serves.
func initRegion() {
	for name, r := range regions {
		if err := checkFilled(reflect.ValueOf(*r), "region"); err != nil {
			log.Fatalf("region %s: %v", name, err)
		}
	}

	regionName = os.Getenv("REGION")
	if regionName == "" {
		regionName = "ru"
	}
	r, ok := regions[regionName]
	if !ok {
		log.Fatalf("unknown REGION %q (want ru or com)", regionName)
	}
	region = r
	domain = r.Domain
	log.Printf("region name=%s domain=%s lang=%s ai_scan=%t", regionName, domain, region.Lang, region.AIScan)
}

// checkFilled walks structs and fails on the first empty string.
func checkFilled(v reflect.Value, path string) error {
	switch v.Kind() {
	case reflect.String:
		if v.Len() == 0 {
			return fmt.Errorf("%s is empty", path)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := checkFilled(v.Field(i), path+"."+v.Type().Field(i).Name); err != nil {
				return err
			}
		}
	}
	return nil
}
