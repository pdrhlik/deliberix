package handler

import (
	"net/http"

	"github.com/pdrhlik/deliberix/server/identity"
	"github.com/pdrhlik/deliberix/server/model"
)

func (h *Handler) ListStatements() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		items, err := h.Store.ListStatementsBySurvey(r.Context(), survey.ID, "approved")
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, items)
	}
}

func (h *Handler) AddSeedStatement() AppHandlerFunc {
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
			return writeError(w, http.StatusForbidden, "only_admin_can_seed", "only survey admins can add seed statements")
		}

		var in model.CreateStatementRequest
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		}

		if in.Text == "" {
			return writeError(w, http.StatusBadRequest, "text_required", "text is required")
		}

		textLen := uint(len([]rune(in.Text)))
		if textLen < survey.StatementCharMin || textLen > survey.StatementCharMax {
			return writeError(w, http.StatusBadRequest, "statement_text_length", "statement text length out of range")
		}

		st := &model.Statement{
			SurveyID: survey.ID,
			Text:     in.Text,
			Type:     "seed",
			Status:   "approved",
			AuthorID: &user.ID,
		}

		id, err := h.Store.CreateStatement(r.Context(), st)
		if err != nil {
			return err
		}
		st.ID = id

		return writeJSON(w, http.StatusCreated, st)
	}
}

func (h *Handler) SubmitStatement() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		actor := identity.GetActorFromContext(r.Context())
		if actor == nil {
			return writeError(w, http.StatusForbidden, "must_be_participant", "must be a participant to submit statements")
		}

		isParticipant, err := h.Store.IsParticipantByActor(r.Context(), survey.ID, actor)
		if err != nil {
			return err
		}
		if !isParticipant {
			return writeError(w, http.StatusForbidden, "must_be_participant", "must be a participant to submit statements")
		}

		if survey.Status != "active" {
			return writeError(w, http.StatusBadRequest, "survey_not_active", "survey is not active")
		}

		if isSurveyClosed(survey) {
			return writeError(w, http.StatusForbidden, "survey_closed", "survey has closed")
		}

		var in model.CreateStatementRequest
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		}

		if in.Text == "" {
			return writeError(w, http.StatusBadRequest, "text_required", "text is required")
		}

		textLen := uint(len([]rune(in.Text)))
		if textLen < survey.StatementCharMin || textLen > survey.StatementCharMax {
			return writeError(w, http.StatusBadRequest, "statement_text_length", "statement text length out of range")
		}

		status := "pending"
		if !survey.ModerationEnabled {
			status = "approved"
		}

		st := &model.Statement{
			SurveyID: survey.ID,
			Text:     in.Text,
			Type:     "user_submitted",
			Status:   status,
			AuthorID: actor.UserID,
		}

		id, err := h.Store.CreateStatement(r.Context(), st)
		if err != nil {
			return err
		}
		st.ID = id

		return writeJSON(w, http.StatusCreated, st)
	}
}

func (h *Handler) GetNextStatement() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		if isSurveyClosed(survey) {
			return writeError(w, http.StatusForbidden, "survey_closed", "survey has closed")
		}

		actor := identity.GetActorFromContext(r.Context())
		if actor == nil {
			return writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		}

		st, err := h.Store.GetNextStatementByActor(r.Context(), survey.ID, actor, survey.StatementOrder)
		if err != nil {
			return err
		}
		if st == nil {
			w.WriteHeader(http.StatusNoContent)
			return nil
		}

		return writeJSON(w, http.StatusOK, st)
	}
}
