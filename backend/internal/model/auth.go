package model

import "time"

var UploadPolicyGroups = []string{
	"image",
	"video",
	"audio",
	"document",
	"archive",
	"other",
}

type ProjectUploadPolicy struct {
	LimitsMB map[string]int64 `json:"limits_mb"`
}

func DefaultProjectUploadPolicy() ProjectUploadPolicy {
	return ProjectUploadPolicy{
		LimitsMB: map[string]int64{
			"image":    10,
			"video":    200,
			"audio":    30,
			"document": 20,
			"archive":  50,
			"other":    5,
		},
	}
}

func NormalizeProjectUploadPolicy(in *ProjectUploadPolicy) ProjectUploadPolicy {
	out := DefaultProjectUploadPolicy()
	if in == nil || in.LimitsMB == nil {
		return out
	}
	for _, group := range UploadPolicyGroups {
		if v, ok := in.LimitsMB[group]; ok && v >= 0 {
			out.LimitsMB[group] = v
		}
	}
	return out
}

func (p ProjectUploadPolicy) LimitBytes(group string) int64 {
	if p.LimitsMB == nil {
		return 0
	}
	return p.LimitsMB[group] * 1024 * 1024
}

type CMSUser struct {
	ID           string    `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Role         string    `db:"role" json:"role"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type APIProject struct {
	ID               string   `db:"id" json:"id"`
	Name             string   `db:"name" json:"name"`
	ClientID         string   `db:"client_id" json:"client_id"`
	ClientSecretHash string   `db:"client_secret_hash" json:"-"`
	Scopes           []string `db:"scopes" json:"scopes"`
	UploadPolicy     ProjectUploadPolicy
	IsActive         bool      `db:"is_active" json:"is_active"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type CreateProjectRequest struct {
	Name         string               `json:"name" binding:"required"`
	Scopes       []string             `json:"scopes"`
	UploadPolicy *ProjectUploadPolicy `json:"upload_policy"`
}

type ProjectSummary struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	ClientID     string              `json:"client_id"`
	Scopes       []string            `json:"scopes"`
	UploadPolicy ProjectUploadPolicy `json:"upload_policy"`
	IsActive     bool                `json:"is_active"`
	CreatedAt    time.Time           `json:"created_at"`
}

type CreateProjectResponse struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	ClientID     string              `json:"client_id"`
	ClientSecret string              `json:"client_secret"`
	Scopes       []string            `json:"scopes"`
	UploadPolicy ProjectUploadPolicy `json:"upload_policy"`
}

type ProjectTokenRequest struct {
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret" binding:"required"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Role     string `json:"role"`
	Password string `json:"password"`
}

type UserSummary struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateProjectRequest struct {
	Name         string               `json:"name"`
	Scopes       []string             `json:"scopes"`
	UploadPolicy *ProjectUploadPolicy `json:"upload_policy"`
	IsActive     *bool                `json:"is_active"`
}

type ProjectUploadLog struct {
	ID            int64     `db:"id" json:"id"`
	ProjectID     string    `db:"project_id" json:"project_id"`
	MediaID       *string   `db:"media_id" json:"media_id,omitempty"`
	FileName      string    `db:"file_name" json:"file_name"`
	MimeType      string    `db:"mime_type" json:"mime_type"`
	Size          int64     `db:"size" json:"size"`
	SourceService string    `db:"source_service" json:"source_service,omitempty"`
	SourceModule  string    `db:"source_module" json:"source_module,omitempty"`
	Status        string    `db:"status" json:"status"`
	ErrorMessage  string    `db:"error_message" json:"error_message,omitempty"`
	UploadedBy    string    `db:"uploaded_by" json:"uploaded_by"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
