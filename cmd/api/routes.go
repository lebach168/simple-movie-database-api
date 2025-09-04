package main

import "net/http"

func (app *application) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /healthcheck", app.healthcheckHandler)
	router.HandleFunc("POST /users", app.registerUserHandler)
	router.HandleFunc("POST /users/authentication", app.authenticationHandler)

	router.HandleFunc("GET /movies", app.authenticate(http.HandlerFunc(app.listMovieHandler)))
	router.HandleFunc("POST /movies", app.createMovieHandler)
	router.HandleFunc("GET /movies/{id}", app.showMovieHandler)
	router.HandleFunc("PATCH /movies/{id}", app.updateMovieHandler)
	router.HandleFunc("DELETE /movies/{id}", app.deleteMovieHandler)

	return router
}
