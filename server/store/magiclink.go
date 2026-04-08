package store

import (
	"context"
	"fmt"
	"time"
)

func (s *Store) CreateMagicLink(ctx context.Context, userID uint, token string, expiresAt time.Time) error {
	q := s.DB.Query(`INSERT INTO magic_link (user_id, token, expires_at) VALUES (?, ?, ?)`,
		userID, token, expiresAt)
	_, err := q.Exec()
	return err
}

func (s *Store) UseMagicLink(ctx context.Context, token string) (uint, error) {
	var userID uint
	var expiresAt time.Time
	q := s.DB.Query(`
		SELECT user_id, expires_at FROM magic_link
		WHERE token = ? AND used_at IS NULL`, token)
	if err := q.ScanRow(&userID, &expiresAt); err != nil {
		return 0, err
	}
	if time.Now().After(expiresAt) {
		return 0, fmt.Errorf("magic link expired (expiresAt=%v, now=%v)", expiresAt, time.Now())
	}

	q = s.DB.Query(`UPDATE magic_link SET used_at = NOW() WHERE token = ?`, token)
	_, err := q.Exec()
	return userID, err
}

func (s *Store) HasRecentMagicLink(ctx context.Context, userID uint, within time.Duration) (bool, error) {
	var count int
	q := s.DB.Query(`
		SELECT COUNT(*) FROM magic_link
		WHERE user_id = ? AND used_at IS NULL AND created_at > ?`, userID, time.Now().Add(-within))
	if err := q.ScanRow(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}
