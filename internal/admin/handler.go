package admin

import (
	"net/http"
	"strconv"

	"github.com/chess-puzzle-app/backend/internal/engine"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	svc       *AdminService
	engineSvc *engine.EngineService
}

func NewAdminHandler(svc *AdminService, engineSvc *engine.EngineService) *AdminHandler {
	return &AdminHandler{svc: svc, engineSvc: engineSvc}
}

func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	users, total, err := h.svc.ListUsers(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *AdminHandler) GetEngineStatus(c *gin.Context) {
	if h.engineSvc == nil {
		c.JSON(http.StatusOK, gin.H{
			"active_workers": 0,
			"total_workers":  0,
			"is_responsive":  false,
			"error":          "engine service not initialized",
		})
		return
	}

	status := h.engineSvc.GetStatus()
	c.JSON(http.StatusOK, status)
}
