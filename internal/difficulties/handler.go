package difficulties

import (
	"net/http"

	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DifficultyHandler struct {
	svc *DifficultyService
}

func NewDifficultyHandler(svc *DifficultyService) *DifficultyHandler {
	return &DifficultyHandler{svc: svc}
}

func (h *DifficultyHandler) Create(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Level       int    `json:"level"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	diff := &database.Difficulty{
		Name:        input.Name,
		Level:       input.Level,
		Description: input.Description,
	}

	if err := h.svc.CreateDifficulty(diff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, diff)
}

func (h *DifficultyHandler) List(c *gin.Context) {
	diffs, err := h.svc.ListDifficulties()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, diffs)
}

func (h *DifficultyHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid difficulty id"})
		return
	}

	var input struct {
		Name        string `json:"name"`
		Level       *int   `json:"level"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	diff, err := h.svc.GetDifficulty(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "difficulty not found"})
		return
	}

	if input.Name != "" {
		diff.Name = input.Name
	}
	if input.Level != nil {
		diff.Level = *input.Level
	}
	if input.Description != "" {
		diff.Description = input.Description
	}

	if err := h.svc.UpdateDifficulty(diff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, diff)
}

func (h *DifficultyHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid difficulty id"})
		return
	}

	if err := h.svc.DeleteDifficulty(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
