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

func GetAllVideos(db *pgxpool.Pool) ([]models.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, title, original_url, status, created_at, updated_at
	           FROM videos`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []models.Video
	for rows.Next() {
		var video models.Video
		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.OriginalURL,
			&video.Status,
			&video.CreatedAt,
			&video.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	return videos, nil
}

func GetVideoById(db *pgxpool.Pool, id string) (*models.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, title, original_url, status, created_at, updated_at
	           FROM videos
	           WHERE id = $1`
	var video models.Video
	err := db.QueryRow(ctx, query, id).Scan(
		&video.ID,
		&video.Title,
		&video.OriginalURL,
		&video.Status,
		&video.CreatedAt,
		&video.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &video, nil

}
