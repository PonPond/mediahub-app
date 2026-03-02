package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
)

// SHA256Reader wraps an io.Reader and computes the SHA-256 digest
// of all bytes read through it — zero extra memory for the file data.
type SHA256Reader struct {
	r io.Reader
	h hash.Hash
	n int64
}

// NewSHA256Reader returns a reader that hashes every byte on the fly.
func NewSHA256Reader(r io.Reader) *SHA256Reader {
	h := sha256.New()
	return &SHA256Reader{r: r, h: h}
}

// Read implements io.Reader.
func (s *SHA256Reader) Read(p []byte) (int, error) {
	n, err := s.r.Read(p)
	if n > 0 {
		_, _ = s.h.Write(p[:n])
		s.n += int64(n)
	}
	return n, err
}

// Checksum returns the final hex-encoded SHA-256 digest.
// Call this only after the reader has returned io.EOF.
func (s *SHA256Reader) Checksum() string {
	return hex.EncodeToString(s.h.Sum(nil))
}

// Size returns the total number of bytes read through this reader.
func (s *SHA256Reader) Size() int64 {
	return s.n
}
