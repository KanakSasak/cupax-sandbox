package api

import (
	"net/http"
	"time"

	"github.com/cupax/cupax/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HandleCreateWhitelist handles creation of a new whitelist entry
func (h *Handler) HandleCreateWhitelist(c *gin.Context) {
	var req models.WhitelistCreate

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validate whitelist type
	if req.Type != models.WhitelistTypeProcess &&
		req.Type != models.WhitelistTypeDomain &&
		req.Type != models.WhitelistTypeIP &&
		req.Type != models.WhitelistTypeRegistry {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid whitelist type"})
		return
	}

	whitelist := &models.Whitelist{
		ID:          uuid.New(),
		Type:        req.Type,
		Value:       req.Value,
		Description: req.Description,
		IsRegex:     req.IsRegex,
		Enabled:     req.Enabled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Default enabled to true if not specified
	if !req.Enabled {
		whitelist.Enabled = true
	}

	if err := h.repo.CreateWhitelist(whitelist); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create whitelist entry"})
		return
	}

	c.JSON(http.StatusCreated, whitelist)
}

// HandleGetWhitelists retrieves all whitelist entries or filtered by type
func (h *Handler) HandleGetWhitelists(c *gin.Context) {
	whitelistType := c.Query("type")
	enabledOnly := c.Query("enabled") == "true"

	var whitelists []models.Whitelist
	var err error

	if enabledOnly {
		whitelists, err = h.repo.GetEnabledWhitelists()
	} else if whitelistType != "" {
		whitelists, err = h.repo.GetWhitelistsByType(models.WhitelistType(whitelistType))
	} else {
		whitelists, err = h.repo.GetAllWhitelists()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve whitelists"})
		return
	}

	// Return empty array instead of null
	if whitelists == nil {
		whitelists = []models.Whitelist{}
	}

	c.JSON(http.StatusOK, whitelists)
}

// HandleGetWhitelistByID retrieves a specific whitelist entry
func (h *Handler) HandleGetWhitelistByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid whitelist ID"})
		return
	}

	whitelist, err := h.repo.GetWhitelistByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Whitelist not found"})
		return
	}

	c.JSON(http.StatusOK, whitelist)
}

// HandleUpdateWhitelist updates a whitelist entry
func (h *Handler) HandleUpdateWhitelist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid whitelist ID"})
		return
	}

	var update models.WhitelistUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.repo.UpdateWhitelist(id, &update); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update whitelist"})
		return
	}

	// Fetch updated whitelist to return
	whitelist, err := h.repo.GetWhitelistByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Whitelist updated successfully"})
		return
	}

	c.JSON(http.StatusOK, whitelist)
}

// HandleDeleteWhitelist deletes a whitelist entry
func (h *Handler) HandleDeleteWhitelist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid whitelist ID"})
		return
	}

	if err := h.repo.DeleteWhitelist(id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Whitelist not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Whitelist deleted successfully"})
}

// HandleBulkCreateWhitelists creates multiple whitelist entries at once
func (h *Handler) HandleBulkCreateWhitelists(c *gin.Context) {
	var requests []models.WhitelistCreate

	if err := c.ShouldBindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var created []models.Whitelist
	var errors []string

	for _, req := range requests {
		whitelist := &models.Whitelist{
			ID:          uuid.New(),
			Type:        req.Type,
			Value:       req.Value,
			Description: req.Description,
			IsRegex:     req.IsRegex,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := h.repo.CreateWhitelist(whitelist); err != nil {
			errors = append(errors, "Failed to create "+req.Value+": "+err.Error())
		} else {
			created = append(created, *whitelist)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"created": created,
		"errors":  errors,
	})
}
