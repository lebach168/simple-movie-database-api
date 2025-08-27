package main

import (
	"errors"
	"fmt"
	"net/http"
	"simplewebapi.moviedb/internal/data"
	"simplewebapi.moviedb/internal/validator"
)

type MovieInput struct {
	Title   *string       `json:"title"`
	Year    *int32        `json:"year"`
	Runtime *data.Runtime `json:"runtime"`
	Genres  []string      `json:"genres"`
}
type QueryInput struct {
	Title  string
	Genres []string
	data.Filter
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	var input MovieInput

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var movie *data.Movie
	movieMapper(input, movie)
	v := validator.New()

	if !data.ValidateMovie(v, movie) {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.repos.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	movie, err := app.repos.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	data := envelope{"movie": movie}
	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	movie, err := app.repos.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	var input MovieInput

	err = app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movieMapper(input, movie)
	v := validator.New()

	if !data.ValidateMovie(v, movie) {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.repos.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.repos.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) listMovieHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query() // query string: map[string] []string

	var input QueryInput
	v := validator.New()
	input.Title = readString(qs, "title", "")
	input.Genres = readCSV(qs, "genres", []string{})

	input.Page = readInt(qs, "page", 1, v)
	input.PageSize = readInt(qs, "page_size", 20, v)
	input.Sort = readString(qs, "sort", "id")

	sortFields := []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}
	v.Check(input.Page > 0, "page", "must be greater than zero")
	v.Check(input.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(input.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.Check(validator.In(input.Sort, sortFields...), "sort", fmt.Sprintf("invalid sort %s field", input.Sort))

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	movies, metadata, err := app.repos.Movies.GetAll(input.Title, input.Genres, input.Filter)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
func movieMapper(input MovieInput, movie *data.Movie) {
	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}
}
