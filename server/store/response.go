package store

import (
	"context"

	"github.com/pdrhlik/deliberix/server/model"
)

func (s *Store) CreateResponse(ctx context.Context, resp *model.Response) error {
	q := s.DB.Query(`INSERT INTO response ?values`, resp)
	_, err := q.Exec()
	return err
}

func (s *Store) GetVoteProgress(ctx context.Context, surveyID, userID uint) (model.VoteProgress, error) {
	var p model.VoteProgress
	q := s.DB.Query(`
		SELECT
			(SELECT COUNT(*) FROM response r
				JOIN statement s ON s.id = r.statement_id
				WHERE s.survey_id = ? AND r.user_id = ?) AS voted,
			(SELECT COUNT(*) FROM statement
				WHERE survey_id = ? AND status = 'approved') AS total`,
		surveyID, userID, surveyID)
	if err := q.ScanRow(&p.Voted, &p.Total); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Store) GetSurveyResults(ctx context.Context, surveyID uint) ([]model.StatementResult, error) {
	items := make([]model.StatementResult, 0)
	q := s.DB.Query(`
		SELECT
			s.id,
			s.text,
			COALESCE(SUM(r.vote = 'agree'), 0) AS agree_count,
			COALESCE(SUM(r.vote = 'disagree'), 0) AS disagree_count,
			COALESCE(SUM(r.vote = 'abstain'), 0) AS abstain_count,
			COALESCE(SUM(r.is_important), 0) AS important_count,
			COUNT(r.id) AS total_votes
		FROM statement s
		LEFT JOIN response r ON r.statement_id = s.id
		WHERE s.survey_id = ? AND s.status = 'approved'
		GROUP BY s.id, s.text
		ORDER BY total_votes DESC`, surveyID)
	if err := q.All(&items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) GetSurveyStats(ctx context.Context, surveyID uint) (model.SurveyStats, error) {
	var stats model.SurveyStats

	q := s.DB.Query(`SELECT COUNT(*) FROM survey_participant WHERE survey_id = ?`, surveyID)
	if err := q.ScanRow(&stats.TotalParticipants); err != nil {
		return stats, err
	}

	q = s.DB.Query(`SELECT COUNT(*) FROM statement WHERE survey_id = ? AND status = 'approved'`, surveyID)
	if err := q.ScanRow(&stats.TotalStatements); err != nil {
		return stats, err
	}

	q = s.DB.Query(`
		SELECT COUNT(*) FROM response r
		JOIN statement s ON s.id = r.statement_id
		WHERE s.survey_id = ?`, surveyID)
	if err := q.ScanRow(&stats.TotalVotes); err != nil {
		return stats, err
	}

	q = s.DB.Query(`SELECT COUNT(*) FROM statement WHERE survey_id = ? AND status = 'pending'`, surveyID)
	if err := q.ScanRow(&stats.PendingStatements); err != nil {
		return stats, err
	}

	if stats.TotalStatements > 0 {
		q = s.DB.Query(`
			SELECT COUNT(*) FROM survey_participant sp
			WHERE sp.survey_id = ?
				AND (SELECT COUNT(*) FROM response r
					JOIN statement s ON s.id = r.statement_id
					WHERE s.survey_id = sp.survey_id AND r.user_id = sp.user_id
				) >= ?`, surveyID, stats.TotalStatements)
		if err := q.ScanRow(&stats.CompletedCount); err != nil {
			return stats, err
		}
	}

	if stats.TotalParticipants > 0 {
		stats.CompletionRate = float64(stats.CompletedCount) / float64(stats.TotalParticipants) * 100
	}

	return stats, nil
}

func (s *Store) GetUserVotesForSurvey(ctx context.Context, surveyID, userID uint) (map[uint]model.UserVote, error) {
	type row struct {
		StatementID uint   `db:"statement_id"`
		Vote        string `db:"vote"`
		IsImportant bool   `db:"is_important"`
	}
	var rows []row
	q := s.DB.Query(`
		SELECT r.statement_id, r.vote, r.is_important
		FROM response r
		JOIN statement s ON s.id = r.statement_id
		WHERE s.survey_id = ? AND r.user_id = ?`, surveyID, userID)
	if err := q.All(&rows); err != nil {
		return nil, err
	}
	result := make(map[uint]model.UserVote, len(rows))
	for _, r := range rows {
		result[r.StatementID] = model.UserVote{Vote: r.Vote, IsImportant: r.IsImportant}
	}
	return result, nil
}

func (s *Store) GetStatementSurveyID(ctx context.Context, statementID uint) (uint, error) {
	var surveyID uint
	q := s.DB.Query(`SELECT survey_id FROM statement WHERE id = ?`, statementID)
	if err := q.ScanRow(&surveyID); err != nil {
		return 0, err
	}
	return surveyID, nil
}
