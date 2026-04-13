package settings

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SettingHandler struct {
	svc *SettingService
}

func NewSettingHandler(svc *SettingService) *SettingHandler {
	return &SettingHandler{svc: svc}
}

func (h *SettingHandler) List(c *gin.Context) {
	settings, err := h.svc.ListSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *SettingHandler) Update(c *gin.Context) {
	var input struct {
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateSetting(input.Key, input.Value, input.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "setting updated"})
}
