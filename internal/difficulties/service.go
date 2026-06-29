package difficulties

import (
	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DifficultyService struct {
	db *gorm.DB
}

func NewDifficultyService(db *gorm.DB) *DifficultyService {
	return &DifficultyService{db: db}
}

func (s *DifficultyService) CreateDifficulty(diff *database.Difficulty) error {
	if diff.ID == uuid.Nil {
		diff.ID = uuid.New()
	}
	return s.db.Create(diff).Error
}

func (s *DifficultyService) ListDifficulties() ([]database.Difficulty, error) {
	var diffs []database.Difficulty
	err := s.db.Order("level asc").Find(&diffs).Error
	return diffs, err
}

func (s *DifficultyService) GetDifficulty(id uuid.UUID) (*database.Difficulty, error) {
	var diff database.Difficulty
	if err := s.db.First(&diff, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &diff, nil
}

func (s *DifficultyService) UpdateDifficulty(diff *database.Difficulty) error {
	return s.db.Save(diff).Error
}

func (s *DifficultyService) DeleteDifficulty(id uuid.UUID) error {
	return s.db.Delete(&database.Difficulty{}, "id = ?", id).Error
}
