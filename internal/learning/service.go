package learning

import (
	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LearningService struct {
	db *gorm.DB
}

func NewLearningService(db *gorm.DB) *LearningService {
	return &LearningService{db: db}
}

// ============================================================================
// Learning Categories
// ============================================================================

func (s *LearningService) ListCategories() ([]database.LearningCategory, error) {
	var cats []database.LearningCategory
	err := s.db.Order("created_at asc").Find(&cats).Error
	return cats, err
}

func (s *LearningService) GetCategory(id uuid.UUID) (*database.LearningCategory, error) {
	var cat database.LearningCategory
	err := s.db.Preload("Lessons").First(&cat, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *LearningService) CreateCategory(cat *database.LearningCategory) error {
	if cat.ID == uuid.Nil {
		cat.ID = uuid.New()
	}
	return s.db.Create(cat).Error
}

func (s *LearningService) UpdateCategory(cat *database.LearningCategory) error {
	return s.db.Save(cat).Error
}

func (s *LearningService) DeleteCategory(id uuid.UUID) error {
	return s.db.Delete(&database.LearningCategory{}, "id = ?", id).Error
}

// ============================================================================
// Lessons
// ============================================================================

func (s *LearningService) ListLessons(categoryID *uuid.UUID) ([]database.Lesson, error) {
	var lessons []database.Lesson
	query := s.db.Order("created_at asc")
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	err := query.Find(&lessons).Error
	return lessons, err
}

func (s *LearningService) GetLesson(id uuid.UUID) (*database.Lesson, error) {
	var lesson database.Lesson
	err := s.db.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order asc")
	}).First(&lesson, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &lesson, nil
}

func (s *LearningService) CreateLesson(lesson *database.Lesson) error {
	if lesson.ID == uuid.Nil {
		lesson.ID = uuid.New()
	}
	return s.db.Create(lesson).Error
}

func (s *LearningService) UpdateLesson(lesson *database.Lesson) error {
	return s.db.Save(lesson).Error
}

func (s *LearningService) DeleteLesson(id uuid.UUID) error {
	return s.db.Delete(&database.Lesson{}, "id = ?", id).Error
}

// ============================================================================
// Lesson Steps
// ============================================================================

func (s *LearningService) ListSteps(lessonID uuid.UUID) ([]database.LessonStep, error) {
	var steps []database.LessonStep
	err := s.db.Order("step_order asc").Find(&steps, "lesson_id = ?", lessonID).Error
	return steps, err
}

func (s *LearningService) GetStep(id uuid.UUID) (*database.LessonStep, error) {
	var step database.LessonStep
	err := s.db.First(&step, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

func (s *LearningService) CreateStep(step *database.LessonStep) error {
	if step.ID == uuid.Nil {
		step.ID = uuid.New()
	}
	// Automatically calculate step order if 0 or not passed
	if step.StepOrder == 0 {
		var maxOrder int
		row := s.db.Model(&database.LessonStep{}).Where("lesson_id = ?", step.LessonID).Select("COALESCE(MAX(step_order), 0)").Row()
		_ = row.Scan(&maxOrder)
		step.StepOrder = maxOrder + 1
	}
	return s.db.Create(step).Error
}

func (s *LearningService) UpdateStep(step *database.LessonStep) error {
	return s.db.Save(step).Error
}

func (s *LearningService) DeleteStep(id uuid.UUID) error {
	return s.db.Delete(&database.LessonStep{}, "id = ?", id).Error
}

func (s *LearningService) ReorderSteps(lessonID uuid.UUID, stepIDs []uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i, stepID := range stepIDs {
			err := tx.Model(&database.LessonStep{}).
				Where("id = ? AND lesson_id = ?", stepID, lessonID).
				Update("step_order", i+1).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}
