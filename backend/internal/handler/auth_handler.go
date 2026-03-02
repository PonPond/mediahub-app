package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"media-cms/internal/model"
	"media-cms/internal/service"
)

type AuthHandler struct {
	svc service.AuthService
	log *zap.Logger
}

func NewAuthHandler(svc service.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var in model.LoginRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.svc.Login(c.Request.Context(), in.Username, in.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		h.log.Error("login failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *AuthHandler) CreateProject(c *gin.Context) {
	var in model.CreateProjectRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role := c.GetString("role")
	out, err := h.svc.CreateProject(c.Request.Context(), role, in)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}
		h.log.Error("create project failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create project failed"})
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *AuthHandler) ListProjects(c *gin.Context) {
	role := c.GetString("role")
	items, err := h.svc.ListProjects(c.Request.Context(), role)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}
		h.log.Error("list projects failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list projects failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *AuthHandler) UpdateProject(c *gin.Context) {
	var in model.UpdateProjectRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role := c.GetString("role")
	out, err := h.svc.UpdateProject(c.Request.Context(), role, c.Param("id"), in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
		case errors.Is(err, service.ErrBadRequest):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		default:
			h.log.Error("update project failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update project failed"})
		}
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *AuthHandler) DeleteProject(c *gin.Context) {
	role := c.GetString("role")
	if err := h.svc.DeleteProject(c.Request.Context(), role, c.Param("id")); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}
		h.log.Error("delete project failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete project failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *AuthHandler) ListProjectUploadLogs(c *gin.Context) {
	role := c.GetString("role")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, err := h.svc.ListProjectUploadLogs(c.Request.Context(), role, c.Param("id"), limit)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}
		h.log.Error("list project upload logs failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list project upload logs failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *AuthHandler) CreateUser(c *gin.Context) {
	var in model.CreateUserRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role := c.GetString("role")
	out, err := h.svc.CreateUser(c.Request.Context(), role, in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
		case errors.Is(err, service.ErrBadRequest):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		default:
			h.log.Error("create user failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
		}
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *AuthHandler) ListUsers(c *gin.Context) {
	role := c.GetString("role")
	items, err := h.svc.ListUsers(c.Request.Context(), role)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}
		h.log.Error("list users failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list users failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *AuthHandler) UpdateUser(c *gin.Context) {
	var in model.UpdateUserRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role := c.GetString("role")
	out, err := h.svc.UpdateUser(c.Request.Context(), role, c.Param("id"), in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
		case errors.Is(err, service.ErrBadRequest):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		default:
			h.log.Error("update user failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update user failed"})
		}
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *AuthHandler) DeleteUser(c *gin.Context) {
	role := c.GetString("role")
	if err := h.svc.DeleteUser(c.Request.Context(), role, c.Param("id")); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}
		h.log.Error("delete user failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete user failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *AuthHandler) IssueProjectToken(c *gin.Context) {
	var in model.ProjectTokenRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.svc.IssueProjectToken(c.Request.Context(), in.ClientID, in.ClientSecret)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		h.log.Error("issue project token failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "issue token failed"})
		return
	}
	c.JSON(http.StatusOK, out)
}
