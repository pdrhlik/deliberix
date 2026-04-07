package model

type SurveyStats struct {
	TotalParticipants int     `db:"total_participants" json:"totalParticipants"`
	TotalStatements   int     `db:"total_statements" json:"totalStatements"`
	TotalVotes        int     `db:"total_votes" json:"totalVotes"`
	CompletedCount    int     `db:"completed_count" json:"completedCount"`
	CompletionRate    float64 `json:"completionRate"`
	PendingStatements int     `db:"pending_statements" json:"pendingStatements"`
}
