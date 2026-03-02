ALTER TABLE api_projects
ADD COLUMN IF NOT EXISTS upload_policy JSONB NOT NULL DEFAULT '{
  "limits_mb": {
    "image": 10,
    "video": 200,
    "audio": 30,
    "document": 20,
    "archive": 50,
    "other": 5
  }
}'::jsonb;
