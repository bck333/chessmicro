package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	svc *UserService
}

func NewUserHandler(svc *UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	user, err := h.svc.TrackLogin(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"rank": GetRank(user.XP),
	})
}
func (h *UserHandler) UseHint(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	canUse, err := h.svc.UseHint(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !canUse {
		c.JSON(http.StatusForbidden, gin.H{"error": "daily hint limit reached", "code": "LIMIT_REACHED"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hint usage tracked"})
}

func (h *UserHandler) AddXP(c *gin.Context) {
	var input struct {
		Amount int `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	if err := h.svc.AddXP(userID, input.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "XP updated successfully"})
}

func (h *UserHandler) SaveProgress(c *gin.Context) {
	var input struct {
		CategoryID   string `json:"category_id"`
		CategoryName string `json:"category_name" binding:"required"`
		ProgressType string `json:"progress_type" binding:"required"` // "learn" or "practice"
		StepNumber   int    `json:"step_number" binding:"required"`
		Status       string `json:"status" binding:"required"`        // "started" or "completed"
		XPEarned     int    `json:"xp_earned"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	
	var categoryID uuid.UUID
	if input.CategoryID != "" {
		parsed, err := uuid.Parse(input.CategoryID)
		if err == nil {
			categoryID = parsed
		}
	}

	progress, err := h.svc.SaveProgress(userID, categoryID, input.CategoryName, input.ProgressType, input.StepNumber, input.Status, input.XPEarned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add XP if earned
	if input.XPEarned > 0 {
		_ = h.svc.AddXP(userID, input.XPEarned)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Progress saved successfully", "progress": progress})
}

func (h *UserHandler) GetProgressList(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	progressList, err := h.svc.GetProgressList(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"progress": progressList})
}
