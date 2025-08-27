package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"strings"
	"time"
)

type MoviesRepoInterface interface {
	Insert(movie *Movie) error
	Get(id int64) (*Movie, error)
	GetAll(title string, genres []string, filter Filter) ([]*Movie, Metadata, error)
	Update(movie *Movie) error
	Delete(id int64) error
}
type MoviesRepo struct {
	DB *sql.DB
}

func (repo MoviesRepo) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title,year,runtime,genres)
		VALUES ($1,$2,$3,$4)
		RETURNING id,created_at,version`

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := repo.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
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
func (repo MoviesRepo) Get(id int64) (*Movie, error) {
	if id <= 0 {
		return nil, ErrRecordNotFound
	}
	var movie Movie
	query := `SELECT id,created_at,title,year,runtime, genres,version
			FROM movies
			WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	err := repo.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

func (repo MoviesRepo) GetAll(title string, genres []string, filter Filter) ([]*Movie, Metadata, error) {

	sortBy := fmt.Sprintf("%s %s", strings.TrimPrefix(filter.Sort, "-"), filter.sortDirection())
	limit := filter.limit()
	offset := filter.offset()

	query := fmt.Sprintf(`SELECT  count(*) OVER(),id, created_at, title, year, runtime, genres, version
        FROM movies
        WHERE (to_tsvector('simple',title) @@ plainto_tsquery('simple',$1) OR $1='')
        	AND ((genres @> $2) OR $2='{}')
        ORDER BY %s
        LIMIT %v OFFSET %v`, sortBy, limit, offset)
	// genres && $2 : nếu cần exists in
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []interface{}{title, pq.Array(genres)}
	rows, err := repo.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	movies := make([]*Movie, 0)
	var totalRecords int
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		movies = append(movies, &movie)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := NewMetadata(totalRecords, filter.Page, filter.PageSize)
	return movies, metadata, nil
}

func (repo MoviesRepo) Update(movie *Movie) error {
	query := `
        UPDATE movies 
        SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
        WHERE id = $5 AND version = $6
        RETURNING version`

	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := repo.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
func (repo MoviesRepo) Delete(id int64) error {
	query := `DELETE FROM movies
			WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := repo.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
