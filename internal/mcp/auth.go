package mcp

import (
	"net/http"
	"strings"
)

// AuthMiddleware provides optional API key authentication.
// If no API key is configured on the server, all requests are allowed (dev-friendly).
// When configured, requests must include one of:
//   - Authorization: Bearer <key>
//   - X-API-Key: <key>
func AuthMiddleware(apiKey string, next http.Handler) http.Handler {
	if apiKey == "" {
		return next // auth disabled
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provided := extractAPIKey(r)
		if provided == "" || provided != apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func extractAPIKey(r *http.Request) string {
	// Authorization: Bearer <key>
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	// X-API-Key
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	return ""
}
