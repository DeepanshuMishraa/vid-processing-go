package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

func Connect(dbUrl string) (*pgxpool.Pool, error) {
	ctx := context.Background()

	cfg, err := pgxpool.ParseConfig(dbUrl)

	if err != nil {
		log.Println("Unable to parse db url")
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)

	if err != nil {
		log.Println("Unable to connect to db")
		pool.Close()
		return nil, err
	}

	log.Println("Connected to db")

	return pool, nil

}
