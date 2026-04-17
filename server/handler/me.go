package handler

import (
	"net/http"
	"strings"

	"github.com/pdrhlik/deliberix/server/identity"
	"github.com/pdrhlik/deliberix/server/service"
)

func (h *Handler) Me() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ident := identity.GetUserFromContext(r.Context())
		if ident == nil {
			return writeError(w, http.StatusUnauthorized, "unauthorized")
		}

		u, err := h.Store.GetUserByID(r.Context(), ident.ID)
		if err != nil {
			return err
		}
		if u == nil {
			return writeError(w, http.StatusUnauthorized, "user not found")
		}

		return writeJSON(w, http.StatusOK, u)
	}
}

func (h *Handler) UpdateProfile() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ident := identity.GetUserFromContext(r.Context())

		var in struct {
			Name   *string `json:"name"`
			Locale *string `json:"locale"`
		}
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		u, err := h.Store.GetUserByID(r.Context(), ident.ID)
		if err != nil {
			return err
		}
		if u == nil {
			return writeError(w, http.StatusNotFound, "user not found")
		}

		name := u.Name
		if in.Name != nil {
			name = strings.TrimSpace(*in.Name)
		}
		locale := u.Locale
		if in.Locale != nil {
			l := *in.Locale
			if l != "en" && l != "cs" {
				return writeError(w, http.StatusBadRequest, "locale must be en or cs")
			}
			locale = l
		}

		if err := h.Store.UpdateUserProfile(r.Context(), u.ID, name, locale); err != nil {
			return err
		}

		updated, err := h.Store.GetUserByID(r.Context(), u.ID)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, updated)
	}
}

func (h *Handler) ChangePassword() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ident := identity.GetUserFromContext(r.Context())

		var in struct {
			CurrentPassword string `json:"currentPassword"`
			NewPassword     string `json:"newPassword"`
		}
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		if in.NewPassword == "" || len(in.NewPassword) < 8 {
			return writeError(w, http.StatusBadRequest, "New password must be at least 8 characters.")
		}

		u, err := h.Store.GetUserByID(r.Context(), ident.ID)
		if err != nil {
			return err
		}
		if u == nil {
			return writeError(w, http.StatusNotFound, "user not found")
		}

		// If user already has a password, require current password
		if u.PasswordHash != "" {
			if in.CurrentPassword == "" {
				return writeError(w, http.StatusBadRequest, "Please enter your current password.")
			}
			if err := service.CheckPassword(u.PasswordHash, in.CurrentPassword); err != nil {
				return writeError(w, http.StatusUnauthorized, "The current password you entered is incorrect.")
			}
		}

		hash, err := service.HashPassword(in.NewPassword)
		if err != nil {
			return err
		}

		if err := h.Store.UpdateUserPassword(r.Context(), u.ID, hash); err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
