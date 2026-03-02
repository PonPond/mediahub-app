package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"media-cms/internal/model"
	"media-cms/internal/repository"
)

// ReferenceService manages file reference tracking.
type ReferenceService interface {
	Add(ctx context.Context, in *model.AddReferenceInput) error
	Remove(ctx context.Context, in *model.RemoveReferenceInput) error
	GetUsage(ctx context.Context, mediaID string) (*model.UsageResult, error)
}

type referenceService struct {
	refRepo   repository.ReferenceRepository
	mediaRepo repository.MediaRepository
	log       *zap.Logger
}

// NewReferenceService returns a wired ReferenceService.
func NewReferenceService(
	refRepo repository.ReferenceRepository,
	mediaRepo repository.MediaRepository,
	log *zap.Logger,
) ReferenceService {
	return &referenceService{
		refRepo:   refRepo,
		mediaRepo: mediaRepo,
		log:       log,
	}
}

// Add creates a new reference, triggering ref_count +1 via DB trigger.
func (s *referenceService) Add(ctx context.Context, in *model.AddReferenceInput) error {
	// Verify the media file exists
	if _, err := s.mediaRepo.GetByID(ctx, in.MediaID); err != nil {
		return fmt.Errorf("reference add: media %s not found", in.MediaID)
	}

	ref := &model.MediaReference{
		MediaID:    in.MediaID,
		RefService: in.RefService,
		RefTable:   in.RefTable,
		RefID:      in.RefID,
		RefField:   in.RefField,
	}

	if err := s.refRepo.Add(ctx, ref); err != nil {
		return fmt.Errorf("reference add: %w", err)
	}

	s.log.Info("reference added",
		zap.String("media_id", in.MediaID),
		zap.String("ref_service", in.RefService),
		zap.String("ref_table", in.RefTable),
		zap.String("ref_id", in.RefID))
	return nil
}

// Remove deletes a reference, triggering ref_count -1 via DB trigger.
func (s *referenceService) Remove(ctx context.Context, in *model.RemoveReferenceInput) error {
	if err := s.refRepo.Remove(ctx, in); err != nil {
		return fmt.Errorf("reference remove: %w", err)
	}

	s.log.Info("reference removed",
		zap.String("media_id", in.MediaID),
		zap.String("ref_service", in.RefService))
	return nil
}

// GetUsage returns the ref_count and full list of references for a file.
func (s *referenceService) GetUsage(ctx context.Context, mediaID string) (*model.UsageResult, error) {
	// Verify the file exists
	m, err := s.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return nil, fmt.Errorf("usage: media %s not found", mediaID)
	}

	refs, err := s.refRepo.ListByMedia(ctx, mediaID)
	if err != nil {
		return nil, fmt.Errorf("usage: list refs: %w", err)
	}

	return &model.UsageResult{
		MediaID:    mediaID,
		RefCount:   m.RefCount,
		References: refs,
	}, nil
}
