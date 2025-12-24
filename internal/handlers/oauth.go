package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/pericles-luz/oauth2-test/internal/models"
	"github.com/pericles-luz/oauth2-test/internal/services"
)

// OAuthLogin initiates the OAuth2 authorization flow
func (h *Handlers) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, SessionName)

	// Get OAuth config from session
	clientID, _ := session.Values[KeyClientID].(string)
	clientSecret, _ := session.Values[KeyClientSecret].(string)
	redirectURI, _ := session.Values[KeyRedirectURI].(string)
	scopesStr, _ := session.Values[KeyScopes].(string)

	if clientID == "" || clientSecret == "" || redirectURI == "" {
		http.Error(w, "OAuth2 configuration not found. Please configure first.", http.StatusBadRequest)
		return
	}

	// Parse scopes
	scopes := strings.Split(scopesStr, " ")

	// Create OAuth config
	oauthConfig := &models.OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		Scopes:       scopes,
		BaseURL:      h.baseURL,
	}

	// Validate config
	if err := oauthConfig.Validate(); err != nil {
		http.Error(w, "Invalid OAuth configuration: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create OAuth service
	oauthService := services.NewOAuthService(oauthConfig, h.historyService)

	// Generate state for CSRF protection
	state, err := services.GenerateRandomState()
	if err != nil {
		log.Printf("Failed to generate state: %v", err)
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Generate auth URL with PKCE
	authURL, verifier, err := oauthService.GenerateAuthURL(state)
	if err != nil {
		log.Printf("Failed to generate auth URL: %v", err)
		http.Error(w, "Failed to generate authorization URL", http.StatusInternalServerError)
		return
	}

	// Store state and verifier in session
	session.Values[KeyState] = state
	session.Values[KeyCodeVerifier] = verifier
	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	log.Printf("Redirecting to authorization URL: %s", authURL)

	// Redirect to authorization URL
	http.Redirect(w, r, authURL, http.StatusFound)
}

// OAuthCallback handles the OAuth2 callback
func (h *Handlers) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, SessionName)

	// Check for OAuth errors
	if errorCode := r.URL.Query().Get("error"); errorCode != "" {
		errorDesc := r.URL.Query().Get("error_description")
		log.Printf("OAuth error: %s - %s", errorCode, errorDesc)
		http.Error(w, "OAuth error: "+errorCode+" - "+errorDesc, http.StatusBadRequest)
		return
	}

	// Get state from query
	state := r.URL.Query().Get("state")
	expectedState, _ := session.Values[KeyState].(string)

	// Verify state (CSRF protection)
	if state != expectedState || state == "" {
		log.Printf("State mismatch: expected=%s, got=%s", expectedState, state)
		http.Error(w, "Invalid state parameter (CSRF check failed)", http.StatusBadRequest)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Get code verifier from session
	verifier, _ := session.Values[KeyCodeVerifier].(string)
	if verifier == "" {
		http.Error(w, "Code verifier not found in session", http.StatusBadRequest)
		return
	}

	// Get OAuth config from session
	clientID, _ := session.Values[KeyClientID].(string)
	clientSecret, _ := session.Values[KeyClientSecret].(string)
	redirectURI, _ := session.Values[KeyRedirectURI].(string)
	scopesStr, _ := session.Values[KeyScopes].(string)
	scopes := strings.Split(scopesStr, " ")

	oauthConfig := &models.OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		Scopes:       scopes,
		BaseURL:      h.baseURL,
	}

	// Create OAuth service
	oauthService := services.NewOAuthService(oauthConfig, h.historyService)

	// Exchange code for tokens
	log.Printf("Exchanging code for tokens...")
	token, err := oauthService.ExchangeCode(code, verifier)
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Token exchange successful! Access token: %s...", token.AccessToken[:20])

	// Store tokens in session
	session.Values[KeyAccessToken] = token.AccessToken
	session.Values[KeyRefreshToken] = token.RefreshToken
	if idToken, ok := token.Extra("id_token").(string); ok {
		session.Values[KeyIDToken] = idToken
	}

	// Get user info
	log.Printf("Fetching user info...")
	userInfo, err := oauthService.GetUserInfo(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		// Don't fail, just log the error
	}

	// Store user info in session
	if userInfo != nil {
		session.Values["user_info"] = userInfo
		log.Printf("User info retrieved: %s (%s)", userInfo.Name, userInfo.CPF)
	}

	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
