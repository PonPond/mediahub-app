-- ============================================================
-- MediaHub – Performance Indexes (scale preparation)
-- ============================================================

-- For ILIKE '%keyword%' filename search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_mf_file_name_trgm
    ON media_files USING gin (file_name gin_trgm_ops)
    WHERE deleted_at IS NULL;

-- For source filters (ILIKE)
CREATE INDEX IF NOT EXISTS idx_mf_source_service_lower
    ON media_files (lower(source_service))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_mf_source_module_lower
    ON media_files (lower(source_module))
    WHERE deleted_at IS NULL;

-- For keyset pagination ORDER BY created_at DESC, id DESC
CREATE INDEX IF NOT EXISTS idx_mf_created_id_desc
    ON media_files (created_at DESC, id DESC)
    WHERE deleted_at IS NULL;
