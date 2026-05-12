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

func AuthJWT(secret string, s *store.Store) func(http.Handler) http.Handler {
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
			actor := &identity.Actor{UserID: &u.ID, User: ident}
			ctx = identity.ContextWithActor(ctx, actor)

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

// AuthOptional tries Bearer JWT first; if absent or invalid, tries the
// anon cookie. If a valid identity is found it's attached to the request
// context; if neither is present the request continues with no Actor —
// handlers decide what to do.
func AuthOptional(secret string, s *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try JWT first
			auth := r.Header.Get("Authorization")
			_, token, _ := strings.Cut(auth, "Bearer ")
			if token != "" {
				if userID, err := service.ValidateToken(token, secret); err == nil {
					if u, err := s.GetUserByID(r.Context(), userID); err == nil && u != nil {
						ident := &identity.User{
							ID:            u.ID,
							Role:          u.Role,
							EmailVerified: u.EmailVerifiedAt != nil,
						}
						ctx := context.WithValue(r.Context(), identity.CtxUserKey, ident)
						actor := &identity.Actor{UserID: &u.ID, User: ident}
						ctx = identity.ContextWithActor(ctx, actor)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// Try anon cookie
			if anonID := identity.ReadAnonCookie(r, secret); anonID != "" {
				actor := &identity.Actor{AnonSessionID: &anonID}
				ctx := identity.ContextWithActor(r.Context(), actor)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// No identity — pass through with no actor
			next.ServeHTTP(w, r)
		})
	}
}
