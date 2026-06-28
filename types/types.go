package types

import "github.com/aws/aws-sdk-go-v2/service/s3"

type R2Service struct {
	R2Client   *s3.Client
	Bucket     string
	AccountID  string
}
