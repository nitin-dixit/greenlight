package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
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
	logger *log.Logger
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

	// initialize new logger which writes messages to the standard out stream,
	// prefixed with currect date and time
	logger := log.New(os.Stdout, "", log.LstdFlags) //Ldate | Ltime // initial values for the standard logger

	db, _, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Printf("database connection pool established")
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

	logger.Printf("starting %s server on: http://localhost%s/", cfg.env, srv.Addr)
	err = srv.ListenAndServe()

	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, *pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.db.dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("db: open %w", err)
	}
	db := stdlib.OpenDBFromPool(pool)

	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("db: ping %w", err)
	}
	return db, pool, nil
}
