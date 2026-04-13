package admin

import (
	"github.com/chess-puzzle-app/backend/internal/database"
	"gorm.io/gorm"
)

type AdminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) GetStats() (map[string]int64, error) {
	var puzzleCount int64
	var userCount int64

	if err := s.db.Model(&database.Puzzle{}).Count(&puzzleCount).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&database.User{}).Count(&userCount).Error; err != nil {
		return nil, err
	}

	return map[string]int64{
		"total_puzzles": puzzleCount,
		"total_users":   userCount,
	}, nil
}

func (s *AdminService) ListUsers(page, limit int) ([]database.User, int64, error) {
	var users []database.User
	var total int64

	s.db.Model(&database.User{}).Count(&total)

	offset := (page - 1) * limit
	err := s.db.Order("created_at desc").Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}
