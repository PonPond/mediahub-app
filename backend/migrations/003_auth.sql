-- ============================================================
-- MediaHub – Authentication schema
-- ============================================================

CREATE TABLE IF NOT EXISTS cms_users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    username      VARCHAR(120) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(50)  NOT NULL DEFAULT 'admin',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_projects (
    id                 UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name               VARCHAR(200) NOT NULL,
    client_id          VARCHAR(120) NOT NULL UNIQUE,
    client_secret_hash VARCHAR(255) NOT NULL,
    scopes             TEXT[]       NOT NULL DEFAULT '{}',
    is_active          BOOLEAN      NOT NULL DEFAULT true,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_projects_active ON api_projects (is_active);
