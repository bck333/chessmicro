package engine

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/chess-puzzle-app/backend/internal/users"
	"github.com/gin-gonic/gin"
)

// recordAnalyze updates observability counters for an analyze request.
func (h *EngineHandler) recordAnalyze(start time.Time, err error, timedOut bool) {
	atomic.AddInt64(&h.svc.metrics.AnalyzeRequests, 1)
	atomic.AddInt64(&h.svc.metrics.TotalLatencyMicro, time.Since(start).Microseconds())
	if timedOut {
		atomic.AddInt64(&h.svc.metrics.AnalyzeTimeouts, 1)
	}
	if err != nil {
		atomic.AddInt64(&h.svc.metrics.AnalyzeFailures, 1)
	}
}

func (h *EngineHandler) recordPlay(start time.Time, err error) {
	atomic.AddInt64(&h.svc.metrics.PlayRequests, 1)
	atomic.AddInt64(&h.svc.metrics.TotalLatencyMicro, time.Since(start).Microseconds())
	if err != nil {
		atomic.AddInt64(&h.svc.metrics.PlayFailures, 1)
	}
}

type EngineHandler struct {
	svc     *EngineService
	userSvc *users.UserService
	cache   *AnalysisCache
}

func NewEngineHandler(svc *EngineService, userSvc *users.UserService) *EngineHandler {
	return &EngineHandler{
		svc:     svc,
		userSvc: userSvc,
		cache:   NewAnalysisCache(256, 10*time.Minute),
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

	// Check cache (bounded + TTL). Only return if depth and line count are sufficient.
	if cachedResult, ok := h.cache.Load(req.FEN, analysisDepth, multiPV); ok {
		c.JSON(http.StatusOK, cachedResult)
		return
	}

	start := time.Now()
	// Create context with timeout for analysis (movetime + 3s buffer)
	timeout := time.Duration(req.Movetime+3000) * time.Millisecond
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	result, err := h.svc.AnalyzePosition(ctx, req.FEN, analysisDepth, req.Movetime)
	timedOut := err == context.DeadlineExceeded
	h.recordAnalyze(start, err, timedOut)
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
}// CacheSize returns the current analysis cache size (for admin/observability).
func (h *EngineHandler) CacheSize() int {
	return h.cache.Size()
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

	start := time.Now()
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
	h.recordPlay(start, err)
	if err != nil {
		fmt.Printf("PlayMove error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "engine play failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

