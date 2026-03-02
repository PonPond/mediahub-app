-- ============================================================
-- MediaHub – Initial Schema
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- -------------------------------------------------------
-- media_files: core metadata table
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS media_files (
    id             UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    bucket         VARCHAR(255) NOT NULL,
    object_key     VARCHAR(1000) NOT NULL,
    file_name      VARCHAR(500) NOT NULL,
    mime_type      VARCHAR(200) NOT NULL,
    size           BIGINT       NOT NULL,
    checksum       VARCHAR(64)  NOT NULL,          -- SHA-256 hex
    source_service VARCHAR(200),
    source_module  VARCHAR(200),
    uploaded_by    VARCHAR(200) NOT NULL,
    is_public      BOOLEAN      NOT NULL DEFAULT false,
    ref_count      INTEGER      NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_mf_created_at    ON media_files (created_at);
CREATE INDEX IF NOT EXISTS idx_mf_mime_type     ON media_files (mime_type);
CREATE INDEX IF NOT EXISTS idx_mf_uploaded_by   ON media_files (uploaded_by);
CREATE INDEX IF NOT EXISTS idx_mf_checksum      ON media_files (checksum);
CREATE INDEX IF NOT EXISTS idx_mf_deleted_at    ON media_files (deleted_at);
CREATE INDEX IF NOT EXISTS idx_mf_ref_count     ON media_files (ref_count);
CREATE INDEX IF NOT EXISTS idx_mf_is_public     ON media_files (is_public);
-- Full-text search on file_name
CREATE INDEX IF NOT EXISTS idx_mf_file_name     ON media_files USING gin (to_tsvector('english', file_name));

-- -------------------------------------------------------
-- media_references: tracks which service/record uses a file
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS media_references (
    id          BIGSERIAL    PRIMARY KEY,
    media_id    UUID         NOT NULL REFERENCES media_files (id),
    ref_service VARCHAR(200) NOT NULL,
    ref_table   VARCHAR(200) NOT NULL,
    ref_id      VARCHAR(200) NOT NULL,
    ref_field   VARCHAR(200) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_media_ref UNIQUE (media_id, ref_service, ref_table, ref_id, ref_field)
);

CREATE INDEX IF NOT EXISTS idx_mr_media_id    ON media_references (media_id);
CREATE INDEX IF NOT EXISTS idx_mr_ref_service ON media_references (ref_service);
CREATE INDEX IF NOT EXISTS idx_mr_ref_table   ON media_references (ref_table);
CREATE INDEX IF NOT EXISTS idx_mr_ref_id      ON media_references (ref_id);

-- -------------------------------------------------------
-- Trigger: keep ref_count in sync
-- -------------------------------------------------------
CREATE OR REPLACE FUNCTION trg_ref_count_inc()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    UPDATE media_files SET ref_count = ref_count + 1 WHERE id = NEW.media_id;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION trg_ref_count_dec()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    UPDATE media_files SET ref_count = GREATEST(0, ref_count - 1) WHERE id = OLD.media_id;
    RETURN OLD;
END;
$$;

DROP TRIGGER IF EXISTS trg_mr_inc ON media_references;
CREATE TRIGGER trg_mr_inc
    AFTER INSERT ON media_references
    FOR EACH ROW EXECUTE FUNCTION trg_ref_count_inc();

DROP TRIGGER IF EXISTS trg_mr_dec ON media_references;
CREATE TRIGGER trg_mr_dec
    AFTER DELETE ON media_references
    FOR EACH ROW EXECUTE FUNCTION trg_ref_count_dec();
