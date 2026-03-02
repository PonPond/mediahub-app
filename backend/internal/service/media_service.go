package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"media-cms/internal/config"
	"media-cms/internal/model"
	"media-cms/internal/repository"
	"media-cms/internal/storage"
	"media-cms/internal/utils"
)

// ErrFileInUse is returned when a deletion is attempted on a referenced file.
var ErrFileInUse = errors.New("file is still referenced and cannot be deleted")

// ErrNotFound is returned when a media file does not exist.
var ErrNotFound = errors.New("media file not found")

// UploadInput carries all information needed for an upload.
type UploadInput struct {
	Reader        io.Reader
	FileName      string
	ContentType   string
	Size          int64 // -1 if unknown
	SourceService string
	SourceModule  string
	UploadedBy    string
	IsPublic      bool
	TokenType     string
	ProjectID     string
}

// MediaService defines the business operations on media files.
type MediaService interface {
	Upload(ctx context.Context, in UploadInput) (*model.MediaFile, error)
	List(ctx context.Context, params model.ListParams) (*model.ListResult, error)
	GetFilterOptions(ctx context.Context) (*model.FilterOptions, error)
	GetByID(ctx context.Context, id string) (*model.MediaFile, error)
	Delete(ctx context.Context, id string) error
	GetURL(ctx context.Context, m *model.MediaFile) (string, error)
	CleanupOrphans(ctx context.Context) (int, error)
}

type mediaService struct {
	repo     repository.MediaRepository
	authRepo repository.AuthRepository
	storage  storage.ObjectStorage
	cfg      *config.Config
	log      *zap.Logger
}

// NewMediaService wires up the service layer.
func NewMediaService(
	repo repository.MediaRepository,
	authRepo repository.AuthRepository,
	store storage.ObjectStorage,
	cfg *config.Config,
	log *zap.Logger,
) MediaService {
	return &mediaService{
		repo:     repo,
		authRepo: authRepo,
		storage:  store,
		cfg:      cfg,
		log:      log,
	}
}

// Upload validates, streams, and persists a file upload.
func (s *mediaService) Upload(ctx context.Context, in UploadInput) (*model.MediaFile, error) {
	var (
		out *model.MediaFile
		err error
	)
	if in.TokenType == "project" && in.ProjectID != "" {
		defer func() {
			status := "success"
			errMsg := ""
			if err != nil {
				status = "failed"
				errMsg = err.Error()
			}
			size := in.Size
			if out != nil && out.Size > 0 {
				size = out.Size
			}
			if size < 0 {
				size = 0
			}
			var mediaID *string
			if out != nil && out.ID != "" {
				mediaID = &out.ID
			}
			logErr := s.authRepo.CreateProjectUploadLog(ctx, model.ProjectUploadLog{
				ProjectID:     in.ProjectID,
				MediaID:       mediaID,
				FileName:      in.FileName,
				MimeType:      in.ContentType,
				Size:          size,
				SourceService: in.SourceService,
				SourceModule:  in.SourceModule,
				Status:        status,
				ErrorMessage:  errMsg,
				UploadedBy:    in.UploadedBy,
			})
			if logErr != nil {
				s.log.Warn("upload log failed", zap.Error(logErr), zap.String("project_id", in.ProjectID))
			}
		}()
	}

	maxUploadBytes := s.cfg.Upload.MaxFileSizeBytes

	if in.TokenType == "project" && in.ProjectID != "" {
		p, err := s.authRepo.GetProjectByID(ctx, in.ProjectID)
		if err != nil {
			err = fmt.Errorf("upload: load project policy: %w", err)
			return nil, err
		}
		group := utils.UploadGroup(in.FileName, in.ContentType)
		projectLimit := p.UploadPolicy.LimitBytes(group)
		if projectLimit <= 0 {
			err = fmt.Errorf("upload: %s files are disabled for this project", group)
			return nil, err
		}
		if maxUploadBytes <= 0 || projectLimit < maxUploadBytes {
			maxUploadBytes = projectLimit
		}
	}

	// 1. Validate MIME type
	if validateErr := utils.ValidateMIME(in.ContentType, s.cfg.Upload.AllowedMIMEs); validateErr != nil {
		err = fmt.Errorf("upload: %w", validateErr)
		return nil, err
	}

	// 2. Validate file size if known
	if in.Size > 0 && maxUploadBytes > 0 && in.Size > maxUploadBytes {
		err = fmt.Errorf("upload: file size %d exceeds limit %d", in.Size, maxUploadBytes)
		return nil, err
	}

	// 3. Wrap reader with a size-limiting and checksum-computing reader
	hashReader := utils.NewSHA256Reader(in.Reader)
	var limitedReader io.Reader = hashReader
	if maxUploadBytes > 0 {
		limitedReader = io.LimitReader(hashReader, maxUploadBytes+1)
	}

	// 4. Generate storage path with visibility prefix:
	// public/YYYY/MM/<uuid>.<ext> or private/YYYY/MM/<uuid>.<ext>
	ext := strings.ToLower(filepath.Ext(in.FileName))
	visibilityPrefix := "private"
	if in.IsPublic {
		visibilityPrefix = "public"
	}
	objectKey := fmt.Sprintf("%s/%s/%s%s",
		visibilityPrefix, time.Now().Format("2006/01"), uuid.New().String(), ext)

	bucket := s.cfg.MinIO.DefaultBucket

	// 5. Ensure bucket exists
	if err := s.storage.EnsureBucket(ctx, bucket); err != nil {
		err = fmt.Errorf("upload: ensure bucket: %w", err)
		return nil, err
	}

	// 6. Stream upload to MinIO (no memory buffering)
	if uploadErr := s.storage.Upload(ctx, bucket, objectKey, limitedReader, in.Size, in.ContentType); uploadErr != nil {
		err = fmt.Errorf("upload: object storage: %w", uploadErr)
		return nil, err
	}

	// 7. Compute final checksum (all bytes have been read)
	checksum := hashReader.Checksum()
	actualSize := hashReader.Size()

	// If input size was unknown, enforce max-size after streaming by checking
	// the counted bytes (LimitReader reads at most max+1 bytes).
	if in.Size <= 0 && maxUploadBytes > 0 && actualSize > maxUploadBytes {
		_ = s.storage.Delete(ctx, bucket, objectKey)
		err = fmt.Errorf("upload: file size exceeds limit %d", maxUploadBytes)
		return nil, err
	}

	// 8. Dedup check — reuse existing file if checksum matches
	if existing, err := s.repo.GetByChecksum(ctx, checksum); err == nil && existing != nil {
		s.log.Info("dedup hit, returning existing file",
			zap.String("id", existing.ID),
			zap.String("checksum", checksum))
		// Clean up the duplicate we just uploaded
		_ = s.storage.Delete(ctx, bucket, objectKey)

		existing.URL, _ = s.GetURL(ctx, existing)
		out = existing
		return out, nil
	}

	// 9. Persist metadata
	meta := &model.MediaFile{
		Bucket:        bucket,
		ObjectKey:     objectKey,
		FileName:      in.FileName,
		MimeType:      in.ContentType,
		Size:          actualSize,
		Checksum:      checksum,
		SourceService: in.SourceService,
		SourceModule:  in.SourceModule,
		UploadedBy:    in.UploadedBy,
		IsPublic:      in.IsPublic,
	}

	created, err := s.repo.Create(ctx, meta)
	if err != nil {
		// Best-effort cleanup
		_ = s.storage.Delete(ctx, bucket, objectKey)
		err = fmt.Errorf("upload: save metadata: %w", err)
		return nil, err
	}

	created.URL, _ = s.GetURL(ctx, created)
	out = created
	return out, nil
}

