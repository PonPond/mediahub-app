package storage

import (
	"context"
	"io"
	"time"
)

// ObjectStorage is the abstraction over any S3-compatible backend.
type ObjectStorage interface {
	// Upload streams an object from r into the given bucket/key.
	// size may be -1 when the length is unknown (streaming chunked transfer).
	Upload(ctx context.Context, bucket, key string, r io.Reader, size int64, mimeType string) error

	// Delete removes an object permanently.
	Delete(ctx context.Context, bucket, key string) error

	// GetSignedURL returns a pre-signed URL valid for expire duration.
	GetSignedURL(ctx context.Context, bucket, key string, expire time.Duration) (string, error)

	// GetPublicURL returns the permanent public CDN / direct URL.
	GetPublicURL(bucket, key string) string

	// EnsureBucket creates the bucket if it does not already exist.
	EnsureBucket(ctx context.Context, bucket string) error
}
