package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Conn(DSN string) (db *pgxpool.Pool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err = pgxpool.New(ctx, DSN)
	if err != nil {
		panic(err)
	}

	return
}
