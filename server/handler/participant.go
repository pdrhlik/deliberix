package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pdrhlik/deliberix/server/identity"
)

func parseUserIDParam(r *http.Request) (uint, error) {
	raw := chi.URLParam(r, "userId")
	id, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (h *Handler) ListParticipants() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		user := identity.GetUserFromContext(r.Context())

		participant, err := h.Store.GetParticipant(r.Context(), survey.ID, user.ID)
		if err != nil {
			return err
		}
		if participant == nil || participant.Role != "admin" {
			return writeError(w, http.StatusForbidden, "only admins can list participants")
		}

		items, err := h.Store.ListParticipants(r.Context(), survey.ID)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, items)
	}
}

func (h *Handler) UpdateParticipantRole() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		user := identity.GetUserFromContext(r.Context())

		participant, err := h.Store.GetParticipant(r.Context(), survey.ID, user.ID)
		if err != nil {
			return err
		}
		if participant == nil || participant.Role != "admin" {
			return writeError(w, http.StatusForbidden, "only admins can change roles")
		}

		targetUserID, err := parseUserIDParam(r)
		if err != nil {
			return writeError(w, http.StatusBadRequest, "invalid user id")
		}

		target, err := h.Store.GetParticipant(r.Context(), survey.ID, targetUserID)
		if err != nil {
			return err
		}
		if target == nil {
			return writeError(w, http.StatusNotFound, "participant not found")
		}
		if target.Role == "admin" {
			return writeError(w, http.StatusForbidden, "cannot change admin role")
		}

		var in struct {
			Role string `json:"role"`
		}
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}
		if in.Role != "participant" && in.Role != "moderator" {
			return writeError(w, http.StatusBadRequest, "role must be participant or moderator")
		}

		if err := h.Store.UpdateParticipantRole(r.Context(), survey.ID, targetUserID, in.Role); err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, map[string]string{"role": in.Role})
	}
}

func (h *Handler) RemoveParticipant() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		user := identity.GetUserFromContext(r.Context())

		participant, err := h.Store.GetParticipant(r.Context(), survey.ID, user.ID)
		if err != nil {
			return err
		}
		if participant == nil || participant.Role != "admin" {
			return writeError(w, http.StatusForbidden, "only admins can remove participants")
		}

		targetUserID, err := parseUserIDParam(r)
		if err != nil {
			return writeError(w, http.StatusBadRequest, "invalid user id")
		}

		target, err := h.Store.GetParticipant(r.Context(), survey.ID, targetUserID)
		if err != nil {
			return err
		}
		if target == nil {
			return writeError(w, http.StatusNotFound, "participant not found")
		}
		if target.Role == "admin" {
			return writeError(w, http.StatusForbidden, "cannot remove admin")
		}

		if err := h.Store.RemoveParticipant(r.Context(), survey.ID, targetUserID); err != nil {
			return err
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
