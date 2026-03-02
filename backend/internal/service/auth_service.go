package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"media-cms/internal/config"
	"media-cms/internal/model"
	"media-cms/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrForbidden = errors.New("forbidden")
var ErrBadRequest = errors.New("bad request")

type AuthService interface {
	EnsureDefaultAdmin(ctx context.Context) error
	Login(ctx context.Context, username, password string) (*model.LoginResponse, error)
	CreateUser(ctx context.Context, actorRole string, in model.CreateUserRequest) (*model.UserSummary, error)
	ListUsers(ctx context.Context, actorRole string) ([]model.UserSummary, error)
	UpdateUser(ctx context.Context, actorRole, id string, in model.UpdateUserRequest) (*model.UserSummary, error)
	DeleteUser(ctx context.Context, actorRole, id string) error
	CreateProject(ctx context.Context, actorRole string, in model.CreateProjectRequest) (*model.CreateProjectResponse, error)
	ListProjects(ctx context.Context, actorRole string) ([]model.ProjectSummary, error)
	UpdateProject(ctx context.Context, actorRole, id string, in model.UpdateProjectRequest) (*model.ProjectSummary, error)
	DeleteProject(ctx context.Context, actorRole, id string) error
	ListProjectUploadLogs(ctx context.Context, actorRole, projectID string, limit int) ([]model.ProjectUploadLog, error)
	IssueProjectToken(ctx context.Context, clientID, clientSecret string) (*model.LoginResponse, error)
}

type authService struct {
	repo repository.AuthRepository
	cfg  *config.Config
	log  *zap.Logger
}

func NewAuthService(repo repository.AuthRepository, cfg *config.Config, log *zap.Logger) AuthService {
	return &authService{repo: repo, cfg: cfg, log: log}
}

func (s *authService) EnsureDefaultAdmin(ctx context.Context) error {
	if s.cfg.Auth.DefaultAdminUsername == "" || s.cfg.Auth.DefaultAdminPassword == "" {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.Auth.DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("auth ensure admin hash: %w", err)
	}
	if err := s.repo.CreateOrUpdateUser(
		ctx,
		s.cfg.Auth.DefaultAdminUsername,
		string(hash),
		"admin",
	); err != nil {
		return fmt.Errorf("auth ensure admin upsert: %w", err)
	}
	s.log.Info("default admin user ensured",
		zap.String("username", s.cfg.Auth.DefaultAdminUsername))
	return nil
}

func (s *authService) Login(ctx context.Context, username, password string) (*model.LoginResponse, error) {
	u, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	token, expiresIn, err := s.issueToken(tokenInput{
		Subject:   u.ID,
		TokenType: "user",
		Role:      u.Role,
		Username:  u.Username,
	})
	if err != nil {
		return nil, err
	}
	return &model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}, nil
}

func (s *authService) CreateProject(
	ctx context.Context,
	actorRole string,
	in model.CreateProjectRequest,
) (*model.CreateProjectResponse, error) {
	if actorRole != "admin" {
		return nil, ErrForbidden
	}
	clientID := "prj_" + uuid.NewString()
	clientSecret := "sec_" + uuid.NewString()
	hash, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth create project hash: %w", err)
	}
	policy := model.NormalizeProjectUploadPolicy(in.UploadPolicy)
	p, err := s.repo.CreateProject(ctx, in.Name, clientID, string(hash), in.Scopes, policy)
	if err != nil {
		return nil, err
	}
	return &model.CreateProjectResponse{
		ID:           p.ID,
		Name:         p.Name,
		ClientID:     p.ClientID,
		ClientSecret: clientSecret,
		Scopes:       p.Scopes,
		UploadPolicy: p.UploadPolicy,
	}, nil
}

func (s *authService) CreateUser(
	ctx context.Context,
	actorRole string,
	in model.CreateUserRequest,
) (*model.UserSummary, error) {
	if actorRole != "admin" {
		return nil, ErrForbidden
	}
	role := in.Role
	if role == "" {
		role = "editor"
	}
	if role != "admin" && role != "editor" && role != "viewer" {
		return nil, ErrBadRequest
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth create user hash: %w", err)
	}
	u, err := s.repo.CreateUser(ctx, in.Username, string(hash), role)
	if err != nil {
		return nil, err
	}
	return &model.UserSummary{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}, nil
}

func (s *authService) ListUsers(ctx context.Context, actorRole string) ([]model.UserSummary, error) {
	if actorRole != "admin" {
		return nil, ErrForbidden
	}
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]model.UserSummary, 0, len(users))
	for _, u := range users {
		out = append(out, model.UserSummary{
			ID:        u.ID,
			Username:  u.Username,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		})
	}
	return out, nil
}

