// Package storage stores documents in an S3-compatible bucket: MinIO locally,
// GCS (through its S3 interoperability API) on GCP.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/sektor-ai/material-review-starter/go/internal/config"
)

type Storage struct {
	client *minio.Client
	bucket string
	region string
}

func New(cfg config.S3) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("creating s3 client: %w", err)
	}
	return &Storage{client: client, bucket: cfg.Bucket, region: cfg.Region}, nil
}

// EnsureBucket creates the bucket if it does not exist.
func (s *Storage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("checking bucket %s: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{Region: s.region}); err != nil {
		return fmt.Errorf("creating bucket %s: %w", s.bucket, err)
	}
	return nil
}

func (s *Storage) Upload(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(body), int64(len(body)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("uploading %s: %w", key, err)
	}
	return nil
}

// Get returns the object's bytes and content type.
func (s *Storage) Get(ctx context.Context, key string) ([]byte, string, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("getting %s: %w", key, err)
	}
	defer obj.Close()
	body, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", fmt.Errorf("reading %s: %w", key, err)
	}
	info, err := obj.Stat()
	if err != nil {
		return nil, "", fmt.Errorf("stat %s: %w", key, err)
	}
	return body, info.ContentType, nil
}

// DownloadURL returns a presigned GET URL. If filename is set, the browser
// downloads the object under that name.
func (s *Storage) DownloadURL(ctx context.Context, key, filename string, expires time.Duration) (string, error) {
	params := url.Values{}
	if filename != "" {
		params.Set("response-content-disposition", fmt.Sprintf("attachment; filename=%q", filename))
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, expires, params)
	if err != nil {
		return "", fmt.Errorf("presigning %s: %w", key, err)
	}
	return u.String(), nil
}
