package services

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWKSService handles JWT validation using JWKS
type JWKSService struct {
	jwksURL        string
	historyService *HistoryService
}

// NewJWKSService creates a new JWKSService
func NewJWKSService(baseURL string, historyService *HistoryService) *JWKSService {
	return &JWKSService{
		jwksURL:        baseURL + "/oauth2/jwks",
		historyService: historyService,
	}
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

// JWKSet represents a JSON Web Key Set
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// FetchJWKS fetches the JWKS from the server
func (s *JWKSService) FetchJWKS() (*JWKSet, error) {
	// Create HTTP client with logging
	client := NewHTTPClient(s.historyService, "jwks")

	// Create request
	req, err := http.NewRequest("GET", s.jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("JWKS request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("JWKS request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var jwks JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS response: %w", err)
	}

	return &jwks, nil
}

// ValidateToken validates a JWT token using JWKS
func (s *JWKSService) ValidateToken(tokenString string) (*jwt.Token, error) {
	// Parse token to get kid
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get kid from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		// Fetch JWKS
		jwks, err := s.FetchJWKS()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
		}

		// Find matching key
		for _, key := range jwks.Keys {
			if key.Kid == kid {
				// Convert JWK to RSA public key
				publicKey, err := jwkToRSAPublicKey(&key)
				if err != nil {
					return nil, fmt.Errorf("failed to convert JWK to public key: %w", err)
				}
				return publicKey, nil
			}
		}

		return nil, fmt.Errorf("no matching key found for kid: %s", kid)
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	return token, nil
}

// GetTokenClaims validates a token and returns its claims
func (s *JWKSService) GetTokenClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse token claims")
	}

	return claims, nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func jwkToRSAPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	// Decode n (modulus)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode e (exponent)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert bytes to big.Int
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return publicKey, nil
}

// ParseTokenWithoutValidation parses a JWT without validating its signature
// Useful for inspecting token contents
func ParseTokenWithoutValidation(tokenString string) (jwt.MapClaims, error) {
	// Split token into parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode payload (second part)
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	// Parse JSON
	var claims jwt.MapClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return claims, nil
}
