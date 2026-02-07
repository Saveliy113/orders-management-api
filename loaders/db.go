package loaders

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://postgres:postgres@localhost:5432/orders_management?sslmode=disable"
		log.Println("Using default DATABASE_URL (set DATABASE_URL env for production)")
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}

	ctx := context.Background()
	DB, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}

	if err := DB.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Database connection pool initialized successfully")
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection pool closed")
	}
}
