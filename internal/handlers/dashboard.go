package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pericles-luz/oauth2-test/internal/models"
)

// Dashboard displays user information and tokens after successful authentication
func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, SessionName)

	// Check if user is authenticated
	accessToken, _ := session.Values[KeyAccessToken].(string)
	if accessToken == "" {
		http.Error(w, "Not authenticated. Please login first.", http.StatusUnauthorized)
		return
	}

	// Get user info from session
	var userInfo *models.UserInfo
	if userInfoData, ok := session.Values["user_info"]; ok {
		// User info is stored as interface{}, need to convert
		// This is a workaround for session serialization
		jsonData, _ := json.Marshal(userInfoData)
		userInfo = &models.UserInfo{}
		json.Unmarshal(jsonData, userInfo)
	}

	// Prepare template data
	data := map[string]interface{}{
		"AccessToken":  accessToken,
		"RefreshToken": session.Values[KeyRefreshToken],
		"IDToken":      session.Values[KeyIDToken],
		"UserInfo":     userInfo,
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard", data); err != nil {
		log.Printf("Error rendering dashboard template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}
