package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	host := os.Getenv("PGHOST")
	if host == "" {
		host = "/run/postgresql"
	}
	database := os.Getenv("PGDATABASE")
	if database == "" {
		database = "civ6"
	}
	user := os.Getenv("PGUSER")
	if user == "" {
		user = "civ6"
	}
	return pgxpool.New(ctx, fmt.Sprintf("host=%s dbname=%s user=%s", host, database, user))
}
