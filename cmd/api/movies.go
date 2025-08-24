package main

import (
	"fmt"
	"net/http"
	"simplewebapi.moviedb/internal/data"
	"simplewebapi.moviedb/internal/validator"
	"time"
)

type MovieInput struct {
	Title   string       `json:"title"`
	Year    int32        `json:"year"`
	Runtime data.Runtime `json:"runtime"`
	Genres  []string     `json:"genres"`
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	var input MovieInput

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()

	if !ValidateMovie(v, &input) {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v \n", input)
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}
	data := envelope{"movie": movie}
	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func ValidateMovie(v *validator.Validator, movieInput *MovieInput) bool {
	v.Check(movieInput.Title != "", "title", "must not be empty")
	v.Check(len(movieInput.Title) <= 500, "title", "must not be more than 500 chars long")

	v.Check(movieInput.Year != 0, "year", "must be provided")
	v.Check(movieInput.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movieInput.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movieInput.Runtime != 0, "runtime", "must be provided")
	v.Check(movieInput.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movieInput.Genres != nil, "genres", "must be provided")
	v.Check(len(movieInput.Genres) >= 1, "genres", "must contain at least 1 genre")

	v.Check(validator.Unique(movieInput.Genres), "genres", "must not contain duplicate values")
	return v.Valid()
}
