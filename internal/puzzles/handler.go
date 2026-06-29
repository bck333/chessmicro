package puzzles

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/chess-puzzle-app/backend/internal/users"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PuzzleResponse struct {
	ID            uuid.UUID          `json:"id"`
	Rating        int                `json:"rating"`
	Plays         int                `json:"plays"`
	SolutionMoves []string           `json:"solution"`
	Themes        []string           `json:"themes"`
	FEN           string             `json:"fen"`
	LastMove      string             `json:"lastMove"`
	Title         string             `json:"title"`
	Description   string             `json:"description"`
	Difficulty    string             `json:"difficulty"`
	Opening       string             `json:"opening"`
	ECO           string             `json:"eco"`
	White         string             `json:"white"`
	Black         string             `json:"black"`
	Year          int                `json:"year"`
	Source        string             `json:"source"`
	CreatedBy     uuid.UUID          `json:"created_by"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

func toPuzzleResponse(p database.Puzzle) PuzzleResponse {
	moves := []string{}
	if p.SolutionMoves != "" {
		moves = strings.Split(p.SolutionMoves, ",")
	}
	themes := make([]string, len(p.Categories))
	for i, cat := range p.Categories {
		themes[i] = cat.Name
	}
	return PuzzleResponse{
		ID:            p.ID,
		Rating:        p.Rating,
		Plays:         p.Plays,
		SolutionMoves: moves,
		Themes:        themes,
		FEN:           p.FEN,
		LastMove:      p.LastMove,
		Title:         p.Title,
		Description:   p.Description,
		Difficulty:    p.Difficulty,
		Opening:       p.Opening,
		ECO:           p.ECO,
		White:         p.White,
		Black:         p.Black,
		Year:          p.Year,
		Source:        p.Source,
		CreatedBy:     p.CreatedBy,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

type PuzzleHandler struct {
	svc     *PuzzleService
	userSvc *users.UserService
}

func NewPuzzleHandler(svc *PuzzleService, userSvc *users.UserService) *PuzzleHandler {
	return &PuzzleHandler{svc: svc, userSvc: userSvc}
}

func (h *PuzzleHandler) CreatePuzzle(c *gin.Context) {
	var input struct {
		Title         string      `json:"title" binding:"required"`
		Description   string      `json:"description" binding:"required"`
		FEN           string      `json:"fen" binding:"required"`
		SolutionMoves []string    `json:"solution_moves" binding:"required"`
		Difficulty    string      `json:"difficulty" binding:"required"`
		Rating        int         `json:"rating"`
		LastMove      string      `json:"lastMove"`
		Opening       string      `json:"opening"`
		ECO           string      `json:"eco"`
		White         string      `json:"white"`
		Black         string      `json:"black"`
		Year          int         `json:"year"`
		Source        string      `json:"source"`
		CategoryIDs   []uuid.UUID `json:"category_ids"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	puzzle := &database.Puzzle{
		Title:         input.Title,
		Description:   input.Description,
		FEN:           input.FEN,
		SolutionMoves: strings.Join(input.SolutionMoves, ","),
		Difficulty:    input.Difficulty,
		Rating:        input.Rating,
		LastMove:      input.LastMove,
		Opening:       input.Opening,
		ECO:           input.ECO,
		White:         input.White,
		Black:         input.Black,
		Year:          input.Year,
		Source:        input.Source,
		CreatedBy:     userID,
	}

	if err := h.svc.CreatePuzzleWithCategories(puzzle, input.CategoryIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"puzzle": toPuzzleResponse(*puzzle)})
}

func (h *PuzzleHandler) GetPuzzle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid puzzle id"})
		return
	}

	// Check and increment puzzle usage limit
	userID := c.MustGet("user_id").(uuid.UUID)
	canUse, err := h.userSvc.UsePuzzle(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check limits"})
		return
	}
	if !canUse {
		c.JSON(http.StatusForbidden, gin.H{"error": "daily puzzle limit reached", "code": "LIMIT_REACHED"})
		return
	}

	puzzle, err := h.svc.GetPuzzle(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "puzzle not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"puzzle": toPuzzleResponse(*puzzle)})
}

func (h *PuzzleHandler) GetDailyPuzzle(c *gin.Context) {
	// Check and increment puzzle usage limit
	userID := c.MustGet("user_id").(uuid.UUID)
	canUse, err := h.userSvc.UsePuzzle(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check limits"})
		return
	}
	if !canUse {
		c.JSON(http.StatusForbidden, gin.H{"error": "daily puzzle limit reached", "code": "LIMIT_REACHED"})
		return
	}

	puzzle, err := h.svc.GetDailyPuzzle()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "daily puzzle not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"puzzle": toPuzzleResponse(*puzzle)})
}

func (h *PuzzleHandler) ListPuzzles(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	categoryIDStr := c.Query("category_id")
	difficulty := c.Query("difficulty")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	var categoryID *uuid.UUID
	if categoryIDStr != "" {
		id, err := uuid.Parse(categoryIDStr)
		if err == nil {
			categoryID = &id
		}
	}

	puzzles, total, err := h.svc.ListPuzzles(page, limit, categoryID, difficulty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]PuzzleResponse, len(puzzles))
	for i, p := range puzzles {
		responses[i] = toPuzzleResponse(p)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *PuzzleHandler) UpdatePuzzle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid puzzle id"})
		return
	}

	var input struct {
		Title         string      `json:"title"`
		Description   string      `json:"description"`
		FEN           string      `json:"fen"`
		SolutionMoves []string    `json:"solution_moves"`
		Difficulty    string      `json:"difficulty"`
		Rating        int         `json:"rating"`
		LastMove      string      `json:"lastMove"`
		Opening       string      `json:"opening"`
		ECO           string      `json:"eco"`
		White         string      `json:"white"`
		Black         string      `json:"black"`
		Year          int         `json:"year"`
		Source        string      `json:"source"`
		CategoryIDs   []uuid.UUID `json:"category_ids"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	puzzle, err := h.svc.GetPuzzle(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "puzzle not found"})
		return
	}

	if input.Title != "" {
		puzzle.Title = input.Title
	}
	if input.Description != "" {
		puzzle.Description = input.Description
	}
	if input.FEN != "" {
		puzzle.FEN = input.FEN
	}
	if input.SolutionMoves != nil {
		puzzle.SolutionMoves = strings.Join(input.SolutionMoves, ",")
	}
	if input.Difficulty != "" {
		puzzle.Difficulty = input.Difficulty
	}
	if input.Rating != 0 {
		puzzle.Rating = input.Rating
	}
	if input.LastMove != "" {
		puzzle.LastMove = input.LastMove
	}
	if input.Opening != "" {
		puzzle.Opening = input.Opening
	}
	if input.ECO != "" {
		puzzle.ECO = input.ECO
	}
	if input.White != "" {
		puzzle.White = input.White
	}
	if input.Black != "" {
		puzzle.Black = input.Black
	}
	if input.Year != 0 {
		puzzle.Year = input.Year
	}
	if input.Source != "" {
		puzzle.Source = input.Source
	}

	if err := h.svc.UpdatePuzzle(puzzle, input.CategoryIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"puzzle": toPuzzleResponse(*puzzle)})
}

func (h *PuzzleHandler) DeletePuzzle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid puzzle id"})
		return
	}

	if err := h.svc.DeletePuzzle(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PuzzleHandler) SolvePuzzle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid puzzle ID"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	if err := h.svc.SolvePuzzle(userID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Progress saved successfully"})
}
