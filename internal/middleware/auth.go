package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ojooji/booking-service-api/internal/domain"
	"github.com/ojooji/booking-service-api/pkg/jwt"
	"github.com/ojooji/booking-service-api/pkg/response"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				response.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization header required")
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			if tokenStr == header {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "bearer token required")
				return
			}

			claims, err := jwt.Validate(tokenStr, secret)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*jwt.Claims)
	return claims, ok
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r.Context())
		if !ok || claims.Role != string(domain.RoleAdmin) {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
