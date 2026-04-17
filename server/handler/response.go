package handler

import (
	"net/http"

	"github.com/pdrhlik/deliberix/server/identity"
	"github.com/pdrhlik/deliberix/server/model"
)

func (h *Handler) SubmitResponse() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		statementID, err := parseIDParam(r, "id")
		if err != nil {
			return writeError(w, http.StatusBadRequest, "invalid statement id")
		}

		user := identity.GetUserFromContext(r.Context())

		// Get the survey this statement belongs to
		surveyID, err := h.Store.GetStatementSurveyID(r.Context(), statementID)
		if err != nil {
			return writeError(w, http.StatusNotFound, "statement not found")
		}

		// Check if survey is past its closing time
		survey, err := h.Store.GetSurvey(r.Context(), surveyID)
		if err != nil {
			return err
		}
		if survey == nil {
			return writeError(w, http.StatusNotFound, "survey not found")
		}
		if isSurveyClosed(survey) {
			return writeError(w, http.StatusForbidden, "survey has closed")
		}

		// Verify user is participant
		isParticipant, err := h.Store.IsParticipant(r.Context(), surveyID, user.ID)
		if err != nil {
			return err
		}
		if !isParticipant {
			return writeError(w, http.StatusForbidden, "must be a participant")
		}

		var in model.SubmitResponseRequest
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid request body")
		}

		if in.Vote != "agree" && in.Vote != "disagree" && in.Vote != "abstain" {
			return writeError(w, http.StatusBadRequest, "vote must be agree, disagree, or abstain")
		}

		resp := &model.Response{
			StatementID: statementID,
			UserID:      user.ID,
			Vote:        in.Vote,
			IsImportant: in.IsImportant,
		}

		if err := h.Store.CreateResponse(r.Context(), resp); err != nil {
			return writeError(w, http.StatusConflict, "already voted on this statement")
		}

		return writeJSON(w, http.StatusCreated, resp)
	}
}

func (h *Handler) GetVoteProgress() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		user := identity.GetUserFromContext(r.Context())

		progress, err := h.Store.GetVoteProgress(r.Context(), survey.ID, user.ID)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, progress)
	}
}
