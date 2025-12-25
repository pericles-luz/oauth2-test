package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pericles-luz/oauth2-test/internal/models"
	"github.com/pericles-luz/oauth2-test/internal/services"
)

// Dashboard displays user information and tokens after successful authentication
func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, SessionName)

	// Check if user is authenticated
	sessionID, _ := session.Values[KeySessionID].(string)
	if sessionID == "" {
		http.Error(w, "Not authenticated. Please login first.", http.StatusUnauthorized)
		return
	}

	// Get tokens and user info from token store
	tokenStore := services.GetTokenStore()
	token, userInfoData, ok := tokenStore.Get(sessionID)
	if !ok || token == nil {
		http.Error(w, "Session expired. Please login again.", http.StatusUnauthorized)
		return
	}

	// Convert user info from interface{} to UserInfo
	var userInfo *models.UserInfo
	if userInfoData != nil {
		jsonData, _ := json.Marshal(userInfoData)
		userInfo = &models.UserInfo{}
		json.Unmarshal(jsonData, userInfo)
	}

	// Extract ID token if present
	var idToken string
	if idTokenVal, ok := token.Extra("id_token").(string); ok {
		idToken = idTokenVal
	}

	// Get scopes from session
	scopesStr, _ := session.Values[KeyScopes].(string)

	// Prepare template data
	data := map[string]interface{}{
		"AccessToken":  token.AccessToken,
		"RefreshToken": token.RefreshToken,
		"IDToken":      idToken,
		"UserInfo":     userInfo,
		"Scopes":       scopesStr,
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard", data); err != nil {
		log.Printf("Error rendering dashboard template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}
