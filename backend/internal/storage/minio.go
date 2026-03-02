package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"media-cms/internal/config"
)

// MinIOStorage implements ObjectStorage using MinIO.
type MinIOStorage struct {
	client         *minio.Client // internal (minio:9000) – for upload/delete
	urlClient      *minio.Client // public endpoint – for presigned URL generation only
	publicEndpoint string
	signedExpiry   time.Duration
}

// NewMinIOStorage creates and returns a configured MinIOStorage.
func NewMinIOStorage(cfg config.MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: failed to create client: %w", err)
	}

	// Build a second client pointed at the public endpoint so that
	// PresignedGetObject signs URLs with the public hostname.
	// This client never makes real network calls – it only computes HMACs.
	urlClient := client
	log.Printf("[minio] internal endpoint=%q publicEndpoint=%q", cfg.Endpoint, cfg.PublicEndpoint)
	if cfg.PublicEndpoint != "" {
		pub, parseErr := url.Parse(cfg.PublicEndpoint)
		log.Printf("[minio] parsed: host=%q scheme=%q err=%v", pub.Host, pub.Scheme, parseErr)
		if parseErr == nil && pub.Host != "" {
			useSSL := pub.Scheme == "https"
			uc, ucErr := minio.New(pub.Host, &minio.Options{
				Creds:        credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
				Secure:       useSSL,
				Region:       "us-east-1",
				BucketLookup: minio.BucketLookupPath,
			})
			log.Printf("[minio] urlClient creation: host=%q ucErr=%v", pub.Host, ucErr)
			if ucErr == nil {
				urlClient = uc
				log.Printf("[minio] urlClient set to public endpoint %q", pub.Host)
			}
		}
	} else {
		log.Printf("[minio] WARNING: PublicEndpoint empty, signed URLs will use internal hostname")
	}

	return &MinIOStorage{
		client:         client,
		urlClient:      urlClient,
		publicEndpoint: cfg.PublicEndpoint,
		signedExpiry:   cfg.SignedURLExpiry,
	}, nil
}

// EnsureBucket creates the bucket if it does not exist.
func (s *MinIOStorage) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("minio: bucket check failed: %w", err)
	}
	if !exists {
		if err = s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("minio: make bucket failed: %w", err)
		}
	}

	// Allow anonymous read only for objects under public/*
	policy, err := buildPublicReadPolicy(bucket)
	if err != nil {
		return fmt.Errorf("minio: build bucket policy failed: %w", err)
	}
	if err = s.client.SetBucketPolicy(ctx, bucket, policy); err != nil {
		return fmt.Errorf("minio: set bucket policy failed: %w", err)
	}
	return nil
}

// Upload streams r directly into MinIO with no intermediate buffering.
func (s *MinIOStorage) Upload(
	ctx context.Context,
	bucket, key string,
	r io.Reader,
	size int64,
	mimeType string,
) error {
	opts := minio.PutObjectOptions{
		ContentType:  mimeType,
		UserMetadata: map[string]string{},
	}

	// size -1 → unknown → MinIO will use chunked transfer
	_, err := s.client.PutObject(ctx, bucket, key, r, size, opts)
	if err != nil {
		return fmt.Errorf("minio: upload failed [bucket=%s key=%s]: %w", bucket, key, err)
	}
	return nil
}

// Delete removes an object from the bucket.
func (s *MinIOStorage) Delete(ctx context.Context, bucket, key string) error {
	err := s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio: delete failed [bucket=%s key=%s]: %w", bucket, key, err)
	}
	return nil
}

// GetSignedURL returns a pre-signed GET URL valid for expire duration.
// urlClient is pointed at the public endpoint so the signature and hostname
// are both correct for browser access.
func (s *MinIOStorage) GetSignedURL(ctx context.Context, bucket, key string, expire time.Duration) (string, error) {
	u, err := s.urlClient.PresignedGetObject(ctx, bucket, key, expire, nil)
	if err != nil {
		log.Printf("[minio] GetSignedURL error bucket=%s key=%s: %v", bucket, key, err)
		return "", fmt.Errorf("minio: presign failed [bucket=%s key=%s]: %w", bucket, key, err)
	}
	log.Printf("[minio] GetSignedURL OK: %s", u.Host)
	return u.String(), nil
}

// GetPublicURL returns the permanent direct URL for a public object.
func (s *MinIOStorage) GetPublicURL(bucket, key string) string {
	return fmt.Sprintf("%s/%s/%s", s.publicEndpoint, bucket, key)
}

func buildPublicReadPolicy(bucket string) (string, error) {
	doc := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Sid":    "AllowPublicReadPrefix",
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					"AWS": []string{"*"},
				},
				"Action":   []string{"s3:GetObject"},
				"Resource": []string{fmt.Sprintf("arn:aws:s3:::%s/public/*", bucket)},
			},
		},
	}
	b, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
