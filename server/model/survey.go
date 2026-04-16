package model

import (
	"encoding/json"
	"time"
)

type Survey struct {
	ID               uint             `db:"id,selectonly" json:"id"`
	OrganizationID   *uint            `db:"organization_id" json:"organizationId,omitempty"`
	Title            string           `db:"title" json:"title"`
	Slug             string           `db:"slug" json:"slug"`
	Description      *string          `db:"description" json:"description,omitempty"`
	Status           string           `db:"status" json:"status"`
	Visibility       string           `db:"visibility" json:"visibility"`
	PrivacyMode      string           `db:"privacy_mode" json:"privacyMode"`
	InvitationMode   string           `db:"invitation_mode" json:"invitationMode"`
	ResultVisibility string           `db:"result_visibility" json:"resultVisibility"`
	StatementOrder   string           `db:"statement_order" json:"statementOrder"`
	StatementCharMin uint             `db:"statement_char_min" json:"statementCharMin"`
	StatementCharMax uint             `db:"statement_char_max" json:"statementCharMax"`
	IntakeConfig     *json.RawMessage `db:"intake_config" json:"intakeConfig,omitempty"`
	ClosesAt         *time.Time       `db:"closes_at" json:"closesAt,omitempty"`
	CreatedBy        uint             `db:"created_by" json:"createdBy"`
	CreatedAt        time.Time        `db:"created_at,selectonly" json:"createdAt"`
	UpdatedAt        time.Time        `db:"updated_at,selectonly" json:"updatedAt"`
}

type CreateSurveyRequest struct {
	Title            string           `json:"title"`
	Description      *string          `json:"description,omitempty"`
	Visibility       string           `json:"visibility"`
	PrivacyMode      string           `json:"privacyMode"`
	InvitationMode   string           `json:"invitationMode"`
	ResultVisibility string           `json:"resultVisibility"`
	StatementOrder   string           `json:"statementOrder"`
	StatementCharMin *uint            `json:"statementCharMin,omitempty"`
	StatementCharMax *uint            `json:"statementCharMax,omitempty"`
	IntakeConfig     *json.RawMessage `json:"intakeConfig,omitempty"`
	ClosesAt         *time.Time       `json:"closesAt,omitempty"`
}

type UpdateSurveyRequest struct {
	Title            *string          `json:"title,omitempty"`
	Description      *string          `json:"description,omitempty"`
	Status           *string          `json:"status,omitempty"`
	Visibility       *string          `json:"visibility,omitempty"`
	PrivacyMode      *string          `json:"privacyMode,omitempty"`
	InvitationMode   *string          `json:"invitationMode,omitempty"`
	ResultVisibility *string          `json:"resultVisibility,omitempty"`
	StatementOrder   *string          `json:"statementOrder,omitempty"`
	StatementCharMin *uint            `json:"statementCharMin,omitempty"`
	StatementCharMax *uint            `json:"statementCharMax,omitempty"`
	IntakeConfig     *json.RawMessage `json:"intakeConfig,omitempty"`
	ClosesAt         *time.Time       `json:"closesAt,omitempty"`
}

type SurveyListItem struct {
	ID               uint      `db:"id" json:"id"`
	Title            string    `db:"title" json:"title"`
	Slug             string    `db:"slug" json:"slug"`
	Description      *string   `db:"description" json:"description,omitempty"`
	Status           string    `db:"status" json:"status"`
	Role             string    `db:"role" json:"role"`
	Voted            int       `db:"voted" json:"voted"`
	Total            int       `db:"total" json:"total"`
	ParticipantCount int       `db:"participant_count" json:"participantCount"`
	StatementCount   int       `db:"statement_count" json:"statementCount"`
	CreatedAt        time.Time `db:"created_at" json:"createdAt"`
}
