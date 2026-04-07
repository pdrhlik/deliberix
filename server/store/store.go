package store

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mibk/dali"
)

type Store struct {
	DB *dali.DB
}

func queryOne[T any](q *dali.Query) (*T, error) {
	var v T
	switch err := q.One(&v); {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	}
	return &v, nil
}

func New(dsn string) (*Store, error) {
	db, err := dali.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Store{DB: db}, nil
}
