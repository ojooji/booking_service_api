package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// NewCORS builds the CORS middleware from the configured origins.
// Credentials are only allowed when origins are explicitly listed:
// combining a wildcard with credentials would let any site make
// credentialed cross-origin requests.
func NewCORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := len(allowedOrigins) == 0
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
			break
		}
	}
	if allowAll {
		allowedOrigins = []string{"*"}
	}

	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: !allowAll,
		MaxAge:           300,
	})
}
