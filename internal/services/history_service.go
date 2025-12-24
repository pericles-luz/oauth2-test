package services

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pericles-luz/oauth2-test/internal/models"
	"github.com/pericles-luz/oauth2-test/internal/storage"
)

// HistoryService handles HTTP request/response logging
type HistoryService struct {
	db *storage.SQLiteDB
}

// NewHistoryService creates a new HistoryService
func NewHistoryService(db *storage.SQLiteDB) *HistoryService {
	return &HistoryService{db: db}
}

// LogRequest logs an HTTP request and response to the database
func (s *HistoryService) LogRequest(
	method, url string,
	reqHeaders http.Header,
	reqBody []byte,
	respStatus int,
	respHeaders http.Header,
	respBody []byte,
	duration time.Duration,
	endpointType string,
) error {
	reqHeadersJSON, err := storage.SerializeHeaders(reqHeaders)
	if err != nil {
		return fmt.Errorf("failed to serialize request headers: %w", err)
	}

	respHeadersJSON, err := storage.SerializeHeaders(respHeaders)
	if err != nil {
		return fmt.Errorf("failed to serialize response headers: %w", err)
	}

	entry := &models.HistoryEntry{
		RequestMethod:   method,
		RequestURL:      url,
		RequestHeaders:  reqHeadersJSON,
		RequestBody:     string(reqBody),
		ResponseStatus:  respStatus,
		ResponseHeaders: respHeadersJSON,
		ResponseBody:    string(respBody),
		DurationMs:      duration.Milliseconds(),
		EndpointType:    endpointType,
	}

	return s.db.SaveHistoryEntry(entry)
}

// GetHistory retrieves paginated history entries
func (s *HistoryService) GetHistory(limit, offset int) ([]models.HistoryEntry, error) {
	return s.db.GetHistoryEntries(limit, offset)
}

// GetHistoryEntry retrieves a single history entry by ID
func (s *HistoryService) GetHistoryEntry(id int64) (*models.HistoryEntry, error) {
	return s.db.GetHistoryEntry(id)
}

// GetHistoryByType retrieves history entries filtered by endpoint type
func (s *HistoryService) GetHistoryByType(endpointType string, limit, offset int) ([]models.HistoryEntry, error) {
	return s.db.GetHistoryEntriesByType(endpointType, limit, offset)
}

// LoggingTransport is an HTTP transport that logs all requests and responses
type LoggingTransport struct {
	Transport    http.RoundTripper
	History      *HistoryService
	EndpointType string
}

// NewLoggingTransport creates a new LoggingTransport
func NewLoggingTransport(history *HistoryService, endpointType string) *LoggingTransport {
	return &LoggingTransport{
		Transport:    http.DefaultTransport,
		History:      history,
		EndpointType: endpointType,
	}
}

// RoundTrip implements http.RoundTripper interface
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Capture request body
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	// Execute request
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		// Log error case
		_ = t.History.LogRequest(
			req.Method,
			req.URL.String(),
			req.Header,
			reqBody,
			0,
			http.Header{},
			[]byte(err.Error()),
			time.Since(start),
			t.EndpointType,
		)
		return nil, err
	}

	// Capture response body
	var respBody []byte
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	}

	duration := time.Since(start)

	// Log to history service
	_ = t.History.LogRequest(
		req.Method,
		req.URL.String(),
		req.Header,
		reqBody,
		resp.StatusCode,
		resp.Header,
		respBody,
		duration,
		t.EndpointType,
	)

	return resp, nil
}

// NewHTTPClient creates an HTTP client with logging transport
func NewHTTPClient(history *HistoryService, endpointType string) *http.Client {
	return &http.Client{
		Transport: NewLoggingTransport(history, endpointType),
		Timeout:   30 * time.Second,
	}
}
