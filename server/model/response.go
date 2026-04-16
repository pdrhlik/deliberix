package model

import "time"

type Response struct {
	ID          uint      `db:"id,selectonly" json:"id"`
	StatementID uint      `db:"statement_id" json:"statementId"`
	UserID      uint      `db:"user_id" json:"userId"`
	Vote        string    `db:"vote" json:"vote"`
	IsImportant bool      `db:"is_important" json:"isImportant"`
	CreatedAt   time.Time `db:"created_at,selectonly" json:"createdAt"`
}

type SubmitResponseRequest struct {
	Vote        string `json:"vote"`
	IsImportant bool   `json:"isImportant"`
}

type VoteProgress struct {
	Voted int `json:"voted"`
	Total int `json:"total"`
}

type UserVote struct {
	Vote        string `json:"vote"`
	IsImportant bool   `json:"isImportant"`
}
