package puzzles

import (
	"log"
	"time"

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

func (s *PuzzleService) ListPuzzles(page, limit int, categoryID *uuid.UUID, difficultyName string) ([]database.Puzzle, int64, error) {
	query := s.db.Model(&database.Puzzle{})

	if categoryID != nil {
		query = query.Where("id IN (SELECT puzzle_id FROM puzzle_categories WHERE category_id = ?)", *categoryID)
	}

	if difficultyName != "" {
		query = query.Where("LOWER(difficulty) = LOWER(?)", difficultyName)
	}

	// Debug prints
	var catStr string
	if categoryID != nil { catStr = categoryID.String() }
	log.Printf("[DEBUG] ListPuzzles: Category=%s, Difficulty=%s", catStr, difficultyName)

	var total int64
	query.Count(&total)

	var puzzles []database.Puzzle
	offset := (page - 1) * limit
	err := query.Preload("Categories").Limit(limit).Offset(offset).Order("created_at desc").Find(&puzzles).Error
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

func (s *PuzzleService) GetDailyPuzzle() (*database.Puzzle, error) {
	var total int64
	s.db.Model(&database.Puzzle{}).Count(&total)

	if total == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	dayOfYear := int64(time.Now().YearDay() + time.Now().Year())
	offset := int(dayOfYear % total)

	var puzzle database.Puzzle
	if err := s.db.Preload("Categories").Order("id asc").Offset(offset).Limit(1).Find(&puzzle).Error; err != nil {
		return nil, err
	}
	return &puzzle, nil
}

func (s *PuzzleService) SolvePuzzle(userID, puzzleID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&database.User{}).Where("id = ?", userID).UpdateColumns(map[string]interface{}{
			"solved_count": gorm.Expr("solved_count + ?", 1),
			"xp":           gorm.Expr("xp + ?", 10), // Give 10 XP for solving
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&database.Puzzle{}).Where("id = ?", puzzleID).Update("plays", gorm.Expr("plays + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}
