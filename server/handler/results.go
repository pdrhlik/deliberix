package handler

import (
	"net/http"

	"github.com/pdrhlik/edemos/server/identity"
	"github.com/pdrhlik/edemos/server/model"
)

type ResultsResponse struct {
	Stats      model.SurveyStats       `json:"stats"`
	Statements []model.StatementResult `json:"statements"`
}

func (h *Handler) GetResults() AppHandlerFunc {
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

		// Enforce result_visibility rules
		switch survey.ResultVisibility {
		case "after_close":
			if survey.Status != "closed" {
				return writeError(w, http.StatusForbidden, "results available after survey closes")
			}
		case "after_completion":
			progress, err := h.Store.GetVoteProgress(r.Context(), surveyID, user.ID)
			if err != nil {
				return err
			}
			if progress.Total > 0 && progress.Voted < progress.Total {
				return writeError(w, http.StatusForbidden, "complete all votes to see results")
			}
		case "continuous":
			// Always visible — no restrictions
		}

		stats, err := h.Store.GetSurveyStats(r.Context(), surveyID)
		if err != nil {
			return err
		}

		statements, err := h.Store.GetSurveyResults(r.Context(), surveyID)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, ResultsResponse{
			Stats:      stats,
			Statements: statements,
		})
	}
}

func (h *Handler) GetSurveyStats() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		surveyID, err := parseIDParam(r, "id")
		if err != nil {
			return writeError(w, http.StatusBadRequest, "invalid survey id")
		}

		stats, err := h.Store.GetSurveyStats(r.Context(), surveyID)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, stats)
	}
}
