package handlers

import (
	"log"
	"net/http"
	"strings"
)

// Home renders the home/configuration page
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, SessionName)

	// Get existing config from session
	data := map[string]interface{}{
		"ClientID":     session.Values[KeyClientID],
		"ClientSecret": session.Values[KeyClientSecret],
		"RedirectURI":  session.Values[KeyRedirectURI],
		"Scopes":       session.Values[KeyScopes],
		"BaseURL":      h.baseURL,
	}

	// Available scopes
	data["AvailableScopes"] = []struct {
		Value    string
		Label    string
		Required bool
	}{
		{"openid", "openid (obrigatório)", true},
		{"profile", "profile - Perfil básico (nome, CPF)", false},
		{"email", "email - Endereço de email", false},
		{"phone", "phone - Número de telefone", false},
		{"address", "address - Endereço", false},
		{"membership", "membership - Status de filiação", false},
		{"permissions", "permissions - Permissões do usuário", false},
		{"union_unit", "union_unit - Detalhes da seccional", false},
	}

	if err := h.templates.ExecuteTemplate(w, "home", data); err != nil {
		log.Printf("Error rendering home template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// SaveConfig saves OAuth2 configuration to session
func (h *Handlers) SaveConfig(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	redirectURI := r.FormValue("redirect_uri")
	scopesStr := r.FormValue("scopes")

	// Validate required fields
	if clientID == "" || clientSecret == "" || redirectURI == "" {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="error">Todos os campos são obrigatórios</div>`))
		return
	}

	// Parse scopes
	scopes := []string{}
	if scopesStr != "" {
		scopes = strings.Split(scopesStr, ",")
	}

	// Ensure openid scope is present
	hasOpenID := false
	for _, scope := range scopes {
		if strings.TrimSpace(scope) == "openid" {
			hasOpenID = true
			break
		}
	}
	if !hasOpenID {
		scopes = append([]string{"openid"}, scopes...)
	}

	// Save to session
	session, _ := h.sessionStore.Get(r, SessionName)
	session.Values[KeyClientID] = clientID
	session.Values[KeyClientSecret] = clientSecret
	session.Values[KeyRedirectURI] = redirectURI
	session.Values[KeyScopes] = strings.Join(scopes, " ")

	if err := session.Save(r, w); err != nil {
		log.Printf("Error saving session: %v", err)
		http.Error(w, "Error saving configuration", http.StatusInternalServerError)
		return
	}

	// Return success message
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<div class="success">
			✓ Configuração salva com sucesso!
			<a href="/auth/login" class="btn btn-primary" style="margin-left: 10px;">
				Iniciar Fluxo OAuth2
			</a>
		</div>
	`))
}
