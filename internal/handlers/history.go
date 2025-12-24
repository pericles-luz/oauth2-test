package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// HistoryList displays all HTTP request/response history
func (h *Handlers) HistoryList(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	// Get history entries
	entries, err := h.historyService.GetHistory(limit, offset)
	if err != nil {
		log.Printf("Error fetching history: %v", err)
		http.Error(w, "Error fetching history", http.StatusInternalServerError)
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"Entries": entries,
		"Limit":   limit,
		"Offset":  offset,
	}

	if err := h.templates.ExecuteTemplate(w, "history", data); err != nil {
		log.Printf("Error rendering history template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// HistoryDetail displays details of a single history entry
func (h *Handlers) HistoryDetail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get history entry
	entry, err := h.historyService.GetHistoryEntry(id)
	if err != nil {
		log.Printf("Error fetching history entry: %v", err)
		http.Error(w, "Error fetching history entry", http.StatusInternalServerError)
		return
	}

	if entry == nil {
		http.Error(w, "History entry not found", http.StatusNotFound)
		return
	}

	// Parse headers
	var reqHeaders, respHeaders map[string][]string
	json.Unmarshal([]byte(entry.RequestHeaders), &reqHeaders)
	json.Unmarshal([]byte(entry.ResponseHeaders), &respHeaders)

	// Prepare template data
	data := map[string]interface{}{
		"Entry":           entry,
		"RequestHeaders":  reqHeaders,
		"ResponseHeaders": respHeaders,
	}

	// Pretty print JSON bodies if applicable
	if isJSON(entry.RequestBody) {
		var prettyReqBody interface{}
		json.Unmarshal([]byte(entry.RequestBody), &prettyReqBody)
		prettyReqJSON, _ := json.MarshalIndent(prettyReqBody, "", "  ")
		data["PrettyRequestBody"] = string(prettyReqJSON)
	}

	if isJSON(entry.ResponseBody) {
		var prettyRespBody interface{}
		json.Unmarshal([]byte(entry.ResponseBody), &prettyRespBody)
		prettyRespJSON, _ := json.MarshalIndent(prettyRespBody, "", "  ")
		data["PrettyResponseBody"] = string(prettyRespJSON)
	}

	if err := h.templates.ExecuteTemplate(w, "history_detail", data); err != nil {
		log.Printf("Error rendering history detail template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// isJSON checks if a string is valid JSON
func isJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
