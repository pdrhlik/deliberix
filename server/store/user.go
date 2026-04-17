package store

import (
	"context"

	"github.com/pdrhlik/deliberix/server/model"
)

func (s *Store) CreateUser(ctx context.Context, u *model.User) (uint, error) {
	q := s.DB.Query(`INSERT INTO user ?values`, u)
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

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return queryOne[model.User](s.DB.Query(`SELECT * FROM user WHERE email = ?`, email))
}

func (s *Store) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	return queryOne[model.User](s.DB.Query(`SELECT * FROM user WHERE id = ?`, id))
}

func (s *Store) UpdateUserProfile(ctx context.Context, id uint, name, locale string) error {
	q := s.DB.Query(`UPDATE user SET name = ?, locale = ? WHERE id = ?`, name, locale, id)
	_, err := q.Exec()
	return err
}
