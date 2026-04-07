package store

import (
	"context"

	"github.com/mibk/dali"
	"github.com/pdrhlik/edemos/server/model"
)

func (s *Store) CreateSurvey(ctx context.Context, survey *model.Survey) (uint, error) {
	q := s.DB.Query(`INSERT INTO survey ?values`, survey)
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

func (s *Store) GetSurvey(ctx context.Context, id uint) (*model.Survey, error) {
	return queryOne[model.Survey](s.DB.Query(`SELECT * FROM survey WHERE id = ?`, id))
}

func (s *Store) ListSurveysByUser(ctx context.Context, userID uint) ([]model.SurveyListItem, error) {
	items := make([]model.SurveyListItem, 0)
	q := s.DB.Query(`
		SELECT s.id, s.title, s.status, sp.role,
			(SELECT COUNT(*) FROM response r
				JOIN statement st ON st.id = r.statement_id
				WHERE st.survey_id = s.id AND r.user_id = sp.user_id) AS voted,
			(SELECT COUNT(*) FROM statement st
				WHERE st.survey_id = s.id AND st.status = 'approved') AS total,
			s.created_at
		FROM survey s
		JOIN survey_participant sp ON sp.survey_id = s.id
		WHERE sp.user_id = ?
		ORDER BY s.created_at DESC`, userID)
	if err := q.All(&items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) ListPublicSurveys(ctx context.Context, userID uint) ([]model.SurveyListItem, error) {
	items := make([]model.SurveyListItem, 0)
	q := s.DB.Query(`
		SELECT s.id, s.title, s.status, '' AS role, 0 AS voted, 0 AS total, s.created_at
		FROM survey s
		WHERE s.status = 'active'
			AND s.visibility IN ('public', 'unlisted')
			AND s.id NOT IN (
				SELECT sp.survey_id FROM survey_participant sp WHERE sp.user_id = ?
			)
		ORDER BY s.created_at DESC`, userID)
	if err := q.All(&items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) UpdateSurvey(ctx context.Context, id uint, fields dali.Map) error {
	q := s.DB.Query(`UPDATE survey ?set WHERE id = ?`, fields, id)
	_, err := q.Exec()
	return err
}

func (s *Store) AddParticipant(ctx context.Context, p *model.SurveyParticipant) error {
	q := s.DB.Query(`INSERT INTO survey_participant ?values`, p)
	_, err := q.Exec()
	return err
}

func (s *Store) GetParticipant(ctx context.Context, surveyID, userID uint) (*model.SurveyParticipant, error) {
	return queryOne[model.SurveyParticipant](s.DB.Query(
		`SELECT * FROM survey_participant WHERE survey_id = ? AND user_id = ?`, surveyID, userID))
}
