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
			return writeError(w, http.StatusBadRequest, "invalid_statement_id", "invalid statement id")
		}

		actor := identity.GetActorFromContext(r.Context())
		if actor == nil {
			return writeError(w, http.StatusForbidden, "must_be_participant", "must be a participant")
		}

		// Get the survey this statement belongs to
		surveyID, err := h.Store.GetStatementSurveyID(r.Context(), statementID)
		if err != nil {
			return writeError(w, http.StatusNotFound, "statement_not_found", "statement not found")
		}

		// Check if survey is past its closing time
		survey, err := h.Store.GetSurvey(r.Context(), surveyID)
		if err != nil {
			return err
		}
		if survey == nil {
			return writeError(w, http.StatusNotFound, "survey_not_found", "survey not found")
		}
		if isSurveyClosed(survey) {
			return writeError(w, http.StatusForbidden, "survey_closed", "survey has closed")
		}

		// Verify actor is participant
		isParticipant, err := h.Store.IsParticipantByActor(r.Context(), surveyID, actor)
		if err != nil {
			return err
		}
		if !isParticipant {
			return writeError(w, http.StatusForbidden, "must_be_participant", "must be a participant")
		}

		var in model.SubmitResponseRequest
		if err := parseJSON(r, &in); err != nil {
			return writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		}

		if in.Vote != "agree" && in.Vote != "disagree" && in.Vote != "abstain" {
			return writeError(w, http.StatusBadRequest, "invalid_vote", "vote must be agree, disagree, or abstain")
		}

		if err := h.Store.CreateResponseByActor(r.Context(), statementID, actor, in.Vote, in.IsImportant); err != nil {
			return writeError(w, http.StatusConflict, "already_voted", "already voted on this statement")
		}

		return writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
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

		actor := identity.GetActorFromContext(r.Context())
		if actor == nil {
			return writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		}

		progress, err := h.Store.GetVoteProgressByActor(r.Context(), survey.ID, actor)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, progress)
	}
}
