package repository

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"media-cms/internal/model"
)

const mediaSelectColumns = `
	id,
	bucket,
	object_key,
	file_name,
	mime_type,
	size,
	checksum,
	source_service,
	source_module,
	COALESCE(
		(
			SELECT cu.username
			FROM cms_users cu
			WHERE cu.id::text = media_files.uploaded_by OR cu.username = media_files.uploaded_by
			LIMIT 1
		),
		(
			SELECT ap.name
			FROM api_projects ap
			WHERE ap.id::text = media_files.uploaded_by
			LIMIT 1
		),
		media_files.uploaded_by
	) AS uploaded_by,
	is_public,
	ref_count,
	created_at,
	deleted_at
`

// MediaRepository defines database operations for media files.
type MediaRepository interface {
	Create(ctx context.Context, m *model.MediaFile) (*model.MediaFile, error)
	GetByID(ctx context.Context, id string) (*model.MediaFile, error)
	GetByChecksum(ctx context.Context, checksum string) (*model.MediaFile, error)
	List(ctx context.Context, params model.ListParams) ([]*model.MediaFile, int, error)
	ListByCursor(ctx context.Context, params model.ListParams) ([]*model.MediaFile, bool, error)
	ListFilterOptions(ctx context.Context) (*model.FilterOptions, error)
	SoftDelete(ctx context.Context, id string) error
	FindStaleOrphans(ctx context.Context, olderThan time.Duration) ([]*model.MediaFile, error)
	HardDelete(ctx context.Context, id string) error
}

type mediaRepository struct {
	db *sqlx.DB
}

// NewMediaRepository returns a PostgreSQL-backed MediaRepository.
func NewMediaRepository(db *sqlx.DB) MediaRepository {
	return &mediaRepository{db: db}
}

// Create inserts a new media_file row and returns the created record.
func (r *mediaRepository) Create(ctx context.Context, m *model.MediaFile) (*model.MediaFile, error) {
	query := `
		INSERT INTO media_files
			(bucket, object_key, file_name, mime_type, size, checksum,
			 source_service, source_module, uploaded_by, is_public)
		VALUES
			(:bucket, :object_key, :file_name, :mime_type, :size, :checksum,
			 :source_service, :source_module, :uploaded_by, :is_public)
		RETURNING *`

	rows, err := r.db.NamedQueryContext(ctx, query, m)
	if err != nil {
		return nil, fmt.Errorf("media repo create: %w", err)
	}
	defer rows.Close()

	var created model.MediaFile
	if rows.Next() {
		if err = rows.StructScan(&created); err != nil {
			return nil, fmt.Errorf("media repo create scan: %w", err)
		}
	}
	return &created, nil
}

// GetByID fetches a non-deleted media file by primary key.
func (r *mediaRepository) GetByID(ctx context.Context, id string) (*model.MediaFile, error) {
	var m model.MediaFile
	query := fmt.Sprintf(`SELECT %s FROM media_files WHERE id = $1 AND deleted_at IS NULL`, mediaSelectColumns)
	err := r.db.GetContext(ctx, &m, query, id)
	if err != nil {
		return nil, fmt.Errorf("media repo get: %w", err)
	}
	return &m, nil
}

