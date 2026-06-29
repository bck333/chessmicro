package learning

import (
	"net/http"

	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LearningHandler struct {
	svc *LearningService
}

func NewLearningHandler(svc *LearningService) *LearningHandler {
	return &LearningHandler{svc: svc}
}

// ============================================================================
// Learning Categories Endpoints
// ============================================================================

func (h *LearningHandler) ListCategories(c *gin.Context) {
	cats, err := h.svc.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

func (h *LearningHandler) GetCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	cat, err := h.svc.GetCategory(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}
	c.JSON(http.StatusOK, cat)
}

func (h *LearningHandler) CreateCategory(c *gin.Context) {
	var input struct {
		Title       string `json:"title" binding:"required"`
		Icon        string `json:"icon"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cat := &database.LearningCategory{
		ID:          uuid.New(),
		Title:       input.Title,
		Icon:        input.Icon,
		Description: input.Description,
	}

	if err := h.svc.CreateCategory(cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

func (h *LearningHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var input struct {
		Title       string `json:"title"`
		Icon        string `json:"icon"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cat, err := h.svc.GetCategory(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	if input.Title != "" {
		cat.Title = input.Title
	}
	if input.Icon != "" {
		cat.Icon = input.Icon
	}
	if input.Description != "" {
		cat.Description = input.Description
	}

	if err := h.svc.UpdateCategory(cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cat)
}

func (h *LearningHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	if err := h.svc.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ============================================================================
// Lessons Endpoints
// ============================================================================

func (h *LearningHandler) ListLessons(c *gin.Context) {
	catIDStr := c.Query("category_id")
	var catID *uuid.UUID
	if catIDStr != "" {
		parsed, err := uuid.Parse(catIDStr)
		if err == nil {
			catID = &parsed
		}
	}

	lessons, err := h.svc.ListLessons(catID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lessons)
}

func (h *LearningHandler) GetLesson(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	lesson, err := h.svc.GetLesson(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
	}
	c.JSON(http.StatusOK, lesson)
}

func (h *LearningHandler) CreateLesson(c *gin.Context) {
	var input struct {
		CategoryID uuid.UUID `json:"category_id" binding:"required"`
		Title      string    `json:"title" binding:"required"`
		Difficulty string    `json:"difficulty"`
		Thumbnail  string    `json:"thumbnail"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lesson := &database.Lesson{
		ID:         uuid.New(),
		CategoryID: input.CategoryID,
		Title:      input.Title,
		Difficulty: input.Difficulty,
		Thumbnail:  input.Thumbnail,
	}

	if err := h.svc.CreateLesson(lesson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, lesson)
}

func (h *LearningHandler) UpdateLesson(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	var input struct {
		Title      string `json:"title"`
		Difficulty string `json:"difficulty"`
		Thumbnail  string `json:"thumbnail"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lesson, err := h.svc.GetLesson(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
	}

	if input.Title != "" {
		lesson.Title = input.Title
	}
	if input.Difficulty != "" {
		lesson.Difficulty = input.Difficulty
	}
	if input.Thumbnail != "" {
		lesson.Thumbnail = input.Thumbnail
	}

	if err := h.svc.UpdateLesson(lesson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lesson)
}

func (h *LearningHandler) DeleteLesson(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	if err := h.svc.DeleteLesson(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ============================================================================
// Lesson Steps Endpoints
// ============================================================================

func (h *LearningHandler) ListSteps(c *gin.Context) {
	idStr := c.Param("id") // Lesson ID
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	steps, err := h.svc.ListSteps(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, steps)
}

func (h *LearningHandler) GetStep(c *gin.Context) {
	idStr := c.Param("id") // Step ID
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step id"})
		return
	}

	step, err := h.svc.GetStep(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "step not found"})
		return
	}
	c.JSON(http.StatusOK, step)
}

func (h *LearningHandler) CreateStep(c *gin.Context) {
	idStr := c.Param("id") // Lesson ID
	lessonID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	var input struct {
		StepOrder          int    `json:"step_order"`
		Title              string `json:"title" binding:"required"`
		Description        string `json:"description"`
		FEN                string `json:"fen"`
		HighlightedSquares string `json:"highlighted_squares"`
		MoveArrows         string `json:"move_arrows"`
		ImageURL           string `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	step := &database.LessonStep{
		ID:                 uuid.New(),
		LessonID:           lessonID,
		StepOrder:          input.StepOrder,
		Title:              input.Title,
		Description:        input.Description,
		FEN:                input.FEN,
		HighlightedSquares: input.HighlightedSquares,
		MoveArrows:         input.MoveArrows,
		ImageURL:           input.ImageURL,
	}

	if err := h.svc.CreateStep(step); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, step)
}

func (h *LearningHandler) UpdateStep(c *gin.Context) {
	idStr := c.Param("id") // Step ID
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step id"})
		return
	}

	var input struct {
		StepOrder          *int    `json:"step_order"`
		Title              string  `json:"title"`
		Description        string  `json:"description"`
		FEN                *string `json:"fen"`
		HighlightedSquares *string `json:"highlighted_squares"`
		MoveArrows         *string `json:"move_arrows"`
		ImageURL           *string `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	step, err := h.svc.GetStep(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "step not found"})
		return
	}

	if input.Title != "" {
		step.Title = input.Title
	}
	if input.Description != "" {
		step.Description = input.Description
	}
	if input.StepOrder != nil {
		step.StepOrder = *input.StepOrder
	}
	if input.FEN != nil {
		step.FEN = *input.FEN
	}
	if input.HighlightedSquares != nil {
		step.HighlightedSquares = *input.HighlightedSquares
	}
	if input.MoveArrows != nil {
		step.MoveArrows = *input.MoveArrows
	}
	if input.ImageURL != nil {
		step.ImageURL = *input.ImageURL
	}

	if err := h.svc.UpdateStep(step); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, step)
}

func (h *LearningHandler) DeleteStep(c *gin.Context) {
	idStr := c.Param("id") // Step ID
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step id"})
		return
	}

	if err := h.svc.DeleteStep(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *LearningHandler) ReorderSteps(c *gin.Context) {
	idStr := c.Param("id") // Lesson ID
	lessonID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	var input struct {
		StepIDs []uuid.UUID `json:"step_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ReorderSteps(lessonID, input.StepIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "reordered"})
}
