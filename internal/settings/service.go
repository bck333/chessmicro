package settings

import (
	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SettingService struct {
	db *gorm.DB
}

func NewSettingService(db *gorm.DB) *SettingService {
	return &SettingService{db: db}
}

func (s *SettingService) UpdateSetting(key string, value string, description string) error {
	var setting database.Setting
	err := s.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		// Create new
		setting = database.Setting{
			ID:          uuid.New(),
			Key:         key,
			Value:       value,
			Description: description,
		}
		return s.db.Create(&setting).Error
	}
	// Update existing
	setting.Value = value
	if description != "" {
		setting.Description = description
	}
	return s.db.Save(&setting).Error
}

func (s *SettingService) GetSetting(key string) (string, error) {
	var setting database.Setting
	err := s.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *SettingService) ListSettings() ([]database.Setting, error) {
	var settings []database.Setting
	err := s.db.Find(&settings).Error
	return settings, err
}
