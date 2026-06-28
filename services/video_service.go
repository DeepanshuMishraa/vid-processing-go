package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DeepanshuMishraa/vid-processing-go.git/models"
	"github.com/DeepanshuMishraa/vid-processing-go.git/repository"
	"github.com/DeepanshuMishraa/vid-processing-go.git/types"
	"github.com/DeepanshuMishraa/vid-processing-go.git/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

func CreateVideo(conn *amqp.Connection, db *pgxpool.Pool, r2Svc *types.R2Service, video models.Video, filePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open video file: %w", err)
	}
	defer file.Close()

	ext := filepath.Ext(filePath)
	key := fmt.Sprintf("videos/%s%s", video.ID, ext)

	if err := utils.Upload(ctx, r2Svc, key, file); err != nil {
		return fmt.Errorf("upload to r2: %w", err)
	}

	video.OriginalURL = fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s",
		r2Svc.AccountID, r2Svc.Bucket, key)

	if err := repository.CreateVideo(conn, db, video); err != nil {
		return fmt.Errorf("create video record: %w", err)
	}

	return nil
}
