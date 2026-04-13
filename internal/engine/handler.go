package engine

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type EngineHandler struct {
	svc *EngineService
}

func NewEngineHandler(svc *EngineService) *EngineHandler {
	return &EngineHandler{svc: svc}
}

type AnalysisRequest struct {
	FEN      string `json:"fen" binding:"required"`
	Depth    int    `json:"depth"`
	Movetime int    `json:"movetime"` // milliseconds, defaults to 1000
}

func (h *EngineHandler) Analyze(c *gin.Context) {
	var req AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Movetime == 0 {
		req.Movetime = 1000
	}

	// Create context with timeout for analysis (movetime + 3s buffer)
	timeout := time.Duration(req.Movetime+3000) * time.Millisecond
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	result, err := h.svc.AnalyzePosition(ctx, req.FEN, req.Depth, req.Movetime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "engine analysis failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

type PlayRequest struct {
	FEN      string `json:"fen" binding:"required"`
	Level    int    `json:"level"`    // 1-6
	Elo      int    `json:"elo"`      // Manual elo
	Depth    int    `json:"depth"`    // Manual depth
	Skill    int    `json:"skill"`    // Manual skill
	Movetime int    `json:"movetime"` // milliseconds
}

func (h *EngineHandler) Play(c *gin.Context) {
	var req PlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Movetime == 0 {
		req.Movetime = 500
	}

	// Create context with timeout for play move (movetime + buffer)
	timeout := time.Duration(req.Movetime+5000) * time.Millisecond
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	opts := PlayOptions{
		Level:    req.Level,
		Elo:      req.Elo,
		Depth:    req.Depth,
		Skill:    req.Skill,
		Movetime: req.Movetime,
	}

	result, err := h.svc.PlayMove(ctx, req.FEN, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "engine play failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

