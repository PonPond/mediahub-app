package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"media-cms/internal/model"
	"media-cms/internal/service"
)

// ReferenceHandler handles HTTP requests for file reference tracking.
type ReferenceHandler struct {
	svc service.ReferenceService
	log *zap.Logger
}

// NewReferenceHandler returns a wired ReferenceHandler.
func NewReferenceHandler(svc service.ReferenceService, log *zap.Logger) *ReferenceHandler {
	return &ReferenceHandler{svc: svc, log: log}
}

// AddReference handles POST /media/reference
func (h *ReferenceHandler) AddReference(c *gin.Context) {
	var in model.AddReferenceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Add(c.Request.Context(), &in); err != nil {
		h.log.Error("add reference failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "reference added"})
}

// RemoveReference handles DELETE /media/reference
func (h *ReferenceHandler) RemoveReference(c *gin.Context) {
	var in model.RemoveReferenceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Remove(c.Request.Context(), &in); err != nil {
		h.log.Error("remove reference failed", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reference removed"})
}

// GetUsage handles GET /media/:id/usage
func (h *ReferenceHandler) GetUsage(c *gin.Context) {
	mediaID := c.Param("id")
	usage, err := h.svc.GetUsage(c.Request.Context(), mediaID)
	if err != nil {
		h.log.Error("get usage failed", zap.Error(err), zap.String("media_id", mediaID))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}
