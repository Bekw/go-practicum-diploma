package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type User struct {
	ID        int64
	Login     string
	Password  string
	CreatedAt time.Time
}

var (
	ErrUserNotFound = errors.New("user not found")
)

func (s *Storage) CreateUser(ctx context.Context, login, passwordHash string) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(
		ctx,
		`INSERT INTO users(login, password) VALUES($1, $2) RETURNING id`,
		login,
		passwordHash,
	).Scan(&id)

	return id, err
}

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, login, password, created_at FROM users WHERE login = $1`,
		login,
	)

	var u User
	if err := row.Scan(&u.ID, &u.Login, &u.Password, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (s *Storage) IsLoginTaken(ctx context.Context, login string) (bool, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT 1 FROM users WHERE login = $1 LIMIT 1`,
		login,
	)

	var dummy int
	err := row.Scan(&dummy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
