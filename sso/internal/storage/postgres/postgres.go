package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"loginform/sso/internal/domain/models"
	"loginform/sso/internal/storage"
)

type Storage struct {
	pool *pgxpool.Pool
}

// New creates a new instance of the SQLite storage
func New(pool *pgxpool.Pool) *Storage {
	return &Storage{
		pool: pool,
	}
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const q = `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`
	var id int64

	err := s.pool.QueryRow(ctx, q, email, passHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
			return 0, fmt.Errorf("storage.postgres.SaveUser: %w", storage.ErrUserExists)
		}

		return 0, fmt.Errorf("storage.postgres.SaveUser: %w", err)
	}

	return id, nil
}

// User returns user by email
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const q = `SELECT id, email, pass_hash FROM users WHERE lower(email) = lower($1)`
	var user models.User

	err := s.pool.QueryRow(ctx, q, email).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		// no rows -> ErrUserNotFound
		return models.User{}, fmt.Errorf("storage.postgres.User: %w", storage.ErrUserNotFound)
	}

	return user, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const q = `SELECT id, name, secret FROM apps WHERE id = $1`

	var app models.App

	err := s.pool.QueryRow(ctx, q, appID).Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		return models.App{}, fmt.Errorf("storage.postgres.App: %w", storage.ErrAppNotFound)
	}

	return app, nil
}
