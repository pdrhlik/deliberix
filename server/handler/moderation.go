package handler

import (
	"net/http"

	"github.com/pdrhlik/deliberix/server/identity"
)

func (h *Handler) GetModerationQueue() AppHandlerFunc {
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
		if participant == nil || (participant.Role != "admin" && participant.Role != "moderator") {
			return writeError(w, http.StatusForbidden, "only admins and moderators can access moderation")
		}

		items, err := h.Store.ListStatementsBySurvey(r.Context(), survey.ID, "pending")
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, items)
	}
}

func (h *Handler) ModerateStatement() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		statementID, err := parseIDParam(r, "id")
		if err != nil {
			return writeError(w, http.StatusBadRequest, "invalid statement id")
		}

		user := identity.GetUserFromContext(r.Context())

		// Get statement to find survey
		st, err := h.Store.GetStatement(r.Context(), statementID)
		if err != nil {
			return err
		}
		if st == nil {
			return writeError(w, http.StatusNotFound, "statement not found")
		}

		// Verify user is admin or moderator of this survey
		participant, err := h.Store.GetParticipant(r.Context(), st.SurveyID, user.ID)
		if err != nil {
			return err
		}
		if participant == nil || (participant.Role != "admin" && participant.Role != "moderator") {
			return writeError(w, http.StatusForbidden, "only admins and moderators can moderate")
		}

		var in struct {
			Status string `json:"status"`
		}
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		if in.Status != "approved" && in.Status != "rejected" {
			return writeError(w, http.StatusBadRequest, "status must be approved or rejected")
		}

		if err := h.Store.ModerateStatement(r.Context(), statementID, user.ID, in.Status); err != nil {
			return err
		}

		updated, err := h.Store.GetStatement(r.Context(), statementID)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, updated)
	}
}
