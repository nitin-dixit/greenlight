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

// configuration settings for application, read from command-line flags
type config struct {
	port int
	env  string
}

// holds dependencies for HTTP handler, helpers, and middlewares.
type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config

	// read value of port flag into config struct, default port 4000
	flag.IntVar(&cfg.port, "port", 4000, "API server port")

	// read value of env flag into config struct, default is development
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production")
	flag.Parse()

	// initialize new logger which writes messages to the standard out stream,
	// prefixed with currect date and time
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		config: cfg,
		logger: logger,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err := srv.ListenAndServe()

	logger.Fatal(err)
}
