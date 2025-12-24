package models

// OAuthConfig holds the OAuth2 client configuration
type OAuthConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURI  string   `json:"redirect_uri"`
	Scopes       []string `json:"scopes"`
	BaseURL      string   `json:"base_url"`
}

// Validate checks if the configuration is valid
func (c *OAuthConfig) Validate() error {
	if c.ClientID == "" {
		return &ValidationError{Field: "client_id", Message: "Client ID is required"}
	}
	if c.ClientSecret == "" {
		return &ValidationError{Field: "client_secret", Message: "Client Secret is required"}
	}
	if c.RedirectURI == "" {
		return &ValidationError{Field: "redirect_uri", Message: "Redirect URI is required"}
	}
	if len(c.Scopes) == 0 {
		return &ValidationError{Field: "scopes", Message: "At least one scope is required"}
	}

	// Check if openid scope is present
	hasOpenID := false
	for _, scope := range c.Scopes {
		if scope == "openid" {
			hasOpenID = true
			break
		}
	}
	if !hasOpenID {
		return &ValidationError{Field: "scopes", Message: "openid scope is required"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
