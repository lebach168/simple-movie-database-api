package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /healthcheck", app.healthcheckHandler)

	router.HandleFunc("GET /movies", app.listMovieHandler)
	router.HandleFunc("POST /movies", app.createMovieHandler)
	router.HandleFunc("GET /movies/{id}", app.showMovieHandler)
	router.HandleFunc("PATCH /movies/{id}", app.updateMovieHandler)
	router.HandleFunc("DELETE /movies/{id}", app.deleteMovieHandler)

	return router
}
