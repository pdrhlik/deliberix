package model

import "time"

type Statement struct {
	ID          uint       `db:"id,selectonly" json:"id"`
	SurveyID    uint       `db:"survey_id" json:"surveyId"`
	Text        string     `db:"text" json:"text"`
	Type        string     `db:"type" json:"type"`
	Status      string     `db:"status" json:"status"`
	AuthorID    *uint      `db:"author_id" json:"authorId,omitempty"`
	ModeratedBy *uint      `db:"moderated_by" json:"moderatedBy,omitempty"`
	ModeratedAt *time.Time `db:"moderated_at" json:"moderatedAt,omitempty"`
	CreatedAt   time.Time  `db:"created_at,selectonly" json:"createdAt"`
}

type CreateStatementRequest struct {
	Text string `json:"text"`
}

type StatementResult struct {
	ID             uint   `db:"id" json:"id"`
	Text           string `db:"text" json:"text"`
	AgreeCount     int    `db:"agree_count" json:"agreeCount"`
	DisagreeCount  int    `db:"disagree_count" json:"disagreeCount"`
	AbstainCount   int    `db:"abstain_count" json:"abstainCount"`
	ImportantCount int    `db:"important_count" json:"importantCount"`
	TotalVotes     int    `db:"total_votes" json:"totalVotes"`
}
