package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Connect() (*pgxpool.Pool, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:root@localhost:5432/tou_db?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DATABASE_URL: %v", err)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	log.Println("Connected to PostgreSQL")
	return pool, nil
}

func InitSchema(pool *pgxpool.Pool) error {
	matches, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return fmt.Errorf("unable to find migration files: %v", err)
	}

	for _, path := range matches {
		log.Printf("Executing migration: %s", path)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("unable to read migration file %s: %v", path, err)
		}

		_, err = pool.Exec(context.Background(), string(content))
		if err != nil {
			return fmt.Errorf("unable to execute migration %s: %v", path, err)
		}
	}

	log.Println("Database schema initialized")
	return nil
}
