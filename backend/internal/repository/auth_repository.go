package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"media-cms/internal/model"
)

type AuthRepository interface {
	EnsureProjectPolicySchema(ctx context.Context) error
	EnsureUploadLogSchema(ctx context.Context) error
	CreateOrUpdateUser(ctx context.Context, username, passwordHash, role string) error
	CreateUser(ctx context.Context, username, passwordHash, role string) (*model.CMSUser, error)
	ListUsers(ctx context.Context) ([]model.CMSUser, error)
	UpdateUser(ctx context.Context, id, role string, passwordHash *string) (*model.CMSUser, error)
	DeleteUser(ctx context.Context, id string) error
	GetUserByUsername(ctx context.Context, username string) (*model.CMSUser, error)
	CreateProject(ctx context.Context, name, clientID, secretHash string, scopes []string, policy model.ProjectUploadPolicy) (*model.APIProject, error)
	ListProjects(ctx context.Context) ([]model.APIProject, error)
	UpdateProject(ctx context.Context, id, name string, scopes []string, hasScopes bool, policy *model.ProjectUploadPolicy, hasPolicy bool, isActive *bool) (*model.APIProject, error)
	DeleteProject(ctx context.Context, id string) error
	GetProjectByClientID(ctx context.Context, clientID string) (*model.APIProject, error)
	GetProjectByID(ctx context.Context, id string) (*model.APIProject, error)
	CreateProjectUploadLog(ctx context.Context, in model.ProjectUploadLog) error
	ListProjectUploadLogs(ctx context.Context, projectID string, limit int) ([]model.ProjectUploadLog, error)
}

type authRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) EnsureProjectPolicySchema(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
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
		}'::jsonb
	`)
	if err != nil {
		return fmt.Errorf("auth repo ensure project policy schema: %w", err)
	}
	return nil
}

func (r *authRepository) EnsureUploadLogSchema(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
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
	`)
	if err != nil {
		return fmt.Errorf("auth repo ensure upload log schema: %w", err)
	}
	return nil
}

