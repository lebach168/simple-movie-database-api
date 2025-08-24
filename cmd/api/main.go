package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
}

type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config
	//default flag value
	flag.IntVar(&cfg.port, "port", 8000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialize a new logger which writes messages to the standard out stream,
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		config: cfg,
		logger: logger,
	}

	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", app.routes()))
	
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      v1,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	logger.Printf("Starting %s server on port %d:...", app.config.env, app.config.port)
	err := server.ListenAndServe()

	logger.Fatal(err)
}
