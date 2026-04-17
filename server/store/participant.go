package store

import (
	"context"

	"github.com/pdrhlik/deliberix/server/model"
)

func (s *Store) JoinSurvey(ctx context.Context, p *model.SurveyParticipant) error {
	q := s.DB.Query(`INSERT INTO survey_participant ?values`, p)
	_, err := q.Exec()
	return err
}

func (s *Store) IsParticipant(ctx context.Context, surveyID, userID uint) (bool, error) {
	var count int
	q := s.DB.Query(`SELECT COUNT(*) FROM survey_participant WHERE survey_id = ? AND user_id = ?`, surveyID, userID)
	if err := q.ScanRow(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) ListParticipants(ctx context.Context, surveyID uint) ([]model.ParticipantListItem, error) {
	items := make([]model.ParticipantListItem, 0)
	q := s.DB.Query(`
		SELECT sp.id, sp.user_id, u.name, u.email, sp.role,
			(SELECT COUNT(*) FROM response r
				JOIN statement st ON st.id = r.statement_id
				WHERE st.survey_id = sp.survey_id AND r.user_id = sp.user_id) AS voted,
			(SELECT COUNT(*) FROM statement st
				WHERE st.survey_id = sp.survey_id AND st.status = 'approved') AS total,
			sp.joined_at
		FROM survey_participant sp
		JOIN user u ON u.id = sp.user_id
		WHERE sp.survey_id = ?
		ORDER BY sp.joined_at ASC`, surveyID)
	if err := q.All(&items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) UpdateParticipantRole(ctx context.Context, surveyID, userID uint, role string) error {
	q := s.DB.Query(`UPDATE survey_participant SET role = ? WHERE survey_id = ? AND user_id = ?`, role, surveyID, userID)
	_, err := q.Exec()
	return err
}

func (s *Store) RemoveParticipant(ctx context.Context, surveyID, userID uint) error {
	q := s.DB.Query(`DELETE FROM survey_participant WHERE survey_id = ? AND user_id = ?`, surveyID, userID)
	_, err := q.Exec()
	return err
}