func (r *authRepository) CreateOrUpdateUser(ctx context.Context, username, passwordHash, role string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cms_users (username, password_hash, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (username)
		DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role
	`, strings.ToLower(strings.TrimSpace(username)), passwordHash, role)
	if err != nil {
		return fmt.Errorf("auth repo upsert user: %w", err)
	}
	return nil
}

func (r *authRepository) GetUserByUsername(ctx context.Context, username string) (*model.CMSUser, error) {
	var user model.CMSUser
	err := r.db.GetContext(ctx, &user,
		`SELECT * FROM cms_users WHERE username = $1`, strings.ToLower(strings.TrimSpace(username)))
	if err != nil {
		return nil, fmt.Errorf("auth repo get user: %w", err)
	}
	return &user, nil
}

func (r *authRepository) CreateUser(
	ctx context.Context,
	username, passwordHash, role string,
) (*model.CMSUser, error) {
	var u model.CMSUser
	err := r.db.GetContext(ctx, &u, `
		INSERT INTO cms_users (username, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING *`,
		strings.ToLower(strings.TrimSpace(username)),
		passwordHash,
		role,
	)
	if err != nil {
		return nil, fmt.Errorf("auth repo create user: %w", err)
	}
	return &u, nil
}

func (r *authRepository) ListUsers(ctx context.Context) ([]model.CMSUser, error) {
	var users []model.CMSUser
	if err := r.db.SelectContext(ctx, &users,
		`SELECT id, username, role, created_at FROM cms_users ORDER BY created_at DESC`); err != nil {
		return nil, fmt.Errorf("auth repo list users: %w", err)
	}
	if users == nil {
		users = make([]model.CMSUser, 0)
	}
	return users, nil
}

func (r *authRepository) UpdateUser(
	ctx context.Context,
	id, role string,
	passwordHash *string,
) (*model.CMSUser, error) {
	var u model.CMSUser
	err := r.db.GetContext(ctx, &u, `
		UPDATE cms_users
		SET role = COALESCE(NULLIF($2, ''), role),
			password_hash = COALESCE($3, password_hash)
		WHERE id = $1
		RETURNING *`,
		id, role, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("auth repo update user: %w", err)
	}
	return &u, nil
}

func (r *authRepository) DeleteUser(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM cms_users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("auth repo delete user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *authRepository) CreateProject(
	ctx context.Context,
	name, clientID, secretHash string,
	scopes []string,
	policy model.ProjectUploadPolicy,
) (*model.APIProject, error) {
	if clientID == "" {
		clientID = "prj_" + uuid.NewString()
	}
	if len(scopes) == 0 {
		scopes = []string{"media:read", "media:write", "reference:write"}
	}

	var row struct {
		ID               string         `db:"id"`
		Name             string         `db:"name"`
		ClientID         string         `db:"client_id"`
		ClientSecretHash string         `db:"client_secret_hash"`
		Scopes           pq.StringArray `db:"scopes"`
		UploadPolicy     []byte         `db:"upload_policy"`
		IsActive         bool           `db:"is_active"`
		CreatedAt        time.Time      `db:"created_at"`
	}
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("auth repo create project marshal policy: %w", err)
	}
	err = r.db.GetContext(ctx, &row, `
		INSERT INTO api_projects (name, client_id, client_secret_hash, scopes, upload_policy, is_active)
		VALUES ($1, $2, $3, $4, $5::jsonb, true)
		RETURNING *`,
		name, clientID, secretHash, pq.StringArray(scopes), string(policyJSON))
	if err != nil {
		return nil, fmt.Errorf("auth repo create project: %w", err)
	}
	uploadPolicy, err := parseUploadPolicy(row.UploadPolicy)
	if err != nil {
		return nil, err
	}
	return &model.APIProject{
		ID:               row.ID,
		Name:             row.Name,
		ClientID:         row.ClientID,
		ClientSecretHash: row.ClientSecretHash,
		Scopes:           []string(row.Scopes),
		UploadPolicy:     uploadPolicy,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt,
	}, nil
}

func (r *authRepository) GetProjectByClientID(ctx context.Context, clientID string) (*model.APIProject, error) {
	var row struct {
		ID               string         `db:"id"`
		Name             string         `db:"name"`
		ClientID         string         `db:"client_id"`
		ClientSecretHash string         `db:"client_secret_hash"`
		Scopes           pq.StringArray `db:"scopes"`
		UploadPolicy     []byte         `db:"upload_policy"`
		IsActive         bool           `db:"is_active"`
		CreatedAt        time.Time      `db:"created_at"`
	}
	err := r.db.GetContext(ctx, &row,
		`SELECT * FROM api_projects WHERE client_id = $1 AND is_active = true`, clientID)
	if err != nil {
		return nil, fmt.Errorf("auth repo get project: %w", err)
	}
	uploadPolicy, err := parseUploadPolicy(row.UploadPolicy)
	if err != nil {
		return nil, err
	}
	return &model.APIProject{
		ID:               row.ID,
		Name:             row.Name,
		ClientID:         row.ClientID,
		ClientSecretHash: row.ClientSecretHash,
		Scopes:           []string(row.Scopes),
		UploadPolicy:     uploadPolicy,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt,
	}, nil
}

func (r *authRepository) GetProjectByID(ctx context.Context, id string) (*model.APIProject, error) {
	var row struct {
		ID               string         `db:"id"`
		Name             string         `db:"name"`
		ClientID         string         `db:"client_id"`
		ClientSecretHash string         `db:"client_secret_hash"`
		Scopes           pq.StringArray `db:"scopes"`
		UploadPolicy     []byte         `db:"upload_policy"`
		IsActive         bool           `db:"is_active"`
		CreatedAt        time.Time      `db:"created_at"`
	}
	err := r.db.GetContext(ctx, &row, `SELECT * FROM api_projects WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("auth repo get project by id: %w", err)
	}
	uploadPolicy, err := parseUploadPolicy(row.UploadPolicy)
	if err != nil {
		return nil, err
	}
	return &model.APIProject{
		ID:               row.ID,
		Name:             row.Name,
		ClientID:         row.ClientID,
		ClientSecretHash: row.ClientSecretHash,
		Scopes:           []string(row.Scopes),
		UploadPolicy:     uploadPolicy,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt,
	}, nil
}

func (r *authRepository) ListProjects(ctx context.Context) ([]model.APIProject, error) {
	rows := []struct {
		ID               string         `db:"id"`
		Name             string         `db:"name"`
		ClientID         string         `db:"client_id"`
		ClientSecretHash string         `db:"client_secret_hash"`
		Scopes           pq.StringArray `db:"scopes"`
		UploadPolicy     []byte         `db:"upload_policy"`
		IsActive         bool           `db:"is_active"`
		CreatedAt        time.Time      `db:"created_at"`
	}{}
	if err := r.db.SelectContext(ctx, &rows,
		`SELECT * FROM api_projects ORDER BY created_at DESC`); err != nil {
		return nil, fmt.Errorf("auth repo list projects: %w", err)
	}
	items := make([]model.APIProject, 0, len(rows))
	for _, row := range rows {
		uploadPolicy, err := parseUploadPolicy(row.UploadPolicy)
		if err != nil {
			return nil, err
		}
		items = append(items, model.APIProject{
			ID:               row.ID,
			Name:             row.Name,
			ClientID:         row.ClientID,
			ClientSecretHash: row.ClientSecretHash,
			Scopes:           []string(row.Scopes),
			UploadPolicy:     uploadPolicy,
			IsActive:         row.IsActive,
			CreatedAt:        row.CreatedAt,
		})
	}
	return items, nil
}

func (r *authRepository) UpdateProject(
	ctx context.Context,
	id, name string,
	scopes []string,
	hasScopes bool,
	policy *model.ProjectUploadPolicy,
	hasPolicy bool,
	isActive *bool,
) (*model.APIProject, error) {
	var scopesArg interface{}
	if hasScopes {
		scopesArg = pq.StringArray(scopes)
	} else {
		scopesArg = nil
	}
	var policyArg interface{}
	if hasPolicy {
		normalized := model.NormalizeProjectUploadPolicy(policy)
		policyJSON, err := json.Marshal(normalized)
		if err != nil {
			return nil, fmt.Errorf("auth repo update project marshal policy: %w", err)
		}
		policyArg = string(policyJSON)
	} else {
		policyArg = nil
	}
	var row struct {
		ID               string         `db:"id"`
		Name             string         `db:"name"`
		ClientID         string         `db:"client_id"`
		ClientSecretHash string         `db:"client_secret_hash"`
		Scopes           pq.StringArray `db:"scopes"`
		UploadPolicy     []byte         `db:"upload_policy"`
		IsActive         bool           `db:"is_active"`
		CreatedAt        time.Time      `db:"created_at"`
	}
	err := r.db.GetContext(ctx, &row, `
		UPDATE api_projects
		SET name = COALESCE(NULLIF($2, ''), name),
			scopes = COALESCE($3, scopes),
			upload_policy = COALESCE($4::jsonb, upload_policy),
			is_active = COALESCE($5, is_active)
		WHERE id = $1
		RETURNING *`,
		id, name, scopesArg, policyArg, isActive)
	if err != nil {
		return nil, fmt.Errorf("auth repo update project: %w", err)
	}
	uploadPolicy, err := parseUploadPolicy(row.UploadPolicy)
	if err != nil {
		return nil, err
	}
	return &model.APIProject{
		ID:               row.ID,
		Name:             row.Name,
		ClientID:         row.ClientID,
		ClientSecretHash: row.ClientSecretHash,
		Scopes:           []string(row.Scopes),
		UploadPolicy:     uploadPolicy,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt,
	}, nil
}

func (r *authRepository) DeleteProject(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM api_projects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("auth repo delete project: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("project not found")
	}
	return nil
}

func (r *authRepository) CreateProjectUploadLog(ctx context.Context, in model.ProjectUploadLog) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO project_upload_logs
		(project_id, media_id, file_name, mime_type, size, source_service, source_module, status, error_message, uploaded_by)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8, $9, $10)
	`,
		in.ProjectID,
		nullableString(in.MediaID),
		in.FileName,
		in.MimeType,
		in.Size,
		in.SourceService,
		in.SourceModule,
		in.Status,
		in.ErrorMessage,
		in.UploadedBy,
	)
	if err != nil {
		return fmt.Errorf("auth repo create project upload log: %w", err)
	}
	return nil
}

func (r *authRepository) ListProjectUploadLogs(ctx context.Context, projectID string, limit int) ([]model.ProjectUploadLog, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	items := make([]model.ProjectUploadLog, 0)
	err := r.db.SelectContext(ctx, &items, `
		SELECT
			id,
			project_id,
			media_id,
			file_name,
			mime_type,
			size,
			COALESCE(source_service, '') AS source_service,
			COALESCE(source_module, '') AS source_module,
			status,
			COALESCE(error_message, '') AS error_message,
			uploaded_by,
			created_at
		FROM project_upload_logs
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("auth repo list project upload logs: %w", err)
	}
	return items, nil
}

func nullableString(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func parseUploadPolicy(raw []byte) (model.ProjectUploadPolicy, error) {
	if len(raw) == 0 {
		p := model.DefaultProjectUploadPolicy()
		return p, nil
	}
	var p model.ProjectUploadPolicy
	if err := json.Unmarshal(raw, &p); err != nil {
		return model.ProjectUploadPolicy{}, fmt.Errorf("auth repo parse upload policy: %w", err)
	}
	normalized := model.NormalizeProjectUploadPolicy(&p)
	return normalized, nil
}
