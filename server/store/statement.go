package store

import (
	"context"

	"github.com/pdrhlik/deliberix/server/model"
)

func (s *Store) CreateStatement(ctx context.Context, st *model.Statement) (uint, error) {
	q := s.DB.Query(`INSERT INTO statement ?values`, st)
	res, err := q.Exec()
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (s *Store) ListStatementsBySurvey(ctx context.Context, surveyID uint, status string) ([]model.Statement, error) {
	items := make([]model.Statement, 0)
	q := s.DB.Query(`
		SELECT * FROM statement
		WHERE survey_id = ? AND status = ?
		ORDER BY created_at`, surveyID, status)
	if err := q.All(&items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) GetNextStatement(ctx context.Context, surveyID, userID uint, order string) (*model.Statement, error) {
	var orderClause string
	switch order {
	case "sequential":
		orderClause = "s.id ASC"
	case "least_voted":
		orderClause = "(SELECT COUNT(*) FROM response r2 WHERE r2.statement_id = s.id) ASC, s.id ASC"
	default: // random
		orderClause = "RAND()"
	}

	return queryOne[model.Statement](s.DB.Query(`
		SELECT s.* FROM statement s
		WHERE s.survey_id = ?
			AND s.status = 'approved'
			AND s.id NOT IN (
				SELECT r.statement_id FROM response r WHERE r.user_id = ?
			)
		ORDER BY `+orderClause+`
		LIMIT 1`, surveyID, userID))
}

func (s *Store) ModerateStatement(ctx context.Context, statementID, moderatorID uint, status string) error {
	q := s.DB.Query(`
		UPDATE statement
		SET status = ?, moderated_by = ?, moderated_at = NOW()
		WHERE id = ?`, status, moderatorID, statementID)
	_, err := q.Exec()
	return err
}

func (s *Store) GetStatement(ctx context.Context, id uint) (*model.Statement, error) {
	return queryOne[model.Statement](s.DB.Query(`SELECT * FROM statement WHERE id = ?`, id))
}

func (s *Store) CountStatements(ctx context.Context, surveyID uint, status string) (int, error) {
	var count int
	q := s.DB.Query(`SELECT COUNT(*) FROM statement WHERE survey_id = ? AND status = ?`, surveyID, status)
	if err := q.ScanRow(&count); err != nil {
		return 0, err
	}
	return count, nil
}
