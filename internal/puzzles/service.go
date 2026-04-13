package puzzles

import (
	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PuzzleService struct {
	db *gorm.DB
}

func NewPuzzleService(db *gorm.DB) *PuzzleService {
	return &PuzzleService{db: db}
}

func (s *PuzzleService) CreatePuzzle(puzzle *database.Puzzle) error {
	puzzle.ID = uuid.New()
	return s.db.Create(puzzle).Error
}

func (s *PuzzleService) GetPuzzle(id uuid.UUID) (*database.Puzzle, error) {
	var puzzle database.Puzzle
	if err := s.db.Preload("Categories").First(&puzzle, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &puzzle, nil
}

func (s *PuzzleService) ListPuzzles(page, limit int) ([]database.Puzzle, int64, error) {
	var total int64
	s.db.Model(&database.Puzzle{}).Count(&total)

	var puzzles []database.Puzzle
	offset := (page - 1) * limit
	err := s.db.Preload("Categories").Limit(limit).Offset(offset).Order("created_at desc").Find(&puzzles).Error
	return puzzles, total, err
}

func (s *PuzzleService) UpdatePuzzle(puzzle *database.Puzzle, categoryIDs []uuid.UUID) error {
	// Update simple fields
	if err := s.db.Save(puzzle).Error; err != nil {
		return err
	}

	// Update associations
	if categoryIDs != nil {
		var categories []database.Category
		if len(categoryIDs) > 0 {
			s.db.Find(&categories, "id IN ?", categoryIDs)
		}
		return s.db.Model(puzzle).Association("Categories").Replace(categories)
	}

	return nil
}

func (s *PuzzleService) CreatePuzzleWithCategories(puzzle *database.Puzzle, categoryIDs []uuid.UUID) error {
	puzzle.ID = uuid.New()

	if len(categoryIDs) > 0 {
		var categories []database.Category
		s.db.Find(&categories, "id IN ?", categoryIDs)
		puzzle.Categories = categories
	}

	return s.db.Create(puzzle).Error
}

func (s *PuzzleService) DeletePuzzle(id uuid.UUID) error {
	return s.db.Delete(&database.Puzzle{}, "id = ?", id).Error
}
func (s *PuzzleService) SolvePuzzle(userID, puzzleID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Increment User SolvedCount
		if err := tx.Model(&database.User{}).Where("id = ?", userID).Update("solved_count", gorm.Expr("solved_count + ?", 1)).Error; err != nil {
			return err
		}

		// Increment Puzzle Plays
		if err := tx.Model(&database.Puzzle{}).Where("id = ?", puzzleID).Update("plays", gorm.Expr("plays + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}
