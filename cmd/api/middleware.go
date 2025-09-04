package main

import (
	"errors"
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
	"simplewebapi.moviedb/internal/data"
	"simplewebapi.moviedb/internal/validator"
	"strings"
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
func (app *application) authenticate(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authentication")

		authorizationHeader := r.Header.Get("Authorization")

		//if authorizationHeader == "" {
		//	r = app.contextSetUser(r, data.AnonymousUser)
		//	next.ServeHTTP(w, r)
		//	return
		//}
		if authorizationHeader == "" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		parts := strings.Split(authorizationHeader, " ")

		if parts[0] != "Bearer" || len(parts) != 2 {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		token := parts[1]
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		user, err := app.repos.Users.GetForToken(data.ScopeAuthentication, token)

		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		app.contextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}
