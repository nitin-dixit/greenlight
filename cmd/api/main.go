package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nitin-dixit/greenlight/internal/data"
	"github.com/nitin-dixit/greenlight/internal/jsonlog"
)

const version = "1.0.0"

// configuration settings for application, read from command-line flags
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// holds dependencies for HTTP handler, helpers, and middlewares.
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	var cfg config

	// read value of port flag into config struct, default port 4000
	flag.IntVar(&cfg.port, "port", 4000, "API server port")

	// read value of env flag into config struct, default is development
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production")

	flag.StringVar(&cfg.db.dsn, "dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")

	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()

	// initialize new jsonlog.Logger which writes any messages *at or above" the INFO
	// severity level to the standard out stream.
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	dbPool, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer dbPool.Close()
	logger.PrintInfo("database connection pool established", nil)
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(dbPool),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
		// Create the new log.Logger isntance with the log.New() function passing in
		// out custom Logger as the first parameter. The "" and 0 indicate that the
		// log.Logger isntance should not use prefix or any flags.
		ErrorLog:     log.New(logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  cfg.env,
	})
	err = srv.ListenAndServe()

	logger.PrintFatal(err, nil)
}

func openDB(cfg config) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.db.dsn)
	if err != nil {
		return nil, fmt.Errorf("db: parse config %w", err)
	}

	config.MaxConns = int32(cfg.db.maxOpenConns)
	config.MinConns = int32(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	config.MaxConnIdleTime = duration

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("db: open %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping %w", err)
	}

	return pool, nil
}
