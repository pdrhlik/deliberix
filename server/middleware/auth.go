package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pdrhlik/deliberix/server/identity"
	"github.com/pdrhlik/deliberix/server/service"
	"github.com/pdrhlik/deliberix/server/store"
)

func Auth(secret string, s *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			_, token, _ := strings.Cut(auth, "Bearer ")
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := service.ValidateToken(token, secret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			u, err := s.GetUserByID(r.Context(), userID)
			if err != nil || u == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
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
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(`{"error":"email not verified"}`))
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
