package services

import (
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// TokenData holds token and user info with expiration
type TokenData struct {
	Token      *oauth2.Token
	UserInfo   interface{}
	ExpiresAt  time.Time
}

// TokenStore provides thread-safe in-memory token storage
type TokenStore struct {
	tokens map[string]*TokenData
	mu     sync.RWMutex
}

var (
	globalTokenStore *TokenStore
	tokenStoreOnce   sync.Once
)

// GetTokenStore returns the singleton token store instance
func GetTokenStore() *TokenStore {
	tokenStoreOnce.Do(func() {
		globalTokenStore = &TokenStore{
			tokens: make(map[string]*TokenData),
		}
		// Start cleanup goroutine
		go globalTokenStore.cleanup()
	})
	return globalTokenStore
}

// Store saves token and user info for a session ID
func (ts *TokenStore) Store(sessionID string, token *oauth2.Token, userInfo interface{}) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.tokens[sessionID] = &TokenData{
		Token:     token,
		UserInfo:  userInfo,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Expire after 24 hours
	}
}

// Get retrieves token and user info for a session ID
func (ts *TokenStore) Get(sessionID string) (*oauth2.Token, interface{}, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	data, exists := ts.tokens[sessionID]
	if !exists || time.Now().After(data.ExpiresAt) {
		return nil, nil, false
	}

	return data.Token, data.UserInfo, true
}

// Delete removes token data for a session ID
func (ts *TokenStore) Delete(sessionID string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	delete(ts.tokens, sessionID)
}

// cleanup periodically removes expired tokens
func (ts *TokenStore) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ts.mu.Lock()
		now := time.Now()
		for sessionID, data := range ts.tokens {
			if now.After(data.ExpiresAt) {
				delete(ts.tokens, sessionID)
			}
		}
		ts.mu.Unlock()
	}
}
