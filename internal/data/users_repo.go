package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type UsersRepoInterface interface {
	Insert(user *User) error
	GetByEmail(email string) (*User, error)
	Update(user *User) error
}

type UsersRepo struct {
	DB *sql.DB
}

func (repo UsersRepo) Insert(user *User) error {
	query := `INSERT INTO users (name, email, password_hash)
			VALUES ($1,$2,$3)
			RETURNING id, created_at,activated,version`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []interface{}{user.Name, user.Email, user.Password.hash}

	err := repo.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Activated, &user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}

	}
	return nil
}

func (repo UsersRepo) GetByEmail(email string) (*User, error) {
	query := `SELECT id,name,email,password_hash,activated,version FROM users
			WHERE email= $1`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user User
	args := []interface{}{email}
	err := repo.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (repo UsersRepo) Update(user *User) error {
	query := `UPDATE users
			SET name=$1,email=$2,password_hash=$3,activated=$4,version =version+1
			WHERE id = $5 AND version = $6
			RETURNING version`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}
	err := repo.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
