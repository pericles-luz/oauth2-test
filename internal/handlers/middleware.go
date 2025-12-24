package handlers

import (
	"context"
	"log"
	"net/http"
	"time"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const isHTMXKey contextKey = "is_htmx"

// HTMXMiddleware detects HTMX requests and adds context
func HTMXMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		isHTMX := r.Header.Get("HX-Request") == "true"
		ctx = context.WithValue(ctx, isHTMXKey, isHTMX)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IsHTMXRequest checks if the current request is from HTMX
func IsHTMXRequest(r *http.Request) bool {
	if val, ok := r.Context().Value(isHTMXKey).(bool); ok {
		return val
	}
	return false
}

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Printf(
			"%s %s %d %v",
			r.Method,
			r.RequestURI,
			wrapped.statusCode,
			time.Since(start),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