func (s *authService) UpdateUser(
	ctx context.Context,
	actorRole, id string,
	in model.UpdateUserRequest,
) (*model.UserSummary, error) {
	if actorRole != "admin" {
		return nil, ErrForbidden
	}
	if in.Role != "" && in.Role != "admin" && in.Role != "editor" && in.Role != "viewer" {
		return nil, ErrBadRequest
	}
	var passHash *string
	if in.Password != "" {
		if len(in.Password) < 8 {
			return nil, ErrBadRequest
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("auth update user hash: %w", err)
		}
		s := string(hash)
		passHash = &s
	}
	u, err := s.repo.UpdateUser(ctx, id, in.Role, passHash)
	if err != nil {
		return nil, err
	}
	return &model.UserSummary{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}, nil
}

func (s *authService) DeleteUser(ctx context.Context, actorRole, id string) error {
	if actorRole != "admin" {
		return ErrForbidden
	}
	return s.repo.DeleteUser(ctx, id)
}

func (s *authService) ListProjects(ctx context.Context, actorRole string) ([]model.ProjectSummary, error) {
	if actorRole != "admin" {
		return nil, ErrForbidden
	}
	projects, err := s.repo.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]model.ProjectSummary, 0, len(projects))
	for _, p := range projects {
		out = append(out, model.ProjectSummary{
			ID:           p.ID,
			Name:         p.Name,
			ClientID:     p.ClientID,
			Scopes:       p.Scopes,
			UploadPolicy: p.UploadPolicy,
			IsActive:     p.IsActive,
			CreatedAt:    p.CreatedAt,
		})
	}
	return out, nil
}

func (s *authService) UpdateProject(
	ctx context.Context,
	actorRole, id string,
	in model.UpdateProjectRequest,
) (*model.ProjectSummary, error) {
	if actorRole != "admin" {
		return nil, ErrForbidden
	}
	hasScopes := in.Scopes != nil
	hasPolicy := in.UploadPolicy != nil
	p, err := s.repo.UpdateProject(ctx, id, in.Name, in.Scopes, hasScopes, in.UploadPolicy, hasPolicy, in.IsActive)
	if err != nil {
		return nil, err
	}
	return &model.ProjectSummary{
		ID:           p.ID,
		Name:         p.Name,
		ClientID:     p.ClientID,
		Scopes:       p.Scopes,
		UploadPolicy: p.UploadPolicy,
		IsActive:     p.IsActive,
		CreatedAt:    p.CreatedAt,
	}, nil
}

func (s *authService) DeleteProject(ctx context.Context, actorRole, id string) error {
	if actorRole != "admin" {
		return ErrForbidden
	}
	return s.repo.DeleteProject(ctx, id)
}

func (s *authService) ListProjectUploadLogs(ctx context.Context, actorRole, projectID string, limit int) ([]model.ProjectUploadLog, error) {
	if actorRole != "admin" {
		return nil, ErrForbidden
	}
	return s.repo.ListProjectUploadLogs(ctx, projectID, limit)
}

func (s *authService) IssueProjectToken(
	ctx context.Context,
	clientID, clientSecret string,
) (*model.LoginResponse, error) {
	p, err := s.repo.GetProjectByClientID(ctx, clientID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(p.ClientSecretHash), []byte(clientSecret)) != nil {
		return nil, ErrInvalidCredentials
	}
	token, expiresIn, err := s.issueToken(tokenInput{
		Subject:   p.ID,
		TokenType: "project",
		Role:      "project",
		ProjectID: p.ID,
		Scopes:    p.Scopes,
	})
	if err != nil {
		return nil, err
	}
	return &model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}, nil
}

type tokenInput struct {
	Subject   string
	TokenType string
	Role      string
	Username  string
	ProjectID string
	Scopes    []string
}

func (s *authService) issueToken(in tokenInput) (string, int64, error) {
	now := time.Now()
	exp := now.Add(s.cfg.JWT.Expiry)
	claims := jwt.MapClaims{
		"sub":        in.Subject,
		"token_type": in.TokenType,
		"role":       in.Role,
		"username":   in.Username,
		"project_id": in.ProjectID,
		"scopes":     in.Scopes,
		"iat":        now.Unix(),
		"exp":        exp.Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", 0, fmt.Errorf("auth sign token: %w", err)
	}
	return signed, int64(s.cfg.JWT.Expiry.Seconds()), nil
}
