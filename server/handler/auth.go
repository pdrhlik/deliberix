package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/pdrhlik/edemos/server/model"
	"github.com/pdrhlik/edemos/server/service"
)

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

		now := time.Now()
		u := &model.User{
			Email:           in.Email,
			PasswordHash:    hash,
			Name:            in.Name,
			Locale:          locale,
			Role:            "user",
			EmailVerifiedAt: &now, // auto-verify for MVP
		}

		id, err := h.Store.CreateUser(r.Context(), u)
		if err != nil {
			return err
		}
		u.ID = id

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
