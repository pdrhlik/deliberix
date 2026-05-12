package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pdrhlik/deliberix/server/identity"
	"github.com/pdrhlik/deliberix/server/service"
	"github.com/pdrhlik/deliberix/server/store"
)

func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"code":"` + code + `","message":"` + message + `"}`))
}

func Auth(secret string, s *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			_, token, _ := strings.Cut(auth, "Bearer ")
			if token == "" {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return
			}

			userID, err := service.ValidateToken(token, secret)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return
			}

			u, err := s.GetUserByID(r.Context(), userID)
			if err != nil || u == nil {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return
			}

			ident := &identity.User{
				ID:            u.ID,
				Role:          u.Role,
				EmailVerified: u.EmailVerifiedAt != nil,
			}
			ctx := context.WithValue(r.Context(), identity.CtxUserKey, ident)

			// Block unverified users from all routes except /auth/me and /auth/resend-verification
			if !ident.EmailVerified {
				path := r.URL.Path
				if path != "/api/v1/auth/me" && path != "/api/v1/auth/resend-verification" {
					writeAuthError(w, http.StatusForbidden, "email_not_verified", "email not verified")
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
