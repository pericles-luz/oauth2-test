package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/pericles-luz/oauth2-test/internal/models"
)

// SQLiteDB represents a SQLite database connection
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SQLiteDB{db: db}, nil
}

// RunMigrations executes the SQL migration file
func (s *SQLiteDB) RunMigrations(migrationPath string) error {
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	if _, err := s.db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// SaveHistoryEntry saves an HTTP request/response to the database
func (s *SQLiteDB) SaveHistoryEntry(entry *models.HistoryEntry) error {
	query := `
		INSERT INTO http_history (
			request_method, request_url, request_headers, request_body,
			response_status, response_headers, response_body,
			duration_ms, endpoint_type
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(
		query,
		entry.RequestMethod,
		entry.RequestURL,
		entry.RequestHeaders,
		entry.RequestBody,
		entry.ResponseStatus,
		entry.ResponseHeaders,
		entry.ResponseBody,
		entry.DurationMs,
		entry.EndpointType,
	)
	if err != nil {
		return fmt.Errorf("failed to save history entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	entry.ID = id
	return nil
}

// GetHistoryEntries retrieves history entries with pagination
func (s *SQLiteDB) GetHistoryEntries(limit, offset int) ([]models.HistoryEntry, error) {
	query := `
		SELECT id, request_method, request_url, request_headers, request_body,
		       response_status, response_headers, response_body,
		       duration_ms, endpoint_type, created_at
		FROM http_history
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var entries []models.HistoryEntry
	for rows.Next() {
		var entry models.HistoryEntry
		err := rows.Scan(
			&entry.ID,
			&entry.RequestMethod,
			&entry.RequestURL,
			&entry.RequestHeaders,
			&entry.RequestBody,
			&entry.ResponseStatus,
			&entry.ResponseHeaders,
			&entry.ResponseBody,
			&entry.DurationMs,
			&entry.EndpointType,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetHistoryEntry retrieves a single history entry by ID
func (s *SQLiteDB) GetHistoryEntry(id int64) (*models.HistoryEntry, error) {
	query := `
		SELECT id, request_method, request_url, request_headers, request_body,
		       response_status, response_headers, response_body,
		       duration_ms, endpoint_type, created_at
		FROM http_history
		WHERE id = ?
	`

	var entry models.HistoryEntry
	err := s.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.RequestMethod,
		&entry.RequestURL,
		&entry.RequestHeaders,
		&entry.RequestBody,
		&entry.ResponseStatus,
		&entry.ResponseHeaders,
		&entry.ResponseBody,
		&entry.DurationMs,
		&entry.EndpointType,
		&entry.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get history entry: %w", err)
	}

	return &entry, nil
}

// GetHistoryEntriesByType retrieves history entries filtered by endpoint type
func (s *SQLiteDB) GetHistoryEntriesByType(endpointType string, limit, offset int) ([]models.HistoryEntry, error) {
	query := `
		SELECT id, request_method, request_url, request_headers, request_body,
		       response_status, response_headers, response_body,
		       duration_ms, endpoint_type, created_at
		FROM http_history
		WHERE endpoint_type = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, endpointType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query history by type: %w", err)
	}
	defer rows.Close()

	var entries []models.HistoryEntry
	for rows.Next() {
		var entry models.HistoryEntry
		err := rows.Scan(
			&entry.ID,
			&entry.RequestMethod,
			&entry.RequestURL,
			&entry.RequestHeaders,
			&entry.RequestBody,
			&entry.ResponseStatus,
			&entry.ResponseHeaders,
			&entry.ResponseBody,
			&entry.DurationMs,
			&entry.EndpointType,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// Helper function to serialize headers to JSON
func SerializeHeaders(headers map[string][]string) (string, error) {
	data, err := json.Marshal(headers)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Helper function to deserialize headers from JSON
func DeserializeHeaders(data string) (map[string][]string, error) {
	var headers map[string][]string
	if err := json.Unmarshal([]byte(data), &headers); err != nil {
		return nil, err
	}
	return headers, nil
}
