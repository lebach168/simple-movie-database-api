package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /healthcheck", app.healthcheckHandler)
	router.HandleFunc("POST /movies", app.createMovieHandler)
	router.HandleFunc("GET /movies/{id}", app.showMovieHandler)

	return router
}
