package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/pdrhlik/edemos/server/identity"
	"github.com/pdrhlik/edemos/server/model"
	"github.com/pdrhlik/edemos/server/notify"
	"github.com/pdrhlik/edemos/server/service"
)

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *Handler) Register() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var in model.RegisterRequest
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		in.Email = strings.ToLower(strings.TrimSpace(in.Email))
		in.Name = strings.TrimSpace(in.Name)

		if in.Email == "" || in.Password == "" {
			return writeError(w, http.StatusBadRequest, "email and password are required")
		}
		if len(in.Password) < 8 {
			return writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		}

		existing, err := h.Store.GetUserByEmail(r.Context(), in.Email)
		if err != nil {
			return err
		}
		if existing != nil {
			return writeError(w, http.StatusConflict, "email already registered")
		}

		hash, err := service.HashPassword(in.Password)
		if err != nil {
			return err
		}

		locale := in.Locale
		if locale == "" {
			locale = "en"
		}

		u := &model.User{
			Email:        in.Email,
			PasswordHash: hash,
			Name:         in.Name,
			Locale:       locale,
			Role:         "user",
		}

		id, err := h.Store.CreateUser(r.Context(), u)
		if err != nil {
			return err
		}
		u.ID = id

		// Send verification email if SMTP is configured
		if h.Notify != nil {
			token, err := generateToken()
			if err != nil {
				return err
			}
			expiresAt := time.Now().Add(24 * time.Hour)
			if err := h.Store.CreateEmailVerification(r.Context(), u.ID, token, expiresAt); err != nil {
				return err
			}

			link := fmt.Sprintf("%s/verify-email/%s", h.Config.BaseURL, token)
			n := &notify.EmailVerification{
				UserID:   u.ID,
				Email:    &mail.Address{Name: u.Name, Address: u.Email},
				Language: locale,
				Link:     link,
			}
			if err := h.Notify.EnqueueEmail(n); err != nil {
				return err
			}
		} else {
			// Auto-verify when SMTP is not configured (dev mode)
			now := time.Now()
			u.EmailVerifiedAt = &now
		}

		token, err := service.GenerateToken(u.ID, h.Config.JWTSecret)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusCreated, model.AuthResponse{
			Token: token,
			User:  *u,
		})
	}
}

func (h *Handler) Login() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var in model.LoginRequest
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		in.Email = strings.ToLower(strings.TrimSpace(in.Email))

		u, err := h.Store.GetUserByEmail(r.Context(), in.Email)
		if err != nil {
			return err
		}
		if u == nil {
			return writeError(w, http.StatusUnauthorized, "invalid email or password")
		}

		if err := service.CheckPassword(u.PasswordHash, in.Password); err != nil {
			return writeError(w, http.StatusUnauthorized, "invalid email or password")
		}

		token, err := service.GenerateToken(u.ID, h.Config.JWTSecret)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, model.AuthResponse{
			Token: token,
			User:  *u,
		})
	}
}

func (h *Handler) VerifyEmail() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Token string `json:"token"`
		}
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}
		if in.Token == "" {
			return writeError(w, http.StatusBadRequest, "token is required")
		}

		userID, err := h.Store.UseEmailVerification(r.Context(), in.Token)
		if err != nil {
			return writeError(w, http.StatusBadRequest, "invalid or expired verification token")
		}

		u, err := h.Store.GetUserByID(r.Context(), userID)
		if err != nil {
			return err
		}

		token, err := service.GenerateToken(u.ID, h.Config.JWTSecret)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, model.AuthResponse{
			Token: token,
			User:  *u,
		})
	}
}

