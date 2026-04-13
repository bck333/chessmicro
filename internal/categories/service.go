package categories

import (
	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) CreateCategory(cat *database.Category) error {
	cat.ID = uuid.New()
	return s.db.Create(cat).Error
}

func (s *CategoryService) ListCategories() ([]database.Category, error) {
	var cats []database.Category
	err := s.db.Find(&cats).Error
	return cats, err
}

func (s *CategoryService) UpdateCategory(cat *database.Category) error {
	return s.db.Save(cat).Error
}

func (s *CategoryService) DeleteCategory(id uuid.UUID) error {
	return s.db.Delete(&database.Category{}, id).Error
}

func (s *CategoryService) GetCategory(id uuid.UUID) (*database.Category, error) {
	var cat database.Category
	err := s.db.First(&cat, id).Error
	return &cat, err
}
