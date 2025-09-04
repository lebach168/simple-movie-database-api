package main

import (
	"errors"
	"net/http"
	"simplewebapi.moviedb/internal/data"
	"simplewebapi.moviedb/internal/validator"
)

type UserInput struct {
	Name     string
	Email    string
	Password string
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input UserInput

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var user data.User
	userMapper(input, &user)

	v := validator.New()
	if data.ValidateUser(v, &user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.repos.Users.Insert(&user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("Email", "a user with this Email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func userMapper(input UserInput, u *data.User) {
	if u == nil {
		u = &data.User{}
	}
	u.Name = input.Name
	u.Email = input.Email
	u.Password.Set(input.Password)
}
