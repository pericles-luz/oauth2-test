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

	refreshToken, _ := session.Values[KeyRefreshToken].(string)
	if refreshToken == "" {
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
	newToken, err := oauthService.RefreshToken(refreshToken)
	if err != nil {
		log.Printf("Token refresh failed: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="error">Token refresh falhou: ` + err.Error() + `</div>`))
		return
	}

	// Update session with new tokens
	session.Values[KeyAccessToken] = newToken.AccessToken
	if newToken.RefreshToken != "" {
		session.Values[KeyRefreshToken] = newToken.RefreshToken
	}
	session.Save(r, w)

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

	accessToken, _ := session.Values[KeyAccessToken].(string)
	if accessToken == "" {
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
	if err := oauthService.RevokeToken(accessToken); err != nil {
		log.Printf("Token revocation failed: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="error">Token revocation falhou: ` + err.Error() + `</div>`))
		return
	}

	// Clear session
	delete(session.Values, KeyAccessToken)
	delete(session.Values, KeyRefreshToken)
	delete(session.Values, KeyIDToken)
	delete(session.Values, "user_info")
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

	idToken, _ := session.Values[KeyIDToken].(string)

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
