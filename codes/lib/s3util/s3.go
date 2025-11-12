package s3util

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/venue-master/platform/lib/config"
)

// StorageProvider defines the behavior shared across services for object storage.
type StorageProvider interface {
	Upload(ctx context.Context, key string, body []byte, contentType string) (string, error)
	Delete(ctx context.Context, key string) error
	PresignGet(ctx context.Context, key string, expires time.Duration) (string, error)
}

// S3Provider implements StorageProvider using AWS S3 (or compatible endpoints like Localstack).
type S3Provider struct {
	client   *s3.Client
	presign  *s3.PresignClient
	bucket   string
	endpoint string
}

// New creates an S3Provider wired to the configured bucket/region.
func New(ctx context.Context, cfg config.AWSConfig) (*S3Provider, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = awsString(cfg.Endpoint)
		}
	})

	return &S3Provider{
		client:   client,
		presign:  s3.NewPresignClient(client),
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
	}, nil
}

// Upload pushes bytes to S3 and returns the absolute object URL.
func (p *S3Provider) Upload(ctx context.Context, key string, body []byte, contentType string) (string, error) {
	_, err := p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &p.bucket,
		Key:         &key,
		Body:        bytes.NewReader(body),
		ContentType: awsString(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return "", err
	}

	return p.objectURL(key), nil
}

// Delete removes an object from S3.
func (p *S3Provider) Delete(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &p.bucket, Key: &key})
	return err
}

// PresignGet generates a signed GET URL valid for the provided duration.
func (p *S3Provider) PresignGet(ctx context.Context, key string, expires time.Duration) (string, error) {
	out, err := p.presign.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: &p.bucket, Key: &key}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (p *S3Provider) objectURL(key string) string {
	if p.endpoint == "" {
		return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", p.bucket, url.PathEscape(key))
	}
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(p.endpoint, "/"), p.bucket, url.PathEscape(key))
}

func awsString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
