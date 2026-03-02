package handler

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"media-cms/internal/model"
	"media-cms/internal/service"
)

// MediaHandler handles HTTP requests for media operations.
type MediaHandler struct {
	svc service.MediaService
	log *zap.Logger
}

// NewMediaHandler creates a new handler wired to the service.
func NewMediaHandler(svc service.MediaService, log *zap.Logger) *MediaHandler {
	return &MediaHandler{svc: svc, log: log}
}

// Upload handles POST /media/upload
// Streams multipart upload — no memory buffering of file data.
func (h *MediaHandler) Upload(c *gin.Context) {
	userID := c.GetString("user_id")
	uploadedBy := firstNonEmpty(c.GetString("username"), userID)
	tokenType := c.GetString("token_type")
	projectID := c.GetString("project_id")

	// Use MultipartReader for true streaming (avoids disk tmp files for large uploads)
	mr, err := c.Request.MultipartReader()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expected multipart/form-data"})
		return
	}

	// Collect form fields; file part is streamed directly
	var (
		sourceService = firstNonEmpty(
			c.Query("source_service"),
			c.GetHeader("X-Source-Service"),
		)
		sourceModule = firstNonEmpty(
			c.Query("source_module"),
			c.GetHeader("X-Source-Module"),
		)
		isPublic = parseBool(
			firstNonEmpty(
				c.Query("is_public"),
				c.GetHeader("X-Is-Public"),
			),
		)
	)

	var result *model.MediaFile

	for {
		part, partErr := mr.NextPart()
		if partErr == io.EOF {
			break
		}
		if partErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read multipart"})
			return
		}

		name := part.FormName()

		// Text fields — small so we read them into memory
		if part.FileName() == "" {
			switch name {
			case "source_service":
				b, _ := io.ReadAll(io.LimitReader(part, 512))
				sourceService = string(b)
			case "source_module":
				b, _ := io.ReadAll(io.LimitReader(part, 512))
				sourceModule = string(b)
			case "is_public":
				b, _ := io.ReadAll(io.LimitReader(part, 16))
				isPublic = string(b) == "true" || string(b) == "1"
			}
			continue
		}

		// File part — stream directly to MinIO
		if name == "file" {
			contentType := resolveContentType(part)
			result, err = h.svc.Upload(c.Request.Context(), service.UploadInput{
				Reader:        part,
				FileName:      part.FileName(),
				ContentType:   contentType,
				Size:          -1, // unknown until streamed
				SourceService: sourceService,
				SourceModule:  sourceModule,
				UploadedBy:    uploadedBy,
				IsPublic:      isPublic,
				TokenType:     tokenType,
				ProjectID:     projectID,
			})
			if err != nil {
				h.log.Error("upload failed", zap.Error(err), zap.String("user_id", userID))
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			break
		}
	}

	if result == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file part found in request"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func parseBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// resolveContentType picks the MIME from the part header, falling back to octet-stream.
func resolveContentType(part *multipart.Part) string {
	ct := part.Header.Get("Content-Type")
	if ct == "" {
		return "application/octet-stream"
	}
	return ct
}

// List handles GET /media
func (h *MediaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	params := model.ListParams{
		Page:          page,
		Limit:         limit,
		Pagination:    c.DefaultQuery("pagination", "offset"),
		Cursor:        c.Query("cursor"),
		MimeGroup:     c.Query("type"),
		Search:        c.Query("search"),
		UploadedBy:    c.Query("uploaded_by"),
		SourceService: c.Query("source_service"),
		SourceModule:  c.Query("source_module"),
		SortBy:        c.DefaultQuery("sort_by", "created_at"),
		SortDir:       c.DefaultQuery("sort_dir", "desc"),
	}

	result, err := h.svc.List(c.Request.Context(), params)
	if err != nil {
		h.log.Error("list failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list media"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// FilterOptions handles GET /media/filter-options
func (h *MediaHandler) FilterOptions(c *gin.Context) {
	options, err := h.svc.GetFilterOptions(c.Request.Context())
	if err != nil {
		h.log.Error("filter options failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load filter options"})
		return
	}
	c.JSON(http.StatusOK, options)
}

// GetByID handles GET /media/:id
func (h *MediaHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	m, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "media file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch media"})
		return
	}
	c.JSON(http.StatusOK, m)
}

// Delete handles DELETE /media/:id
func (h *MediaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFileInUse):
			c.JSON(http.StatusConflict, gin.H{"error": "file is still in use by other services"})
		case errors.Is(err, service.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "media file not found"})
		default:
			h.log.Error("delete failed", zap.Error(err), zap.String("id", id))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete media"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// HealthCheck handles GET /health
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
