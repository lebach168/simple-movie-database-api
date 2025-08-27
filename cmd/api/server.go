package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {

	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", app.routes()))

	middlewareChain := CreateChain(
		app.LoggingHTTPHandler,
		app.RecoverPanic,
		app.RateLimit,
	)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      middlewareChain(v1),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		signal := <-quit

		app.logger.Info("Shutting down server\n", "signal: ", signal.String())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("Completing background tasks...")

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info(fmt.Sprintf("Starting %s server on port %d:...", app.config.env, app.config.port))

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}
	app.logger.Info("Server exited")
	return nil
}