// GetByChecksum returns a file with the same SHA-256 (dedup support).
func (r *mediaRepository) GetByChecksum(ctx context.Context, checksum string) (*model.MediaFile, error) {
	var m model.MediaFile
	query := fmt.Sprintf(`SELECT %s FROM media_files WHERE checksum = $1 AND deleted_at IS NULL LIMIT 1`, mediaSelectColumns)
	err := r.db.GetContext(ctx, &m, query, checksum)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// List returns paginated results with optional MIME-group + full-text search.
func (r *mediaRepository) List(ctx context.Context, params model.ListParams) ([]*model.MediaFile, int, error) {
	where := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	idx := 1

	if params.MimeGroup != "" {
		switch params.MimeGroup {
		case "image":
			where = append(where, fmt.Sprintf("mime_type LIKE $%d", idx))
			args = append(args, "image/%")
			idx++
		case "video":
			where = append(where, fmt.Sprintf("mime_type LIKE $%d", idx))
			args = append(args, "video/%")
			idx++
		case "audio":
			where = append(where, fmt.Sprintf("mime_type LIKE $%d", idx))
			args = append(args, "audio/%")
			idx++
		case "document":
			where = append(where,
				fmt.Sprintf("(mime_type = 'application/pdf' OR mime_type LIKE $%d OR mime_type LIKE $%d)",
					idx, idx+1))
			args = append(args, "text/%", "%document%")
			idx += 2
		case "other":
			where = append(where,
				"(mime_type NOT LIKE 'image/%' AND mime_type NOT LIKE 'video/%' AND mime_type NOT LIKE 'audio/%' "+
					"AND mime_type <> 'application/pdf' AND mime_type NOT LIKE 'text/%' "+
					"AND mime_type NOT LIKE '%document%' AND mime_type NOT LIKE '%spreadsheet%' AND mime_type NOT LIKE '%presentation%')")
		}
	}

	if params.Search != "" {
		// Use ILIKE for robust filename search across mixed languages/extensions.
		where = append(where, fmt.Sprintf("file_name ILIKE $%d", idx))
		args = append(args, "%"+params.Search+"%")
		idx++
	}

	if params.UploadedBy != "" {
		where = append(where, fmt.Sprintf(`(
			uploaded_by ILIKE $%d
			OR EXISTS (
				SELECT 1
				FROM cms_users cu
				WHERE (cu.id::text = media_files.uploaded_by OR cu.username = media_files.uploaded_by)
				  AND cu.username ILIKE $%d
			)
			OR EXISTS (
				SELECT 1
				FROM api_projects ap
				WHERE ap.id::text = media_files.uploaded_by
				  AND ap.name ILIKE $%d
			)
		)`, idx, idx, idx))
		args = append(args, "%"+params.UploadedBy+"%")
		idx++
	}

	if params.SourceService != "" {
		where = append(where, fmt.Sprintf("source_service ILIKE $%d", idx))
		args = append(args, "%"+params.SourceService+"%")
		idx++
	}

	if params.SourceModule != "" {
		where = append(where, fmt.Sprintf("source_module ILIKE $%d", idx))
		args = append(args, "%"+params.SourceModule+"%")
		idx++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	// Count total
	var total int
	countQ := "SELECT COUNT(*) FROM media_files " + whereClause
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("media repo list count: %w", err)
	}

	// Sort
	allowedSort := map[string]bool{"created_at": true, "size": true, "file_name": true}
	sortBy := "created_at"
	if allowedSort[params.SortBy] {
		sortBy = params.SortBy
	}
	sortDir := "DESC"
	if strings.ToUpper(params.SortDir) == "ASC" {
		sortDir = "ASC"
	}

	// Pagination
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	page := params.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	dataQ := fmt.Sprintf(`
		SELECT %s FROM media_files %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		mediaSelectColumns, whereClause, sortBy, sortDir, idx, idx+1)
	args = append(args, limit, offset)

	var items []*model.MediaFile
	if err := r.db.SelectContext(ctx, &items, dataQ, args...); err != nil {
		return nil, 0, fmt.Errorf("media repo list select: %w", err)
	}
	return items, total, nil
}

// ListByCursor returns cursor-based pages ordered by created_at DESC, id DESC.
// It avoids deep OFFSET scans for large datasets.
func (r *mediaRepository) ListByCursor(ctx context.Context, params model.ListParams) ([]*model.MediaFile, bool, error) {
	where := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	idx := 1

	if params.MimeGroup != "" {
		switch params.MimeGroup {
		case "image":
			where = append(where, fmt.Sprintf("mime_type LIKE $%d", idx))
			args = append(args, "image/%")
			idx++
		case "video":
			where = append(where, fmt.Sprintf("mime_type LIKE $%d", idx))
			args = append(args, "video/%")
			idx++
		case "audio":
			where = append(where, fmt.Sprintf("mime_type LIKE $%d", idx))
			args = append(args, "audio/%")
			idx++
		case "document":
			where = append(where,
				fmt.Sprintf("(mime_type = 'application/pdf' OR mime_type LIKE $%d OR mime_type LIKE $%d)",
					idx, idx+1))
			args = append(args, "text/%", "%document%")
			idx += 2
		case "other":
			where = append(where,
				"(mime_type NOT LIKE 'image/%' AND mime_type NOT LIKE 'video/%' AND mime_type NOT LIKE 'audio/%' "+
					"AND mime_type <> 'application/pdf' AND mime_type NOT LIKE 'text/%' "+
					"AND mime_type NOT LIKE '%document%' AND mime_type NOT LIKE '%spreadsheet%' AND mime_type NOT LIKE '%presentation%')")
		}
	}

	if params.Search != "" {
		where = append(where, fmt.Sprintf("file_name ILIKE $%d", idx))
		args = append(args, "%"+params.Search+"%")
		idx++
	}

	if params.UploadedBy != "" {
		where = append(where, fmt.Sprintf(`(
			uploaded_by ILIKE $%d
			OR EXISTS (
				SELECT 1
				FROM cms_users cu
				WHERE (cu.id::text = media_files.uploaded_by OR cu.username = media_files.uploaded_by)
				  AND cu.username ILIKE $%d
			)
			OR EXISTS (
				SELECT 1
				FROM api_projects ap
				WHERE ap.id::text = media_files.uploaded_by
				  AND ap.name ILIKE $%d
			)
		)`, idx, idx, idx))
		args = append(args, "%"+params.UploadedBy+"%")
		idx++
	}

	if params.SourceService != "" {
		where = append(where, fmt.Sprintf("source_service ILIKE $%d", idx))
		args = append(args, "%"+params.SourceService+"%")
		idx++
	}

	if params.SourceModule != "" {
		where = append(where, fmt.Sprintf("source_module ILIKE $%d", idx))
		args = append(args, "%"+params.SourceModule+"%")
		idx++
	}

	if params.Cursor != "" {
		cursorAt, cursorID, err := decodeCursor(params.Cursor)
		if err != nil {
			return nil, false, fmt.Errorf("media repo cursor decode: %w", err)
		}
		where = append(where, fmt.Sprintf("(created_at, id) < ($%d, $%d)", idx, idx+1))
		args = append(args, cursorAt, cursorID)
		idx += 2
	}

	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")
	dataQ := fmt.Sprintf(`
		SELECT %s FROM media_files %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d`, mediaSelectColumns, whereClause, idx)
	args = append(args, limit+1)

	var items []*model.MediaFile
	if err := r.db.SelectContext(ctx, &items, dataQ, args...); err != nil {
		return nil, false, fmt.Errorf("media repo cursor list select: %w", err)
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	return items, hasMore, nil
}

func decodeCursor(cursor string) (time.Time, string, error) {
	b, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid base64 cursor")
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid cursor format")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor timestamp")
	}
	if parts[1] == "" {
		return time.Time{}, "", fmt.Errorf("invalid cursor id")
	}
	return t, parts[1], nil
}

func (r *mediaRepository) ListFilterOptions(ctx context.Context) (*model.FilterOptions, error) {
	serviceQ := `
		SELECT DISTINCT source_service
		FROM media_files
		WHERE deleted_at IS NULL
		  AND source_service IS NOT NULL
		  AND btrim(source_service) <> ''
		ORDER BY source_service
	`
	moduleQ := `
		SELECT DISTINCT source_module
		FROM media_files
		WHERE deleted_at IS NULL
		  AND source_module IS NOT NULL
		  AND btrim(source_module) <> ''
		ORDER BY source_module
	`

	var services []string
	if err := r.db.SelectContext(ctx, &services, serviceQ); err != nil {
		return nil, fmt.Errorf("media repo list filter options services: %w", err)
	}

	var modules []string
	if err := r.db.SelectContext(ctx, &modules, moduleQ); err != nil {
		return nil, fmt.Errorf("media repo list filter options modules: %w", err)
	}

	return &model.FilterOptions{
		SourceServices: services,
		SourceModules:  modules,
	}, nil
}

// SoftDelete marks a record as deleted without removing the row.
func (r *mediaRepository) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE media_files SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("media repo soft delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("media file %s not found", id)
	}
	return nil
}

// FindStaleOrphans returns media files with ref_count = 0 older than olderThan.
func (r *mediaRepository) FindStaleOrphans(ctx context.Context, olderThan time.Duration) ([]*model.MediaFile, error) {
	cutoff := time.Now().Add(-olderThan)
	var items []*model.MediaFile
	err := r.db.SelectContext(ctx, &items,
		`SELECT * FROM media_files
		 WHERE ref_count = 0 AND created_at < $1 AND deleted_at IS NULL`,
		cutoff)
	if err != nil {
		return nil, fmt.Errorf("media repo find orphans: %w", err)
	}
	return items, nil
}

// HardDelete permanently removes a row (used by cleanup job).
func (r *mediaRepository) HardDelete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM media_files WHERE id = $1`, id)
	return err
}
