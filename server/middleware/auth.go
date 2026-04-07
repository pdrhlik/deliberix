package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pdrhlik/edemos/server/identity"
	"github.com/pdrhlik/edemos/server/service"
	"github.com/pdrhlik/edemos/server/store"
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

			ctx := context.WithValue(r.Context(), identity.CtxUserKey, &identity.User{
				ID:   u.ID,
				Role: u.Role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
