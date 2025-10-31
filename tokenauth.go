package tokenauth

import (
	"context"
	"crypto/hmac" // For secure comparison
	"net/http"
)

// Config holds the plugin configuration.
type Config struct {
	// AllowedTokens holds the list of secret API keys (without the "Bearer " prefix).
	AllowedTokens []string `json:"allowedTokens,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// tokenAuth holds the middleware state.
type tokenAuth struct {
	next           http.Handler
	name           string
	// allowedHeaders stores the full "Bearer <token>" strings for comparison.
	allowedHeaders []string
}

// New creates a new tokenAuth middleware.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.AllowedTokens) == 0 {
		// Note: No tokens configured, middleware will block all requests.
	}

	// Build the list of full, expected header values
	headers := make([]string, len(config.AllowedTokens))
	for i, token := range config.AllowedTokens {
		headers[i] = "Bearer " + token
	}

	return &tokenAuth{
		next:           next,
		name:           name,
		allowedHeaders: headers,
	}, nil
}

// ServeHTTP checks the Authorization header and blocks or forwards the request.
func (t *tokenAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Get the Authorization header from the request.
	authHeader := req.Header.Get("Authorization")

	// Check if the received header is in our list of allowed headers.
	if t.isTokenValid(authHeader) {
		// --- Success ---
		// Token is valid, forward the request.
		t.next.ServeHTTP(rw, req)
		return
	}

	// --- Failure ---
	// Token is missing or invalid.
	rw.Header().Set("WWW-Authenticate", "Bearer realm=\"Restricted\"")
	http.Error(rw, "Unauthorized", http.StatusUnauthorized)
}

// isTokenValid securely checks the received header against the allowed list.
func (t *tokenAuth) isTokenValid(receivedHeader string) bool {
	if len(t.allowedHeaders) == 0 || receivedHeader == "" {
		return false // No tokens configured or no header received
	}

	receivedHeaderBytes := []byte(receivedHeader)

	// Iterate over all allowed headers
	for _, allowedHeader := range t.allowedHeaders {
		// Use hmac.Equal for constant-time comparison
		// to prevent timing attacks.
		if hmac.Equal(receivedHeaderBytes, []byte(allowedHeader)) {
			return true
		}
	}

	// No matching token found
	return false
}
