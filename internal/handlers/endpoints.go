package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/pericles-luz/oauth2-test/internal/models"
	"github.com/pericles-luz/oauth2-test/internal/services"
)

// TestRefresh tests token refresh functionality
func (h *Handlers) TestRefresh(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, SessionName)

	// Get session ID
	sessionID, _ := session.Values[KeySessionID].(string)
	if sessionID == "" {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Get token from token store
	tokenStore := services.GetTokenStore()
	token, userInfo, ok := tokenStore.Get(sessionID)
	if !ok || token == nil {
		http.Error(w, "Session expired", http.StatusUnauthorized)
		return
	}

	if token.RefreshToken == "" {
		http.Error(w, "No refresh token available", http.StatusBadRequest)
		return
	}

	// Get OAuth config
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

	oauthService := services.NewOAuthService(oauthConfig, h.historyService)

	// Refresh token
	newToken, err := oauthService.RefreshToken(token.RefreshToken)
	if err != nil {
		log.Printf("Token refresh failed: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="error">Token refresh falhou: ` + err.Error() + `</div>`))
		return
	}

	// Update token store with new tokens
	tokenStore.Store(sessionID, newToken, userInfo)

	// Return success with new token info
	w.Header().Set("Content-Type", "text/html")
	tokenJSON, _ := json.MarshalIndent(map[string]interface{}{
		"access_token":  newToken.AccessToken,
		"refresh_token": newToken.RefreshToken,
		"expires_in":    newToken.Expiry,
		"token_type":    "Bearer",
	}, "", "  ")

	w.Write([]byte(`
		<div class="success">
			✓ Token atualizado com sucesso!
			<pre>` + string(tokenJSON) + `</pre>
			<button onclick="location.reload()">Recarregar Dashboard</button>
		</div>
	`))
}

// TestRevoke tests token revocation
func (h *Handlers) TestRevoke(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, SessionName)

	// Get session ID
	sessionID, _ := session.Values[KeySessionID].(string)
	if sessionID == "" {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Get token from token store
	tokenStore := services.GetTokenStore()
	token, _, ok := tokenStore.Get(sessionID)
	if !ok || token == nil {
		http.Error(w, "Session expired", http.StatusUnauthorized)
		return
	}

	if token.AccessToken == "" {
		http.Error(w, "No access token available", http.StatusBadRequest)
		return
	}

	// Get OAuth config
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

	oauthService := services.NewOAuthService(oauthConfig, h.historyService)

	// Revoke token
	if err := oauthService.RevokeToken(token.AccessToken); err != nil {
		log.Printf("Token revocation failed: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="error">Token revocation falhou: ` + err.Error() + `</div>`))
		return
	}

	// Clear token store and session
	tokenStore.Delete(sessionID)
	delete(session.Values, KeySessionID)
	session.Save(r, w)

	// Return success
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<div class="success">
			✓ Token revogado com sucesso!
			<p>A sessão foi limpa. <a href="/">Voltar para home</a></p>
		</div>
	`))
}

// TestJWKS tests JWKS fetching and JWT validation
func (h *Handlers) TestJWKS(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, SessionName)

	// Get session ID
	sessionID, _ := session.Values[KeySessionID].(string)

	var idToken string
	if sessionID != "" {
		// Get token from token store
		tokenStore := services.GetTokenStore()
		token, _, ok := tokenStore.Get(sessionID)
		if ok && token != nil {
			if idTokenVal, ok := token.Extra("id_token").(string); ok {
				idToken = idTokenVal
			}
		}
	}

	jwksService := services.NewJWKSService(h.baseURL, h.historyService)

	// Fetch JWKS
	jwks, err := jwksService.FetchJWKS()
	if err != nil {
		log.Printf("JWKS fetch failed: %v", err)
		http.Error(w, "JWKS fetch failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"JWKS": jwks,
	}

	// Validate ID token if available
	if idToken != "" {
		claims, err := jwksService.GetTokenClaims(idToken)
		if err != nil {
			log.Printf("JWT validation failed: %v", err)
			data["ValidationError"] = err.Error()
		} else {
			data["Valid"] = true
			data["Claims"] = claims
		}
	}

	if err := h.templates.ExecuteTemplate(w, "jwks", data); err != nil {
		log.Printf("Error rendering JWKS template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// TestDiscovery tests OIDC discovery endpoint
func (h *Handlers) TestDiscovery(w http.ResponseWriter, r *http.Request) {
	// Get OAuth config
	session, _ := h.sessionStore.Get(r, SessionName)
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

	oauthService := services.NewOAuthService(oauthConfig, h.historyService)

	// Fetch discovery document
	discovery, err := oauthService.FetchDiscovery()
	if err != nil {
		log.Printf("Discovery fetch failed: %v", err)
		http.Error(w, "Discovery fetch failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Pretty print JSON
	discoveryJSON, _ := json.MarshalIndent(discovery, "", "  ")

	data := map[string]interface{}{
		"Discovery":     discovery,
		"DiscoveryJSON": string(discoveryJSON),
	}

	if err := h.templates.ExecuteTemplate(w, "discovery", data); err != nil {
		log.Printf("Error rendering discovery template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}
