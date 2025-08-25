package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Repo struct {
	Movies MoviesRepoInterface
}

func NewRepo(db *sql.DB) Repo {
	return Repo{
		Movies: MoviesRepo{DB: db},
	}
}
