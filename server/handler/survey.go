package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mibk/dali"
	"github.com/pdrhlik/deliberix/server/identity"
	"github.com/pdrhlik/deliberix/server/model"
	"github.com/pdrhlik/deliberix/server/slug"
)

var (
	validVisibilities      = map[string]bool{"public": true, "private": true, "unlisted": true}
	validPrivacyModes      = map[string]bool{"anonymous": true, "public": true, "participant_choice": true}
	validInvitationModes   = map[string]bool{"none": true, "admin_only": true, "participants_can_invite": true}
	validResultVisibility  = map[string]bool{"after_completion": true, "continuous": true, "after_close": true}
	validStatementOrders   = map[string]bool{"random": true, "sequential": true, "least_voted": true}
	validFieldTypes        = map[string]bool{"text": true, "select": true, "radio": true, "checkbox": true}
)

func validateIntakeConfig(raw json.RawMessage) error {
	var config struct {
		Fields []struct {
			ID        string `json:"id"`
			Type      string `json:"type"`
			Options   []struct {
				Value string `json:"value"`
			} `json:"options"`
			Condition *struct {
				FieldID string `json:"fieldId"`
			} `json:"condition"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return fmt.Errorf("invalid intake config JSON")
	}

	ids := map[string]bool{}
	for _, f := range config.Fields {
		if f.ID == "" {
			return fmt.Errorf("each intake field must have an ID")
		}
		if ids[f.ID] {
			return fmt.Errorf("duplicate intake field ID: %s", f.ID)
		}
		ids[f.ID] = true

		if !validFieldTypes[f.Type] {
			return fmt.Errorf("invalid field type: %s", f.Type)
		}

		if f.Type == "select" || f.Type == "radio" || f.Type == "checkbox" {
			if len(f.Options) == 0 {
				return fmt.Errorf("field %s must have at least one option", f.ID)
			}
			for _, opt := range f.Options {
				if opt.Value == "" {
					return fmt.Errorf("field %s has an option with empty value", f.ID)
				}
			}
		}

		if f.Condition != nil && f.Condition.FieldID != "" {
			if !ids[f.Condition.FieldID] {
				return fmt.Errorf("field %s references unknown condition field: %s", f.ID, f.Condition.FieldID)
			}
		}
	}
	return nil
}

// isSurveyClosed returns true if the survey has a closesAt time that has passed.
func isSurveyClosed(survey *model.Survey) bool {
	return survey.ClosesAt != nil && survey.ClosesAt.Before(time.Now())
}

// getSurveyFromSlug resolves the survey from the {slug} URL param.
func (h *Handler) getSurveyFromSlug(w http.ResponseWriter, r *http.Request) (*model.Survey, error) {
	s := chi.URLParam(r, "slug")
	survey, err := h.Store.GetSurveyBySlug(r.Context(), s)
	if err != nil {
		return nil, err
	}
	if survey == nil {
		writeError(w, http.StatusNotFound, "survey not found")
		return nil, nil
	}
	return survey, nil
}

func (h *Handler) generateUniqueSlug(ctx context.Context, title string) (string, error) {
	base := slug.Generate(title)
	candidate := base

	exists, err := h.Store.SlugExists(ctx, candidate)
	if err != nil {
		return "", err
	}
	for exists {
		candidate = slug.WithSuffix(base)
		exists, err = h.Store.SlugExists(ctx, candidate)
		if err != nil {
			return "", err
		}
	}
	return candidate, nil
}

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

		surveySlug, err := h.generateUniqueSlug(r.Context(), in.Title)
		if err != nil {
			return err
		}

		charMin := uint(20)
		charMax := uint(150)
		if in.StatementCharMin != nil {
			charMin = *in.StatementCharMin
		}
		if in.StatementCharMax != nil {
			charMax = *in.StatementCharMax
		}

		if charMin > charMax {
			return writeError(w, http.StatusBadRequest, "statement_char_min must be less than or equal to statement_char_max")
		}

		visibility := "private"
		if in.Visibility != "" {
			if !validVisibilities[in.Visibility] {
				return writeError(w, http.StatusBadRequest, "invalid visibility value")
			}
			visibility = in.Visibility
		}
		privacyMode := "anonymous"
		if in.PrivacyMode != "" {
			if !validPrivacyModes[in.PrivacyMode] {
				return writeError(w, http.StatusBadRequest, "invalid privacy_mode value")
			}
			privacyMode = in.PrivacyMode
		}
		invitationMode := "none"
		if in.InvitationMode != "" {
			if !validInvitationModes[in.InvitationMode] {
				return writeError(w, http.StatusBadRequest, "invalid invitation_mode value")
			}
			invitationMode = in.InvitationMode
		}
		resultVisibility := "after_completion"
		if in.ResultVisibility != "" {
			if !validResultVisibility[in.ResultVisibility] {
				return writeError(w, http.StatusBadRequest, "invalid result_visibility value")
			}
			resultVisibility = in.ResultVisibility
		}
		statementOrder := "random"
		if in.StatementOrder != "" {
			if !validStatementOrders[in.StatementOrder] {
				return writeError(w, http.StatusBadRequest, "invalid statement_order value")
			}
			statementOrder = in.StatementOrder
		}

		if in.IntakeConfig != nil {
			if err := validateIntakeConfig(*in.IntakeConfig); err != nil {
				return writeError(w, http.StatusBadRequest, err.Error())
			}
		}

		survey := &model.Survey{
			Title:            in.Title,
			Slug:             surveySlug,
			Description:      in.Description,
			Status:           "draft",
			Visibility:       visibility,
			PrivacyMode:      privacyMode,
			InvitationMode:   invitationMode,
			ResultVisibility: resultVisibility,
			StatementOrder:   statementOrder,
			StatementCharMin: charMin,
			StatementCharMax: charMax,
			IntakeConfig: in.IntakeConfig,
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
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil // 404 already written
		}

		// Private surveys are only visible to participants
		if survey.Visibility == "private" {
			user := identity.GetUserFromContext(r.Context())
			isParticipant, err := h.Store.IsParticipant(r.Context(), survey.ID, user.ID)
			if err != nil {
				return err
			}
			if !isParticipant {
				return writeError(w, http.StatusNotFound, "survey not found")
			}
		}

		return writeJSON(w, http.StatusOK, survey)
	}
}

func (h *Handler) UpdateSurvey() AppHandlerFunc {
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
			if !validVisibilities[*in.Visibility] {
				return writeError(w, http.StatusBadRequest, "invalid visibility value")
			}
			fields["visibility"] = *in.Visibility
		}
		if in.PrivacyMode != nil {
			if !validPrivacyModes[*in.PrivacyMode] {
				return writeError(w, http.StatusBadRequest, "invalid privacy_mode value")
			}
			fields["privacy_mode"] = *in.PrivacyMode
		}
		if in.InvitationMode != nil {
			if !validInvitationModes[*in.InvitationMode] {
				return writeError(w, http.StatusBadRequest, "invalid invitation_mode value")
			}
			fields["invitation_mode"] = *in.InvitationMode
		}
		if in.ResultVisibility != nil {
			if !validResultVisibility[*in.ResultVisibility] {
				return writeError(w, http.StatusBadRequest, "invalid result_visibility value")
			}
			fields["result_visibility"] = *in.ResultVisibility
		}
		if in.StatementOrder != nil {
			if !validStatementOrders[*in.StatementOrder] {
				return writeError(w, http.StatusBadRequest, "invalid statement_order value")
			}
			fields["statement_order"] = *in.StatementOrder
		}
		if in.StatementCharMin != nil {
			fields["statement_char_min"] = *in.StatementCharMin
		}
		if in.StatementCharMax != nil {
			fields["statement_char_max"] = *in.StatementCharMax
		}
		// Validate charMin <= charMax (consider both new values and existing survey values)
		charMin := survey.StatementCharMin
		if in.StatementCharMin != nil {
			charMin = *in.StatementCharMin
		}
		charMax := survey.StatementCharMax
		if in.StatementCharMax != nil {
			charMax = *in.StatementCharMax
		}
		if charMin > charMax {
			return writeError(w, http.StatusBadRequest, "statement_char_min must be less than or equal to statement_char_max")
		}
		if in.IntakeConfig != nil {
			if err := validateIntakeConfig(*in.IntakeConfig); err != nil {
				return writeError(w, http.StatusBadRequest, err.Error())
			}
			fields["intake_config"] = *in.IntakeConfig
		}
		if in.ClosesAt != nil {
			fields["closes_at"] = *in.ClosesAt
		}

		if len(fields) == 0 {
			return writeError(w, http.StatusBadRequest, "no fields to update")
		}

		if err := h.Store.UpdateSurvey(r.Context(), survey.ID, fields); err != nil {
			return err
		}

		updated, err := h.Store.GetSurvey(r.Context(), survey.ID)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, updated)
	}
}

func (h *Handler) GetMyParticipation() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		user := identity.GetUserFromContext(r.Context())
		p, err := h.Store.GetParticipant(r.Context(), survey.ID, user.ID)
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
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		user := identity.GetUserFromContext(r.Context())

		if survey.Status != "active" {
			return writeError(w, http.StatusBadRequest, "survey is not active")
		}

		if isSurveyClosed(survey) {
			return writeError(w, http.StatusForbidden, "survey has closed")
		}

		isParticipant, err := h.Store.IsParticipant(r.Context(), survey.ID, user.ID)
		if err != nil {
			return err
		}
		if isParticipant {
			return writeError(w, http.StatusConflict, "already a participant")
		}

		var body struct {
			IntakeData *json.RawMessage `json:"intakeData,omitempty"`
		}
		if err := parseJSON(r, &body); err != nil {
			body.IntakeData = nil
		}

		p := &model.SurveyParticipant{
			SurveyID:   survey.ID,
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
