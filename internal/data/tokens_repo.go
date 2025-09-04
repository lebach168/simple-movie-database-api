package data

import (
	"context"
	"database/sql"
	"time"
)

type TokensRepoInterface interface {
	Insert(token *Token) error
	New(userID int64, ttl time.Duration, scope string) (*Token, error)
	DeleteAllForUser(scope string, userID int64) error
}

type TokensRepo struct {
	DB *sql.DB
}

func (repo TokensRepo) Insert(token *Token) error {
	query := `INSERT INTO tokens (hash,user_id, expiry,scope )
			VALUES ($1,$2,$3,$4)`

	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := repo.DB.ExecContext(ctx, query, args...)
	return err
}

func (repo TokensRepo) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := GenerateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = repo.DeleteAllForUser(ScopeAuthentication, userID)
	err = repo.Insert(token)

	return token, err
}

func (repo TokensRepo) DeleteAllForUser(scope string, userID int64) error {
	query := `DELETE FROM tokens
			WHERE scope like $1 AND user_id = $2 `
	args := []interface{}{scope, userID}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := repo.DB.ExecContext(ctx, query, args...)

	return err

}
