package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"

	"github.com/pericles-luz/oauth2-test/internal/handlers"
	"github.com/pericles-luz/oauth2-test/internal/services"
	"github.com/pericles-luz/oauth2-test/internal/storage"
)

func main() {
	// Load configuration from environment
	config := loadConfig()

	// Initialize database
	db, err := storage.NewSQLiteDB(config.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	migrationPath := filepath.Join("migrations", "001_initial.sql")
	if err := db.RunMigrations(migrationPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed successfully")

	// Initialize session store
	sessionStore := sessions.NewCookieStore([]byte(config.SessionSecret))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	// Initialize services
	historyService := services.NewHistoryService(db)

	// Initialize templates
	tmpl := loadTemplates()

	// Initialize handlers
	h := handlers.NewHandlers(
		sessionStore,
		historyService,
		tmpl,
		config.BaseURL,
	)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(handlers.LoggingMiddleware)
	r.Use(handlers.RecoveryMiddleware)
	r.Use(middleware.Compress(5))
	r.Use(handlers.HTMXMiddleware)

	// Static files
	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Routes
	setupRoutes(r, h)

	// Start server
	port := config.ServerPort
	log.Printf("Server starting on port %s", port)
	log.Printf("Visit http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Config holds application configuration
type Config struct {
	BaseURL       string
	SessionSecret string
	ServerPort    string
	DatabasePath  string
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	return &Config{
		BaseURL:       getEnv("OAUTH2_BASE_URL", "https://api.sindireceita.org.br"),
		SessionSecret: getEnv("SESSION_SECRET", "change-this-secret-in-production-32bytes!!"),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		DatabasePath:  getEnv("DATABASE_PATH", "./oauth2-test.db"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// loadTemplates loads all HTML templates
func loadTemplates() *template.Template {
	tmpl := template.New("")

	// Define custom template functions
	funcMap := template.FuncMap{
		"substr": func(s string, start, length int) string {
			if start < 0 || start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
	}
	tmpl.Funcs(funcMap)

	// Load all template files
	templatePattern := filepath.Join("templates", "*.html")
	tmpl = template.Must(tmpl.ParseGlob(templatePattern))

	// Load endpoint templates
	endpointPattern := filepath.Join("templates", "endpoints", "*.html")
	if matches, _ := filepath.Glob(endpointPattern); len(matches) > 0 {
		tmpl = template.Must(tmpl.ParseGlob(endpointPattern))
	}

	log.Println("Templates loaded successfully")
	return tmpl
}

// setupRoutes configures all application routes
func setupRoutes(r chi.Router, h *handlers.Handlers) {
	// Home page
	r.Get("/", h.Home)
	r.Post("/config", h.SaveConfig)

	// OAuth flow
	r.Get("/auth/login", h.OAuthLogin)
	r.Get("/auth/callback", h.OAuthCallback)

	// Dashboard (post-auth)
	r.Get("/dashboard", h.Dashboard)

	// Endpoint testing
	r.Post("/test/refresh", h.TestRefresh)
	r.Post("/test/revoke", h.TestRevoke)
	r.Get("/test/jwks", h.TestJWKS)
	r.Get("/test/discovery", h.TestDiscovery)

	// History
	r.Get("/history", h.HistoryList)
	r.Get("/history/{id}", h.HistoryDetail)
}
