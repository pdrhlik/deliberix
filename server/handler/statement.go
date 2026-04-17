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
			return writeError(w, http.StatusForbidden, "only survey admins can add seed statements")
		}

		var in model.CreateStatementRequest
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		if in.Text == "" {
			return writeError(w, http.StatusBadRequest, "text is required")
		}

		textLen := uint(len([]rune(in.Text)))
		if textLen < survey.StatementCharMin || textLen > survey.StatementCharMax {
			return writeError(w, http.StatusBadRequest, "statement text length out of range")
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

		user := identity.GetUserFromContext(r.Context())

		isParticipant, err := h.Store.IsParticipant(r.Context(), survey.ID, user.ID)
		if err != nil {
			return err
		}
		if !isParticipant {
			return writeError(w, http.StatusForbidden, "must be a participant to submit statements")
		}

		if survey.Status != "active" {
			return writeError(w, http.StatusBadRequest, "survey is not active")
		}

		if isSurveyClosed(survey) {
			return writeError(w, http.StatusForbidden, "survey has closed")
		}

		var in model.CreateStatementRequest
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		if in.Text == "" {
			return writeError(w, http.StatusBadRequest, "text is required")
		}

		textLen := uint(len([]rune(in.Text)))
		if textLen < survey.StatementCharMin || textLen > survey.StatementCharMax {
			return writeError(w, http.StatusBadRequest, "statement text length out of range")
		}

		st := &model.Statement{
			SurveyID: survey.ID,
			Text:     in.Text,
			Type:     "user_submitted",
			Status:   "pending",
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
			return writeError(w, http.StatusForbidden, "survey has closed")
		}

		user := identity.GetUserFromContext(r.Context())

		st, err := h.Store.GetNextStatement(r.Context(), survey.ID, user.ID, survey.StatementOrder)
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
