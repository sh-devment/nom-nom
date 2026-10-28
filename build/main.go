package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var buildTime = "unknown"

// defaultAppName is the fallback when APP_NAME is unset, so a service file that
// forgets the variable still starts and still gets nom-nom.db rather than a
// database named after the empty string.
const defaultAppName = "nom-nom"

// appName is the app's runtime identity: the --version banner, the start log
// line, and (when DB_PATH is unset) the database filename all derive from it.
var appName string

//go:embed web
var webFiles embed.FS

var (
	authURL      string
	authInternal string
	appURL       string
	appToken     string
	tmpl         *template.Template
	httpClient   = &http.Client{}
)

type pageData struct {
	User       *User
	Error      string
	Version    string // buildTime — cache-bust token for static assets (?v=)
	*regionDef        // Domain, Lang, Text… — see region.go
}

func initTemplate() {
	src, err := webFiles.ReadFile("web/index.html")
	if err != nil {
		log.Fatalf("web/index.html not found: %v", err)
	}
	tmpl = template.Must(template.New("index").Parse(string(src)))
}

// ── request logging ───────────────────────────────────────────────────────────

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Millisecond))
	})
}

// cacheStatic wraps a handler with a 30-day immutable cache header.
func cacheStatic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=2592000, immutable") // 30 days; URL is busted per-deploy via ?v=
		h.ServeHTTP(w, r)
	})
}

// ── app routes ────────────────────────────────────────────────────────────────
// Add your app-specific handlers here.

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if code := r.URL.Query().Get("code"); code != "" {
		handleCallback(w, r, code)
		return
	}
	var user *User
	if uid := sessionUserID(r); uid != 0 {
		user, _ = getUserByID(uid)
	}
	// HTML must always revalidate so the fresh ?v= asset token reaches the browser.
	w.Header().Set("Cache-Control", "no-cache")
	tmpl.Execute(w, pageData{User: user, Version: buildTime, regionDef: region}) //nolint:errcheck
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	godotenv.Load() //nolint:errcheck

	// Resolved before anything else: --version prints it, and the database is
	// named after it. Everything downstream reads appName, never a literal.
	appName = os.Getenv("APP_NAME")
	if appName == "" {
		appName = defaultAppName
	}

	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "--info") {
		fmt.Printf("%s built: %s\n", appName, buildTime)
		os.Exit(0)
	}

	// An explicit DB_PATH always wins — that is how dev-compose puts the file on
	// the /data volume (modernc sqlite locking fails on macOS virtiofs). With
	// nothing set the name follows the app instead of a literal, so a unit file
	// copied from another app cannot leave this one quietly writing into
	// auth-client.db. db.go reads DB_PATH, so setting it here keeps that
	// template file untouched and makes its own fallback unreachable.
	if os.Getenv("DB_PATH") == "" {
		os.Setenv("DB_PATH", appName+".db") //nolint:errcheck
	}

	authURL = os.Getenv("AUTH_URL")
	authInternal = os.Getenv("AUTH_INTERNAL")
	appURL = os.Getenv("APP_URL")
	appToken = os.Getenv("APP_TOKEN")
	initRegion()

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		secretKey = "dev-secret"
	}
	jwtSecret = []byte(secretKey)

	aiProvider = os.Getenv("AI_PROVIDER")
	if aiProvider == "" {
		aiProvider = "gemini"
	}
	aiModel = os.Getenv("AI_MODEL")
	aiKey = os.Getenv("AI_KEY")
	aiGatewayURL = strings.TrimRight(os.Getenv("AI_GATEWAY_URL"), "/")
	aiGatewayToken = os.Getenv("AI_GATEWAY_TOKEN")

	// Free AI ops granted to a new user. Must be set before initDB() — appMigrate()
	// bakes it into the users AFTER INSERT trigger.
	if v := os.Getenv("FREE_SCANS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			log.Fatalf("FREE_SCANS must be a non-negative integer, got %q", v)
		}
		freeScansDefault = n
	}

	initDB()

	// Admin CLI (run inside the container / on the server): nom-nom --set-role <auth_id> pro
	if len(os.Args) > 1 {
		runAdmin(os.Args[1:])
		return
	}

	initTemplate()

	webFS, _ := fs.Sub(webFiles, "web")
	fileServer := http.FileServer(http.FS(webFS))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("GET /login", handleLogin)
	mux.HandleFunc("GET /logout", handleLogout)
	mux.HandleFunc("GET /apps", handleOpenApps)
	mux.Handle("GET /favicon.svg", cacheStatic(fileServer))
	mux.Handle("GET /apple-touch-icon.png", cacheStatic(fileServer))
	mux.Handle("GET /style.css", cacheStatic(fileServer))
	mux.Handle("GET /script.js", cacheStatic(fileServer))

	// nom-nom app API — all behind requireAuth (so sessionUserID is always set).
	api := func(h http.HandlerFunc) http.Handler { return requireAuth(h) }
	mux.Handle("GET /api/state", api(handleState))
	mux.Handle("POST /api/meal", api(handleMealAdd))
	mux.Handle("POST /api/meal/{id}", api(handleMealUpdate))
	mux.Handle("DELETE /api/meal/{id}", api(handleMealDelete))
	mux.Handle("POST /api/weight", api(handleWeightSet))
	mux.Handle("GET /api/weight/graph", api(handleWeightGraph))
	mux.Handle("POST /api/goals", api(handleGoalsSet))
	mux.Handle("POST /api/favorite", api(handleFavoriteAdd))
	mux.Handle("DELETE /api/favorite/{id}", api(handleFavoriteDelete))
	mux.Handle("POST /api/scan", api(handleScan))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8890"
	}
	log.Printf("start app=%s db=%s port=%s", appName, os.Getenv("DB_PATH"), port)
	log.Fatal(http.ListenAndServe(":"+port, logMiddleware(mux)))
}
