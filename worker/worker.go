package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DeepanshuMishraa/vid-processing-go.git/models"
	"github.com/DeepanshuMishraa/vid-processing-go.git/queue"
	"github.com/DeepanshuMishraa/vid-processing-go.git/repository"
	"github.com/DeepanshuMishraa/vid-processing-go.git/types"
	"github.com/DeepanshuMishraa/vid-processing-go.git/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

type rendition struct {
	filePath string
	destKey  string
	urlField *string
}

const maxRetries = 3

func Start(conn *amqp.Connection, db *pgxpool.Pool, r2Svc *types.R2Service) error {
	msgs, ch, err := queue.Consume(conn)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}
	defer ch.Close()

	log.Println("Worker started, waiting for jobs...")

	for msg := range msgs {
		job, err := parseJob(msg.Body)
		if err != nil {
			log.Printf("Invalid job, discarding: %v", err)
			msg.Nack(false, false) // discard
			continue
		}

		retryCount := deliveryCount(msg.Headers)
		if retryCount >= maxRetries {
			log.Printf("Job %s exceeded max retries (%d), discarding", job.VideoID, maxRetries)
			msg.Nack(false, false) // discard
			continue
		}

		log.Printf("Processing video %s (attempt %d)...", job.VideoID, retryCount+1)

		if err := processJob(db, r2Svc, job); err != nil {
			log.Printf("Job %s failed (attempt %d): %v — requeueing", job.VideoID, retryCount+1, err)
			msg.Nack(false, true) // requeue
			continue
		}

		msg.Ack(false)
		log.Printf("Video %s processed successfully", job.VideoID)
	}

	return nil
}

func deliveryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	raw, ok := headers["x-delivery-count"]
	if !ok {
		return 0
	}
	// quorum queues store x-delivery-count as int64
	switch v := raw.(type) {
	case int64:
		return int(v)
	case int32:
		return int(v)
	case int:
		return v
	}
	return 0
}

func parseJob(body []byte) (*types.VideoJob, error) {
	var job types.VideoJob
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("unmarshal job: %w", err)
	}
	if job.VideoID == "" {
		return nil, fmt.Errorf("job has empty video_id")
	}
	return &job, nil
}

func processJob(db *pgxpool.Pool, r2Svc *types.R2Service, job *types.VideoJob) error {
	// 1. Fetch video record
	video, err := repository.GetVideoById(db, job.VideoID)
	if err != nil {
		return fmt.Errorf("get video %s: %w", job.VideoID, err)
	}

	// 2. Mark as PROCESSING
	video.Status = models.PROCESSING
	if err := repository.UpdateVideo(db, *video); err != nil {
		return fmt.Errorf("mark processing: %w", err)
	}

	// 3. Extract R2 key from the original URL and download
	key, err := keyFromURL(video.OriginalURL)
	if err != nil {
		setFailed(db, video)
		return fmt.Errorf("parse original url: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	body, err := utils.Download(ctx, r2Svc, key)
	if err != nil {
		setFailed(db, video)
		return fmt.Errorf("download from r2: %w", err)
	}
	defer body.Close()

	// 4. Write to a temp file for ffmpeg
	tmpFile, err := os.CreateTemp("", "vid-pro-*.mp4")
	if err != nil {
		setFailed(db, video)
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, body); err != nil {
		tmpFile.Close()
		setFailed(db, video)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		setFailed(db, video)
		return fmt.Errorf("close temp file: %w", err)
	}

	// 5. Transcode to 360p, 480p, 720p, 1080p
	outputs, err := utils.TranscodeVideo(tmpFile.Name())
	if err != nil {
		setFailed(db, video)
		return fmt.Errorf("transcode: %w", err)
	}

	// 6. Upload each rendition to R2 and update URLs
	baseKey := strings.TrimSuffix(key, filepath.Ext(key))
	ext := ".mp4"

	renditions := []rendition{
		{outputs.Video360, fmt.Sprintf("%s_360%s", baseKey, ext), &video.Video360URL},
		{outputs.Video480, fmt.Sprintf("%s_480%s", baseKey, ext), &video.Video480URL},
		{outputs.Video720, fmt.Sprintf("%s_720%s", baseKey, ext), &video.Video720URL},
		{outputs.Video1080, fmt.Sprintf("%s_1080%s", baseKey, ext), &video.Video1080URL},
	}

	for _, r := range renditions {
		f, err := os.Open(r.filePath)
		if err != nil {
			setFailed(db, video)
			return fmt.Errorf("open rendition %s: %w", r.filePath, err)
		}

		if err := utils.Upload(ctx, r2Svc, r.destKey, f); err != nil {
			f.Close()
			setFailed(db, video)
			return fmt.Errorf("upload rendition %s: %w", r.destKey, err)
		}
		f.Close()

		*r.urlField = publicURL(r2Svc, r.destKey)
	}

	// 7. Clean up renditions from /tmp
	for _, r := range renditions {
		os.Remove(r.filePath)
	}

	// 8. Mark as READY
	video.Status = models.READY
	if err := repository.UpdateVideo(db, *video); err != nil {
		return fmt.Errorf("mark ready: %w", err)
	}

	return nil
}

func keyFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// path: /<bucket>/videos/<id>.ext
	parts := strings.SplitN(u.Path, "/", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected r2 url path: %q", u.Path)
	}
	return parts[2], nil
}

func publicURL(svc *types.R2Service, key string) string {
	return fmt.Sprintf("%s/%s/%s",
		svc.PublicDomain, svc.Bucket, key)
}

func setFailed(db *pgxpool.Pool, video *models.Video) {
	video.Status = models.FAILED
	if err := repository.UpdateVideo(db, *video); err != nil {
		log.Printf("Failed to mark video %s as failed: %v", video.ID, err)
	}
}
