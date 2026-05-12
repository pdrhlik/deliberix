package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pdrhlik/deliberix/server/identity"
)

func (h *Handler) secureCookies() bool {
	return strings.HasPrefix(h.Config.BaseURL, "https://")
}

func (h *Handler) AnonJoin() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		if !survey.AllowAnonymous {
			return writeError(w, http.StatusForbidden, "anon_not_allowed", "this survey does not accept anonymous voting")
		}
		if survey.Status != "active" {
			return writeError(w, http.StatusBadRequest, "survey_not_active", "survey is not active")
		}
		if isSurveyClosed(survey) {
			return writeError(w, http.StatusForbidden, "survey_closed", "survey has closed")
		}

		// If the caller already has a participant row (via JWT or anon cookie), treat as conflict.
		if actor := identity.GetActorFromContext(r.Context()); actor != nil {
			p, err := h.Store.GetParticipantByActor(r.Context(), survey.ID, actor)
			if err != nil {
				return err
			}
			if p != nil {
				return writeError(w, http.StatusConflict, "already_a_participant", "already a participant")
			}
		}

		var body struct {
			IntakeData *json.RawMessage `json:"intakeData,omitempty"`
		}
		_ = parseJSON(r, &body) // body is optional

		sessionID := identity.NewAnonSessionID()
		if err := h.Store.CreateAnonParticipant(r.Context(), survey.ID, sessionID, body.IntakeData); err != nil {
			// A concurrent request may have created a participant for this caller
			// between our check and our insert (e.g., double-tapped button).
			// Re-check via the request's actor — if a row exists now, treat as 409.
			if actor := identity.GetActorFromContext(r.Context()); actor != nil {
				if p, _ := h.Store.GetParticipantByActor(r.Context(), survey.ID, actor); p != nil {
					return writeError(w, http.StatusConflict, "already_a_participant", "already a participant")
				}
			}
			return err
		}

		identity.SetAnonCookie(w, identity.SignAnonSessionID(sessionID, h.Config.JWTSecret), h.secureCookies())

		return writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
	}
}

func (h *Handler) AnonLogout() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		identity.ClearAnonCookie(w, h.secureCookies())
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
