// Package middleware provides chi middleware for JWT authentication and
// role-based access control, replacing Laravel's auth:karyawan / auth:user
// guard middleware and adding RBAC the legacy admin panel never had (D-7).
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/eprisi/absensi-next/services/api/internal/platform/authtoken"
)

type ctxKey string

const claimsCtxKey ctxKey = "authClaims"

func RequireAuth(issuer authtoken.Issuer, aud authtoken.Audience) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimPrefix(header, "Bearer ")

			claims, err := issuer.Parse(tokenString)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}
			if claims.Audience != aud {
				http.Error(w, "token not valid for this endpoint", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAnyAuth accepts a token from any of the given audiences -- used
// only by the handful of endpoints genuinely shared by both sides (self
// profile, notifications), where the handler itself branches on
// claims.Audience rather than the route dictating a single fixed audience.
func RequireAnyAuth(issuer authtoken.Issuer, auds ...authtoken.Audience) func(http.Handler) http.Handler {
	allowed := make(map[authtoken.Audience]struct{}, len(auds))
	for _, a := range auds {
		allowed[a] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimPrefix(header, "Bearer ")

			claims, err := issuer.Parse(tokenString)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}
			if _, ok := allowed[claims.Audience]; !ok {
				http.Error(w, "token not valid for this endpoint", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole enforces RBAC on top of RequireAuth -- e.g. superadmin-only
// config endpoints (docs/openapi.yaml security: [bearerAuth: [superadmin]]).
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}
			if _, permitted := allowed[claims.Role]; !permitted {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*authtoken.Claims, bool) {
	claims, ok := ctx.Value(claimsCtxKey).(*authtoken.Claims)
	return claims, ok
}
