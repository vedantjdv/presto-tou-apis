package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Connect() (*pgxpool.Pool, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:root@localhost:5439/tou_db?sslmode=disable"
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
	migrationPath := "migrations/0001_init.sql"
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("unable to read migration file: %v", err)
	}

	_, err = pool.Exec(context.Background(), string(content))
	if err != nil {
		return fmt.Errorf("unable to execute migration: %v", err)
	}

	log.Println("Database schema initialized")
	return nil
}
