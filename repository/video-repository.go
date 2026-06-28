package repository

import (
	"context"
	"time"

	"github.com/DeepanshuMishraa/vid-processing-go.git/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateVideo(db *pgxpool.Pool, video models.Video) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now().UTC()
	query := `INSERT INTO videos (id, title, original_url, status, created_at, updated_at)
	           VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.Exec(ctx, query,
		video.ID,
		video.Title,
		video.OriginalURL,
		models.UPLOADED,
		now,
		now,
	)
	if err != nil {
		return err
	}

	return nil
}