// List returns a paginated list of media files.
func (s *mediaService) List(ctx context.Context, params model.ListParams) (*model.ListResult, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	if params.Pagination == "cursor" {
		items, hasMore, err := s.repo.ListByCursor(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("list cursor: %w", err)
		}
		for _, item := range items {
			item.URL, _ = s.GetURL(ctx, item)
		}

		nextCursor := ""
		if hasMore && len(items) > 0 {
			last := items[len(items)-1]
			nextCursor = encodeCursor(last.CreatedAt, last.ID)
		}

		return &model.ListResult{
			Items:      items,
			Limit:      limit,
			HasMore:    hasMore,
			NextCursor: nextCursor,
		}, nil
	}

	items, total, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	// Attach URL to each item
	for _, item := range items {
		item.URL, _ = s.GetURL(ctx, item)
	}

	return &model.ListResult{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		Limit:      limit,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}, nil
}

func (s *mediaService) GetFilterOptions(ctx context.Context) (*model.FilterOptions, error) {
	options, err := s.repo.ListFilterOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list filter options: %w", err)
	}
	return options, nil
}

func encodeCursor(createdAt time.Time, id string) string {
	raw := fmt.Sprintf("%s|%s", createdAt.UTC().Format(time.RFC3339Nano), id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// GetByID fetches a single media file with its URL.
func (s *mediaService) GetByID(ctx context.Context, id string) (*model.MediaFile, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	m.URL, _ = s.GetURL(ctx, m)
	return m, nil
}

// Delete soft-deletes a media file only if ref_count == 0.
func (s *mediaService) Delete(ctx context.Context, id string) error {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if m.RefCount > 0 {
		return ErrFileInUse
	}

	// Delete from object storage first, then DB
	if err = s.storage.Delete(ctx, m.Bucket, m.ObjectKey); err != nil {
		s.log.Warn("could not delete object from storage",
			zap.String("id", id), zap.Error(err))
	}

	return s.repo.SoftDelete(ctx, id)
}

// GetURL returns the appropriate URL for a media file.
// Public files get a direct URL; private files get a signed URL.
func (s *mediaService) GetURL(ctx context.Context, m *model.MediaFile) (string, error) {
	if m.IsPublic {
		return s.storage.GetPublicURL(m.Bucket, m.ObjectKey), nil
	}
	return s.storage.GetSignedURL(ctx, m.Bucket, m.ObjectKey, s.cfg.MinIO.SignedURLExpiry)
}

// CleanupOrphans removes unreferenced files older than 24 hours.
// Returns the number of files deleted.
func (s *mediaService) CleanupOrphans(ctx context.Context) (int, error) {
	orphans, err := s.repo.FindStaleOrphans(ctx, 24*time.Hour)
	if err != nil {
		return 0, fmt.Errorf("cleanup: find orphans: %w", err)
	}

	deleted := 0
	for _, m := range orphans {
		if err = s.storage.Delete(ctx, m.Bucket, m.ObjectKey); err != nil {
			s.log.Warn("cleanup: storage delete failed",
				zap.String("id", m.ID), zap.Error(err))
		}
		if err = s.repo.HardDelete(ctx, m.ID); err != nil {
			s.log.Warn("cleanup: db hard delete failed",
				zap.String("id", m.ID), zap.Error(err))
			continue
		}
		deleted++
	}

	s.log.Info("cleanup complete", zap.Int("deleted", deleted), zap.Int("found", len(orphans)))
	return deleted, nil
}
