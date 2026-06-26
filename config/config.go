package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DATABASE_URL         string
	PORT                 string
	BUCKET_NAME          string
	R2_ACCOUNT_ID        string
	R2_ACCESS_KEY_ID     string
	R2_ACCESS_KEY_SECRET string
	RABBIT_MQ_URL        string
}

func (c *Config) LoadConfig() (*Config, error) {
	err := godotenv.Load()

	if err != nil {
		return nil, err
	}

	db := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	bucket_name := os.Getenv("BUCKET_NAME")
	r2_account_id := os.Getenv("R2_ACCOUNT_ID")
	r2_access_key_id := os.Getenv("R2_ACCESS_KEY_ID")
	r2_access_key_secret := os.Getenv("R2_ACCESS_KEY_SECRET")
	rabbit_mq := os.Getenv("RABBIT_MQ_URL")

	if db == "" || port == "" || bucket_name == "" || r2_account_id == "" || r2_access_key_id == "" || r2_access_key_secret == "" || rabbit_mq == "" {
		return nil, err
	}

	return &Config{
		DATABASE_URL:         db,
		PORT:                 port,
		BUCKET_NAME:          bucket_name,
		R2_ACCOUNT_ID:        r2_account_id,
		R2_ACCESS_KEY_ID:     r2_access_key_id,
		R2_ACCESS_KEY_SECRET: r2_access_key_secret,
		RABBIT_MQ_URL:        rabbit_mq,
	}, nil
}
