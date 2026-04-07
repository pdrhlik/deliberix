package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mibk/dali"
	"github.com/pdrhlik/edemos/server/identity"
	"github.com/pdrhlik/edemos/server/model"
)

func (h *Handler) CreateSurvey() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var in model.CreateSurveyRequest
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		if in.Title == "" {
			return writeError(w, http.StatusBadRequest, "title is required")
		}

		user := identity.GetUserFromContext(r.Context())

		charMin := uint(20)
		charMax := uint(150)
		if in.StatementCharMin != nil {
			charMin = *in.StatementCharMin
		}
		if in.StatementCharMax != nil {
			charMax = *in.StatementCharMax
		}

		visibility := "private"
		if in.Visibility != "" {
			visibility = in.Visibility
		}
		privacyMode := "anonymous"
		if in.PrivacyMode != "" {
			privacyMode = in.PrivacyMode
		}
		invitationMode := "none"
		if in.InvitationMode != "" {
			invitationMode = in.InvitationMode
		}
		resultVisibility := "after_completion"
		if in.ResultVisibility != "" {
			resultVisibility = in.ResultVisibility
		}
		statementOrder := "random"
		if in.StatementOrder != "" {
			statementOrder = in.StatementOrder
		}

		survey := &model.Survey{
			Title:            in.Title,
			Description:      in.Description,
			Status:           "draft",
			Visibility:       visibility,
			PrivacyMode:      privacyMode,
			InvitationMode:   invitationMode,
			ResultVisibility: resultVisibility,
			StatementOrder:   statementOrder,
			StatementCharMin: charMin,
			StatementCharMax: charMax,
			IntakeConfig:     in.IntakeConfig,
			ClosesAt:         in.ClosesAt,
			CreatedBy:        user.ID,
		}

		id, err := h.Store.CreateSurvey(r.Context(), survey)
		if err != nil {
			return err
		}
		survey.ID = id

		// Auto-add creator as survey admin
		err = h.Store.AddParticipant(r.Context(), &model.SurveyParticipant{
			SurveyID: id,
			UserID:   user.ID,
			Role:     "admin",
		})
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusCreated, survey)
	}
}

func (h *Handler) ListSurveys() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user := identity.GetUserFromContext(r.Context())
		items, err := h.Store.ListSurveysByUser(r.Context(), user.ID)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, items)
	}
}

func (h *Handler) ListPublicSurveys() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user := identity.GetUserFromContext(r.Context())
		items, err := h.Store.ListPublicSurveys(r.Context(), user.ID)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, items)
	}
}

func (h *Handler) GetSurvey() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := parseIDParam(r, "id")
		if err != nil {
			return writeError(w, http.StatusBadRequest, "invalid survey id")
		}

		survey, err := h.Store.GetSurvey(r.Context(), id)
		if err != nil {
			return err
		}
		if survey == nil {
			return writeError(w, http.StatusNotFound, "survey not found")
		}

		return writeJSON(w, http.StatusOK, survey)
	}
}

func (h *Handler) UpdateSurvey() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := parseIDParam(r, "id")
		if err != nil {
			return writeError(w, http.StatusBadRequest, "invalid survey id")
		}

		user := identity.GetUserFromContext(r.Context())

		// Verify user is survey admin
		participant, err := h.Store.GetParticipant(r.Context(), id, user.ID)
		if err != nil {
			return err
		}
		if participant == nil || participant.Role != "admin" {
			return writeError(w, http.StatusForbidden, "only survey admins can update")
		}

		var in model.UpdateSurveyRequest
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		fields := dali.Map{}
		if in.Title != nil {
			fields["title"] = *in.Title
		}
		if in.Description != nil {
			fields["description"] = *in.Description
		}
		if in.Status != nil {
			fields["status"] = *in.Status
		}
		if in.Visibility != nil {
			fields["visibility"] = *in.Visibility
		}
		if in.PrivacyMode != nil {
			fields["privacy_mode"] = *in.PrivacyMode
		}
		if in.InvitationMode != nil {
			fields["invitation_mode"] = *in.InvitationMode
		}
		if in.ResultVisibility != nil {
			fields["result_visibility"] = *in.ResultVisibility
		}
		if in.StatementOrder != nil {
			fields["statement_order"] = *in.StatementOrder
		}
		if in.StatementCharMin != nil {
			fields["statement_char_min"] = *in.StatementCharMin
		}
		if in.StatementCharMax != nil {
			fields["statement_char_max"] = *in.StatementCharMax
		}
		if in.IntakeConfig != nil {
			fields["intake_config"] = *in.IntakeConfig
		}
		if in.ClosesAt != nil {
			fields["closes_at"] = *in.ClosesAt
		}

		if len(fields) == 0 {
			return writeError(w, http.StatusBadRequest, "no fields to update")
		}

		if err := h.Store.UpdateSurvey(r.Context(), id, fields); err != nil {
			return err
		}

		survey, err := h.Store.GetSurvey(r.Context(), id)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, survey)
	}
}

func (h *Handler) GetMyParticipation() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		surveyID, err := parseIDParam(r, "id")
		if err != nil {
			return writeError(w, http.StatusBadRequest, "invalid survey id")
		}

		user := identity.GetUserFromContext(r.Context())
		p, err := h.Store.GetParticipant(r.Context(), surveyID, user.ID)
		if err != nil {
			return err
		}
		if p == nil {
			return writeError(w, http.StatusNotFound, "not a participant")
		}

		return writeJSON(w, http.StatusOK, p)
	}
}

func (h *Handler) JoinSurvey() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		surveyID, err := parseIDParam(r, "id")
		if err != nil {
			return writeError(w, http.StatusBadRequest, "invalid survey id")
		}

		user := identity.GetUserFromContext(r.Context())

		survey, err := h.Store.GetSurvey(r.Context(), surveyID)
		if err != nil {
			return err
		}
		if survey == nil {
			return writeError(w, http.StatusNotFound, "survey not found")
		}
		if survey.Status != "active" {
			return writeError(w, http.StatusBadRequest, "survey is not active")
		}

		// Check if already a participant
		isParticipant, err := h.Store.IsParticipant(r.Context(), surveyID, user.ID)
		if err != nil {
			return err
		}
		if isParticipant {
			return writeError(w, http.StatusConflict, "already a participant")
		}

		// Parse optional intake data
		var body struct {
			IntakeData *json.RawMessage `json:"intakeData,omitempty"`
		}
		if err := parseJSON(r, &body); err != nil {
			// Allow empty body for surveys without intake config
			body.IntakeData = nil
		}

		p := &model.SurveyParticipant{
			SurveyID:   surveyID,
			UserID:     user.ID,
			Role:       "participant",
			IntakeData: body.IntakeData,
		}

		if err := h.Store.JoinSurvey(r.Context(), p); err != nil {
			return err
		}

		return writeJSON(w, http.StatusCreated, p)
	}
}

func parseIDParam(r *http.Request, name string) (uint, error) {
	raw := chi.URLParam(r, name)
	id, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