func (h *Handler) ForgotPassword() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Email string `json:"email"`
		}
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		in.Email = strings.ToLower(strings.TrimSpace(in.Email))
		if in.Email == "" {
			return writeError(w, http.StatusBadRequest, "email is required")
		}

		// Always return success to prevent email enumeration
		u, err := h.Store.GetUserByEmail(r.Context(), in.Email)
		if err != nil {
			return err
		}
		if u == nil || h.Notify == nil {
			return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		}

		token, err := generateToken()
		if err != nil {
			return err
		}
		expiresAt := time.Now().Add(1 * time.Hour)
		if err := h.Store.CreatePasswordReset(r.Context(), u.ID, token, expiresAt); err != nil {
			return err
		}

		link := fmt.Sprintf("%s/reset-password/%s", h.Config.BaseURL, token)
		n := &notify.PasswordReset{
			UserID:   u.ID,
			Email:    &mail.Address{Name: u.Name, Address: u.Email},
			Language: u.Locale,
			Link:     link,
		}
		if err := h.Notify.EnqueueEmail(n); err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func (h *Handler) ResetPassword() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Token    string `json:"token"`
			Password string `json:"password"`
		}
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}
		if in.Token == "" || in.Password == "" {
			return writeError(w, http.StatusBadRequest, "token and password are required")
		}
		if len(in.Password) < 8 {
			return writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		}

		userID, err := h.Store.UsePasswordReset(r.Context(), in.Token)
		if err != nil {
			return writeError(w, http.StatusBadRequest, "invalid or expired reset token")
		}

		hash, err := service.HashPassword(in.Password)
		if err != nil {
			return err
		}

		if err := h.Store.UpdateUserPassword(r.Context(), userID, hash); err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func (h *Handler) ResendVerification() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ident := identity.GetUserFromContext(r.Context())

		u, err := h.Store.GetUserByID(r.Context(), ident.ID)
		if err != nil {
			return err
		}
		if u == nil {
			return writeError(w, http.StatusNotFound, "user not found")
		}

		if u.EmailVerifiedAt != nil {
			return writeError(w, http.StatusBadRequest, "email already verified")
		}

		if h.Notify == nil {
			return writeError(w, http.StatusServiceUnavailable, "email service not configured")
		}

		recent, err := h.Store.HasRecentVerification(r.Context(), u.ID, 30*time.Second)
		if err != nil {
			return err
		}
		if recent {
			return writeError(w, http.StatusTooManyRequests, "please wait before requesting another verification email")
		}

		token, err := generateToken()
		if err != nil {
			return err
		}
		expiresAt := time.Now().Add(24 * time.Hour)
		if err := h.Store.CreateEmailVerification(r.Context(), u.ID, token, expiresAt); err != nil {
			return err
		}

		link := fmt.Sprintf("%s/verify-email/%s", h.Config.BaseURL, token)
		n := &notify.EmailVerification{
			UserID:   u.ID,
			Email:    &mail.Address{Name: u.Name, Address: u.Email},
			Language: u.Locale,
			Link:     link,
		}
		if err := h.Notify.EnqueueEmail(n); err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func (h *Handler) RequestMagicLink() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Email string `json:"email"`
		}
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		in.Email = strings.ToLower(strings.TrimSpace(in.Email))
		if in.Email == "" {
			return writeError(w, http.StatusBadRequest, "email is required")
		}

		// Always return success to prevent email enumeration
		u, err := h.Store.GetUserByEmail(r.Context(), in.Email)
		if err != nil {
			return err
		}
		if u == nil || h.Notify == nil {
			return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		}

		// Rate limit: 30s between requests
		recent, err := h.Store.HasRecentMagicLink(r.Context(), u.ID, 30*time.Second)
		if err != nil {
			return err
		}
		if recent {
			return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		}

		token, err := generateToken()
		if err != nil {
			return err
		}
		expiresAt := time.Now().Add(15 * time.Minute)
		if err := h.Store.CreateMagicLink(r.Context(), u.ID, token, expiresAt); err != nil {
			return err
		}

		link := fmt.Sprintf("%s/magic-link/%s", h.Config.BaseURL, token)
		n := &notify.MagicLink{
			UserID:   u.ID,
			Email:    &mail.Address{Name: u.Name, Address: u.Email},
			Language: u.Locale,
			Link:     link,
		}
		if err := h.Notify.EnqueueEmail(n); err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func (h *Handler) VerifyMagicLink() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Token string `json:"token"`
		}
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}
		if in.Token == "" {
			return writeError(w, http.StatusBadRequest, "token is required")
		}

		userID, err := h.Store.UseMagicLink(r.Context(), in.Token)
		if err != nil {
			log.Printf("magic link verify failed for token %.8s...: %v", in.Token, err)
			return writeError(w, http.StatusBadRequest, "invalid or expired sign-in link")
		}

		u, err := h.Store.GetUserByID(r.Context(), userID)
		if err != nil {
			return err
		}
		if u == nil {
			return writeError(w, http.StatusBadRequest, "user not found")
		}

		// Auto-verify email if not yet verified (they proved email ownership)
		if u.EmailVerifiedAt == nil {
			if err := h.Store.VerifyUserEmail(r.Context(), u.ID); err != nil {
				return err
			}
			now := time.Now()
			u.EmailVerifiedAt = &now
		}

		token, err := service.GenerateToken(u.ID, h.Config.JWTSecret)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, model.AuthResponse{
			Token: token,
			User:  *u,
		})
	}
}
