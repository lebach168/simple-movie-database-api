package main

import (
	"errors"
	"net/http"
	"simplewebapi.moviedb/internal/data"
	"simplewebapi.moviedb/internal/validator"
	"time"
)

type CredentialsInput struct {
	Email    string
	Password string
}

func (app *application) authenticationHandler(w http.ResponseWriter, r *http.Request) {
	var input CredentialsInput
	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(validator.Match(input.Email, validator.EmailRX), "Email", "invalid Email address")
	v.Check(input.Password != "", "Password", "Password must be provided")
	if !v.Valid() {
		app.invalidCredentialsResponse(w, r)
		return
	}

	user, err := app.repos.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)

		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	ok, err := user.Password.Match(input.Password)
	if !ok {
		app.invalidCredentialsResponse(w, r)
		return
	}
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	token, err := app.repos.Tokens.New(user.ID, 1*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
