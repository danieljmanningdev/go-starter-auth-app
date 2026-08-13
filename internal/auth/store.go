package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) CreateUser(
	ctx context.Context,
	email string,
	passwordHash string,
) (int64, error) {
	result, err := s.db.ExecContext(
		ctx,
		`
			INSERT INTO users (
				email,
				password_hash
			)
			VALUES (?, ?)
		`,
		email,
		passwordHash,
	)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read created user id: %w", err)
	}

	return id, nil
}

func (s *Store) FindUserByEmail(
	ctx context.Context,
	email string,
) (User, error) {
	var user User

	err := s.db.QueryRowContext(
		ctx,
		`
			SELECT
				id,
				email,
				password_hash
			FROM users
			WHERE email = ?
		`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}

		return User{}, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

func (s *Store) CreateSession(
	ctx context.Context,
	userID int64,
	tokenHash string,
	expiresAt time.Time,
) error {
	_, err := s.db.ExecContext(
		ctx,
		`
			INSERT INTO sessions (
				user_id,
				token_hash,
				expires_at
			)
			VALUES (?, ?, ?)
		`,
		userID,
		tokenHash,
		expiresAt,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}

func (s *Store) FindSessionUserID(
	ctx context.Context,
	tokenHash string,
) (int64, bool, error) {
	var userID int64

	err := s.db.QueryRowContext(
		ctx,
		`
			SELECT user_id
			FROM sessions
			WHERE token_hash = ?
			  AND expires_at > CURRENT_TIMESTAMP
		`,
		tokenHash,
	).Scan(&userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}

		return 0, false, fmt.Errorf("find session: %w", err)
	}

	return userID, true, nil
}

func (s *Store) DeleteSession(
	ctx context.Context,
	tokenHash string,
) error {
	_, err := s.db.ExecContext(
		ctx,
		`
			DELETE FROM sessions
			WHERE token_hash = ?
		`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}
