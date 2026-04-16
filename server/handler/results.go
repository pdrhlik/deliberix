package handler

import (
	"net/http"

	"github.com/pdrhlik/edemos/server/identity"
	"github.com/pdrhlik/edemos/server/model"
)

type ResultsResponse struct {
	Stats      model.SurveyStats            `json:"stats"`
	Statements []model.StatementResult      `json:"statements"`
	MyVotes    map[uint]model.UserVote      `json:"myVotes"`
}

func (h *Handler) GetResults() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		user := identity.GetUserFromContext(r.Context())

		switch survey.ResultVisibility {
		case "after_close":
			if survey.Status != "closed" {
				return writeError(w, http.StatusForbidden, "Results are available after the survey closes.")
			}
		case "after_completion":
			progress, err := h.Store.GetVoteProgress(r.Context(), survey.ID, user.ID)
			if err != nil {
				return err
			}
			if progress.Total > 0 && progress.Voted < progress.Total {
				return writeError(w, http.StatusForbidden, "Please complete all votes to see results.")
			}
		case "continuous":
			// Always visible
		}

		stats, err := h.Store.GetSurveyStats(r.Context(), survey.ID)
		if err != nil {
			return err
		}

		statements, err := h.Store.GetSurveyResults(r.Context(), survey.ID)
		if err != nil {
			return err
		}

		myVotes, err := h.Store.GetUserVotesForSurvey(r.Context(), survey.ID, user.ID)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, ResultsResponse{
			Stats:      stats,
			Statements: statements,
			MyVotes:    myVotes,
		})
	}
}

func (h *Handler) GetSurveyStats() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		stats, err := h.Store.GetSurveyStats(r.Context(), survey.ID)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, stats)
	}
}
