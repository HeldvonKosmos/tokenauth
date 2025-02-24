package tokenauth

import (
	"context"
	"net/http"
	"net/url"
)

const (
	TokenKey  = "ta_token"
	CookieKey = "ta_session_token"
)

// Config holds the middleware configuration
type Config struct {
	TokenParam    string   `json:"tokenParam,omitempty"`
	CookieName    string   `json:"cookieName,omitempty"`
	AllowedTokens []string `json:"allowedTokens,omitempty"`
}

// CreateConfig creates and initializes the config with default values
func CreateConfig() *Config {
	return &Config{
		TokenParam: TokenKey,
		CookieName: CookieKey,
	}
}

type tokenAuth struct {
	next          http.Handler
	name          string
	tokenParam    string
	cookieName    string
	allowedTokens []string
}

// New creates and returns a new token authentication middleware
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.TokenParam) == 0 {
		config.TokenParam = TokenKey
	}
	if len(config.CookieName) == 0 {
		config.CookieName = CookieKey
	}

	return &tokenAuth{
		next:          next,
		name:          name,
		tokenParam:    config.TokenParam,
		cookieName:    config.CookieName,
		allowedTokens: config.AllowedTokens,
	}, nil
}

func (t *tokenAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Check for existing session cookie
	cookie, err := req.Cookie(t.cookieName)
	if err == nil && t.isTokenValid(cookie.Value) {
		// Valid session exists, proceed
		t.next.ServeHTTP(rw, req)
		return
	}

	// Check for token in query parameters
	token := req.URL.Query().Get(t.tokenParam)
	if token == "" {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate token
	if !t.isTokenValid(token) {
		http.Error(rw, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Set session cookie
	cookie = &http.Cookie{
		Name:     t.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(rw, cookie)

	// Remove token from URL
	q := req.URL.Query()
	q.Del(t.tokenParam)
	req.URL.RawQuery = q.Encode()

	// Create new URL without token parameter
	newURL := &url.URL{
		Scheme:   req.URL.Scheme,
		Host:     req.URL.Host,
		Path:     req.URL.Path,
		RawQuery: q.Encode(),
	}

	// Redirect to clean URL
	http.Redirect(rw, req, newURL.String(), http.StatusTemporaryRedirect)
}

// isTokenValid checks if the token is in the list of allowed tokens
func (t *tokenAuth) isTokenValid(token string) bool {
	if len(t.allowedTokens) == 0 {
		return false
	}

	for _, allowedToken := range t.allowedTokens {
		if token == allowedToken {
			return true
		}
	}
	return false
}
