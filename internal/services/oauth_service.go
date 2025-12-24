package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"

	"github.com/pericles-luz/oauth2-test/internal/models"
)

// OAuthService handles OAuth2 operations
type OAuthService struct {
	config         *oauth2.Config
	historyService *HistoryService
}

// NewOAuthService creates a new OAuthService
func NewOAuthService(oauthConfig *models.OAuthConfig, historyService *HistoryService) *OAuthService {
	config := &oauth2.Config{
		ClientID:     oauthConfig.ClientID,
		ClientSecret: oauthConfig.ClientSecret,
		RedirectURL:  oauthConfig.RedirectURI,
		Scopes:       oauthConfig.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthConfig.BaseURL + "/oauth2/authorize",
			TokenURL: oauthConfig.BaseURL + "/oauth2/token",
		},
	}

	return &OAuthService{
		config:         config,
		historyService: historyService,
	}
}

// GenerateAuthURL generates the authorization URL with PKCE
func (s *OAuthService) GenerateAuthURL(state string) (authURL, verifier string, err error) {
	// Generate PKCE verifier
	verifier = oauth2.GenerateVerifier()

	// Build authorization URL with PKCE S256 challenge
	authURL = s.config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(verifier),
	)

	return authURL, verifier, nil
}

// ExchangeCode exchanges the authorization code for tokens
func (s *OAuthService) ExchangeCode(code, verifier string) (*oauth2.Token, error) {
	ctx := context.Background()

	// Create HTTP client with logging
	client := NewHTTPClient(s.historyService, "token")
	ctx = context.WithValue(ctx, oauth2.HTTPClient, client)

	// Exchange code for token with PKCE verifier
	token, err := s.config.Exchange(
		ctx,
		code,
		oauth2.VerifierOption(verifier),
	)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	return token, nil
}

// GetUserInfo fetches user information from the /oauth2/userinfo endpoint
func (s *OAuthService) GetUserInfo(accessToken string) (*models.UserInfo, error) {
	// Create HTTP client with logging
	client := NewHTTPClient(s.historyService, "userinfo")

	// Build userinfo URL
	userinfoURL := s.config.Endpoint.AuthURL
	userinfoURL = strings.Replace(userinfoURL, "/oauth2/authorize", "/oauth2/userinfo", 1)

	// Create request
	req, err := http.NewRequest("GET", userinfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var userInfo models.UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return &userInfo, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *OAuthService) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	ctx := context.Background()

	// Create HTTP client with logging
	client := NewHTTPClient(s.historyService, "refresh")
	ctx = context.WithValue(ctx, oauth2.HTTPClient, client)

	// Create token source
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}
	tokenSource := s.config.TokenSource(ctx, token)

	// Get new token
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	return newToken, nil
}

// RevokeToken revokes an access token
func (s *OAuthService) RevokeToken(token string) error {
	// Create HTTP client with logging
	client := NewHTTPClient(s.historyService, "revoke")

	// Build revoke URL
	revokeURL := s.config.Endpoint.AuthURL
	revokeURL = strings.Replace(revokeURL, "/oauth2/authorize", "/oauth2/revoke", 1)

	// Prepare form data
	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", s.config.ClientID)
	data.Set("client_secret", s.config.ClientSecret)

	// Create request
	req, err := http.NewRequest("POST", revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer resp.Body.Close()

	// RFC 7009 specifies that revocation endpoint should return 200 even if token is invalid
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// FetchDiscovery fetches the OIDC discovery document
func (s *OAuthService) FetchDiscovery() (map[string]interface{}, error) {
	// Create HTTP client with logging
	client := NewHTTPClient(s.historyService, "discovery")

	// Build discovery URL
	discoveryURL := s.config.Endpoint.AuthURL
	discoveryURL = strings.Replace(discoveryURL, "/oauth2/authorize", "/.well-known/openid-configuration", 1)

	// Create request
	req, err := http.NewRequest("GET", discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discovery request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discovery request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var discovery map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("failed to decode discovery response: %w", err)
	}

	return discovery, nil
}

// GenerateRandomState generates a cryptographically secure random state
func GenerateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
