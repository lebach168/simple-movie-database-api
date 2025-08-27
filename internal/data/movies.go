package data

import (
	"errors"
	"fmt"
	"simplewebapi.moviedb/internal/validator"
	"strconv"
	"strings"
	"time"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres"`
	Version   int32     `json:"version"`
}

var ErrInvalidRuntimeFormat = errors.New("invalid Runtime field format") //Runtime tên riêng

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	rString := fmt.Sprintf("%d mins", r)

	quotedValue := strconv.Quote(rString)

	return []byte(quotedValue), nil
}
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))

	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	parts := strings.Split(unquotedJSONValue, " ")
	if len(parts) != 2 || (parts[1] != "mins" && parts[1] != "minutes") {
		return ErrInvalidRuntimeFormat
	}
	val, err := strconv.ParseInt(parts[0], 10, 32)

	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	*r = Runtime(val)

	return nil
}
func ValidateMovie(v *validator.Validator, movie *Movie) bool {
	v.Check(movie.Title != "", "title", "must not be empty")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 chars long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")

	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
	return v.Valid()
}
