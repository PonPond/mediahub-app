CREATE TABLE IF NOT EXISTS project_upload_logs (
  id BIGSERIAL PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES api_projects(id) ON DELETE CASCADE,
  media_id UUID NULL REFERENCES media_files(id) ON DELETE SET NULL,
  file_name TEXT NOT NULL,
  mime_type TEXT NOT NULL,
  size BIGINT NOT NULL DEFAULT 0,
  source_service TEXT,
  source_module TEXT,
  status TEXT NOT NULL CHECK (status IN ('success', 'failed')),
  error_message TEXT NOT NULL DEFAULT '',
  uploaded_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_project_upload_logs_project_created
  ON project_upload_logs(project_id, created_at DESC);
