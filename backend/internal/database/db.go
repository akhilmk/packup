package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func New(ctx context.Context) (*pgxpool.Pool, error) {
	// Load environment file `env.dev` if present
	// We might need to adjust the path if running from a different working dir,
	// but usually if run from root, just ".env.dev" or similar is fine.
	// The original code used "env.dev".
	// Try loading environment files from various locations
	envFiles := []string{"env.dev", ".env.dev", "../.env.dev", "../../.env.dev"}
	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			log.Printf("loaded env file: %s", file)
			break
		}
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASS")
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		name := os.Getenv("DB_NAME")
		ssl := os.Getenv("SSLMODE")

		if user == "" {
			user = "postgres"
		}
		if pass == "" {
			pass = "postgres"
		}
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5432"
		}
		if name == "" {
			name = "todos"
		}
		if ssl == "" {
			ssl = "disable"
		}

		dbURL = "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + name + "?sslmode=" + ssl
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	if err := ensureSchema(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func ensureSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		google_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		name TEXT,
		avatar_url TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		expires_at TIMESTAMPTZ NOT NULL
	);

	CREATE TABLE IF NOT EXISTS todos (
		id TEXT PRIMARY KEY,
		text TEXT NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		created TIMESTAMPTZ NOT NULL DEFAULT now(),
		position DOUBLE PRECISION NOT NULL DEFAULT 0,
		user_id TEXT REFERENCES users(id) ON DELETE CASCADE
	);
	
	-- Migrations
	ALTER TABLE todos ADD COLUMN IF NOT EXISTS position DOUBLE PRECISION NOT NULL DEFAULT 0;
	ALTER TABLE todos ADD COLUMN IF NOT EXISTS user_id TEXT REFERENCES users(id) ON DELETE CASCADE;
	`)
	return err
}
