package models

import "time"

// HistoryEntry represents a logged HTTP request/response
type HistoryEntry struct {
	ID              int64     `json:"id"`
	RequestMethod   string    `json:"request_method"`
	RequestURL      string    `json:"request_url"`
	RequestHeaders  string    `json:"request_headers"`   // JSON serialized
	RequestBody     string    `json:"request_body"`
	ResponseStatus  int       `json:"response_status"`
	ResponseHeaders string    `json:"response_headers"`  // JSON serialized
	ResponseBody    string    `json:"response_body"`
	DurationMs      int64     `json:"duration_ms"`
	EndpointType    string    `json:"endpoint_type"`
	CreatedAt       time.Time `json:"created_at"`
}
