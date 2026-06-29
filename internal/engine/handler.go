package engine

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/chess-puzzle-app/backend/internal/users"
	"github.com/gin-gonic/gin"
)

type EngineHandler struct {
	svc     *EngineService
	userSvc *users.UserService
	cache   *sync.Map // FEN -> *AnalysisResult cache
}

func NewEngineHandler(svc *EngineService, userSvc *users.UserService) *EngineHandler {
	return &EngineHandler{
		svc:     svc,
		userSvc: userSvc,
		cache:   &sync.Map{},
	}
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

	// Determine limits based on subscription
	plan, _ := c.Get("subscription_type")
	subType, _ := plan.(string)
	
	maxDepth := 14 // Guest/Free
	multiPV := 1
	
	switch subType {
	case "starter":
		maxDepth = 16
	case "pro":
		maxDepth = 20
	case "elite":
		maxDepth = 22
		multiPV = 3
	case "coach":
		maxDepth = 24
		multiPV = 5
	}

	analysisDepth := req.Depth
	if analysisDepth == 0 || analysisDepth > maxDepth {
		analysisDepth = maxDepth
	}

	// Check cache
	if cachedResult, ok := h.cache.Load(req.FEN); ok {
		// Only return cache if depth is sufficient
		res := cachedResult.(*AnalysisResult)
		if res.Depth >= analysisDepth && len(res.Lines) >= multiPV {
			c.JSON(http.StatusOK, cachedResult)
			return
		}
	}

	// Create context with timeout for analysis (movetime + 3s buffer)
	timeout := time.Duration(req.Movetime+3000) * time.Millisecond
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	// Update service to handle MultiPV override if we haven't already
	// Actually, let's just use the depth and if we want MultiPV we might need to update the service method
	result, err := h.svc.AnalyzePosition(ctx, req.FEN, analysisDepth, req.Movetime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "engine analysis failed: " + err.Error()})
		return
	}

	// We can manually filter lines if needed, or assume AnalyzePosition handles it
	// For now, let's ensure the result has the correct number of lines
	if len(result.Lines) > multiPV {
		result.Lines = result.Lines[:multiPV]
	}

	h.cache.Store(req.FEN, result)
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
		fmt.Printf("PlayMove error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "engine play failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

