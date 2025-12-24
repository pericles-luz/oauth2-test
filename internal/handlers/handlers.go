package handlers

import (
	"html/template"

	"github.com/gorilla/sessions"

	"github.com/pericles-luz/oauth2-test/internal/services"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	sessionStore   *sessions.CookieStore
	historyService *services.HistoryService
	templates      *template.Template
	baseURL        string
}

// NewHandlers creates a new Handlers instance
func NewHandlers(
	sessionStore *sessions.CookieStore,
	historyService *services.HistoryService,
	templates *template.Template,
	baseURL string,
) *Handlers {
	return &Handlers{
		sessionStore:   sessionStore,
		historyService: historyService,
		templates:      templates,
		baseURL:        baseURL,
	}
}

// Session keys
const (
	SessionName     = "oauth-session"
	KeyClientID     = "client_id"
	KeyClientSecret = "client_secret"
	KeyRedirectURI  = "redirect_uri"
	KeyScopes       = "scopes"
	KeyCodeVerifier = "code_verifier"
	KeyState        = "state"
	KeyAccessToken  = "access_token"
	KeyRefreshToken = "refresh_token"
	KeyIDToken      = "id_token"
)
