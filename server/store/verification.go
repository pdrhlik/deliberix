package store

import (
	"context"
	"fmt"
	"time"
)

func (s *Store) HasRecentVerification(ctx context.Context, userID uint, within time.Duration) (bool, error) {
	var count int
	q := s.DB.Query(`
		SELECT COUNT(*) FROM email_verification
		WHERE user_id = ? AND used_at IS NULL AND created_at > ?`, userID, time.Now().Add(-within))
	if err := q.ScanRow(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) CreateEmailVerification(ctx context.Context, userID uint, token string, expiresAt time.Time) error {
	q := s.DB.Query(`INSERT INTO email_verification (user_id, token, expires_at) VALUES (?, ?, ?)`,
		userID, token, expiresAt)
	_, err := q.Exec()
	return err
}

func (s *Store) UseEmailVerification(ctx context.Context, token string) (uint, error) {
	var userID uint
	var expiresAt time.Time
	q := s.DB.Query(`
		SELECT user_id, expires_at FROM email_verification
		WHERE token = ? AND used_at IS NULL`, token)
	if err := q.ScanRow(&userID, &expiresAt); err != nil {
		return 0, err
	}
	if time.Now().After(expiresAt) {
		return 0, fmt.Errorf("verification token expired")
	}

	// Mark all tokens for this user as used (the clicked one + any older ones)
	q = s.DB.Query(`UPDATE email_verification SET used_at = NOW() WHERE user_id = ? AND used_at IS NULL`, userID)
	if _, err := q.Exec(); err != nil {
		return 0, err
	}

	q = s.DB.Query(`UPDATE user SET email_verified_at = NOW() WHERE id = ?`, userID)
	if _, err := q.Exec(); err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *Store) CreatePasswordReset(ctx context.Context, userID uint, token string, expiresAt time.Time) error {
	q := s.DB.Query(`INSERT INTO password_reset (user_id, token, expires_at) VALUES (?, ?, ?)`,
		userID, token, expiresAt)
	_, err := q.Exec()
	return err
}

func (s *Store) UsePasswordReset(ctx context.Context, token string) (uint, error) {
	var userID uint
	var expiresAt time.Time
	q := s.DB.Query(`
		SELECT user_id, expires_at FROM password_reset
		WHERE token = ? AND used_at IS NULL`, token)
	if err := q.ScanRow(&userID, &expiresAt); err != nil {
		return 0, err
	}
	if time.Now().After(expiresAt) {
		return 0, fmt.Errorf("reset token expired")
	}

	q = s.DB.Query(`UPDATE password_reset SET used_at = NOW() WHERE token = ?`, token)
	_, err := q.Exec()
	return userID, err
}

func (s *Store) VerifyUserEmail(ctx context.Context, userID uint) error {
	q := s.DB.Query(`UPDATE user SET email_verified_at = NOW() WHERE id = ?`, userID)
	_, err := q.Exec()
	return err
}

func (s *Store) UpdateUserPassword(ctx context.Context, userID uint, passwordHash string) error {
	q := s.DB.Query(`UPDATE user SET password_hash = ? WHERE id = ?`, passwordHash, userID)
	_, err := q.Exec()
	return err
}
