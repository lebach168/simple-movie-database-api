package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
)

type Middleware func(handler http.Handler) http.Handler

func CreateChain(m ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(m) - 1; i >= 0; i-- {
			x := m[i]
			next = x(next)
		}
		return next
	}

}

func (app *application) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}

		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) RateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled && !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
