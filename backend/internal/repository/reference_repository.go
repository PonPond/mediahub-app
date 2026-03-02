package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"media-cms/internal/model"
)

// ReferenceRepository manages media_references rows.
type ReferenceRepository interface {
	Add(ctx context.Context, ref *model.MediaReference) error
	Remove(ctx context.Context, ref *model.RemoveReferenceInput) error
	ListByMedia(ctx context.Context, mediaID string) ([]model.MediaReference, error)
	CountByMedia(ctx context.Context, mediaID string) (int, error)
}

type referenceRepository struct {
	db *sqlx.DB
}

// NewReferenceRepository returns a PostgreSQL-backed ReferenceRepository.
func NewReferenceRepository(db *sqlx.DB) ReferenceRepository {
	return &referenceRepository{db: db}
}

// Add inserts a new reference; the DB trigger increments ref_count.
// Duplicate inserts are silently ignored (ON CONFLICT DO NOTHING).
func (r *referenceRepository) Add(ctx context.Context, ref *model.MediaReference) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO media_references (media_id, ref_service, ref_table, ref_id, ref_field)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT ON CONSTRAINT uq_media_ref DO NOTHING`,
		ref.MediaID, ref.RefService, ref.RefTable, ref.RefID, ref.RefField)
	if err != nil {
		return fmt.Errorf("reference repo add: %w", err)
	}
	return nil
}

// Remove deletes a specific reference; the DB trigger decrements ref_count.
func (r *referenceRepository) Remove(ctx context.Context, in *model.RemoveReferenceInput) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM media_references
		WHERE media_id = $1 AND ref_service = $2
		  AND ref_table = $3 AND ref_id = $4 AND ref_field = $5`,
		in.MediaID, in.RefService, in.RefTable, in.RefID, in.RefField)
	if err != nil {
		return fmt.Errorf("reference repo remove: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("reference not found")
	}
	return nil
}

// ListByMedia returns all references for a media file.
func (r *referenceRepository) ListByMedia(ctx context.Context, mediaID string) ([]model.MediaReference, error) {
	var refs []model.MediaReference
	err := r.db.SelectContext(ctx, &refs,
		`SELECT * FROM media_references WHERE media_id = $1 ORDER BY created_at DESC`, mediaID)
	if err != nil {
		return nil, fmt.Errorf("reference repo list: %w", err)
	}
	if refs == nil {
		refs = make([]model.MediaReference, 0)
	}
	return refs, nil
}

// CountByMedia returns the number of active references for a file.
func (r *referenceRepository) CountByMedia(ctx context.Context, mediaID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM media_references WHERE media_id = $1`, mediaID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("reference repo count: %w", err)
	}
	return count, nil
}
