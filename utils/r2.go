package utils

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/DeepanshuMishraa/vid-processing-go.git/config"
	"github.com/DeepanshuMishraa/vid-processing-go.git/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewR2Service(cfg *config.Config) (*types.R2Service, error) {
	bucket := cfg.BUCKET_NAME
	accountID := cfg.R2_ACCOUNT_ID
	accessKeySecret := cfg.R2_ACCESS_KEY_SECRET
	accessKeyID := cfg.R2_ACCESS_KEY_ID

	cfx, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, "")),
		awsconfig.WithRegion("apac"),
	)
	if err != nil {
		return nil, err
	}

	r2Client := s3.NewFromConfig(cfx, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
	})

	return &types.R2Service{
		R2Client:  r2Client,
		Bucket:    bucket,
		AccountID: accountID,
	}, nil
}

var videoContentTypes = map[string]string{
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".mov":  "video/quicktime",
	".avi":  "video/x-msvideo",
	".mkv":  "video/x-matroska",
	".m4v":  "video/mp4",
	".ts":   "video/mp2t",
}

func contentTypeFromKey(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	if ct, ok := videoContentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

func Upload(ctx context.Context, svc *types.R2Service, key string, file io.Reader) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(svc.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentTypeFromKey(key)),
	}

	_, err := svc.R2Client.PutObject(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
