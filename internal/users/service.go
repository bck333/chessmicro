package users

import (
	"time"

	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetUserByID(id uuid.UUID) (*database.User, error) {
	var user database.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) TrackLogin(id uuid.UUID) (*database.User, error) {
	var user database.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	
	// Handle Daily Limit Resets
	if user.LastLimitReset == nil || user.LastLimitReset.Day() != now.Day() || user.LastLimitReset.Month() != now.Month() || user.LastLimitReset.Year() != now.Year() {
		user.PuzzlesUsedToday = 0
		user.GamesUsedToday = 0
		user.HintsUsedToday = 0
		user.LastLimitReset = &now
	}

	if user.LastActiveDate == nil {
		user.StreakCount = 1
	} else {
		y1, m1, d1 := user.LastActiveDate.Date()
		y2, m2, d2 := now.Date()
		// ... existing streak logic ...
		if y1 == y2 && m1 == m2 && d1 == d2 {
			// Same day
		} else {
			date1 := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
			date2 := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
			diff := date2.Sub(date1).Hours()
			if diff == 24 {
				user.StreakCount++
			} else {
				user.StreakCount = 1
			}
		}
	}

	user.LastActiveDate = &now
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) AddXP(id uuid.UUID, amount int) error {
	return s.db.Model(&database.User{}).Where("id = ?", id).UpdateColumn("xp", gorm.Expr("xp + ?", amount)).Error
}

func (s *UserService) UseHint(id uuid.UUID) (bool, error) {
	var user database.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return false, err
	}

	now := time.Now()
	if user.LastLimitReset == nil || user.LastLimitReset.Day() != now.Day() || user.LastLimitReset.Month() != now.Month() || user.LastLimitReset.Year() != now.Year() {
		user.PuzzlesUsedToday = 0
		user.GamesUsedToday = 0
		user.HintsUsedToday = 0
		user.LastLimitReset = &now
	}

	// Limit checks based on Subscription
	limit := 3 // Guest
	if user.SubscriptionType == "free" {
		limit = 3
	} else if user.SubscriptionType == "starter" {
		limit = 5
	} else if user.SubscriptionType == "pro" {
		limit = 10
	} else if user.SubscriptionType == "elite" || user.SubscriptionType == "coach" {
		limit = 9999 // Unlimited
	}

	if user.HintsUsedToday >= limit {
		return false, nil
	}

	user.HintsUsedToday++
	if err := s.db.Save(&user).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (s *UserService) UsePuzzle(id uuid.UUID) (bool, error) {
	var user database.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return false, err
	}

	now := time.Now()
	// Check reset
	if user.LastLimitReset == nil || user.LastLimitReset.Day() != now.Day() || user.LastLimitReset.Month() != now.Month() || user.LastLimitReset.Year() != now.Year() {
		user.PuzzlesUsedToday = 0
		user.GamesUsedToday = 0
		user.HintsUsedToday = 0
		user.LastLimitReset = &now
	}

	limit := 3 // Guest
	switch user.SubscriptionType {
	case "free":
		limit = 10
	case "starter":
		limit = 30
	case "pro":
		limit = 80
	case "elite", "coach":
		limit = 99999
	}

	if user.PuzzlesUsedToday >= limit {
		return false, nil
	}

	user.PuzzlesUsedToday++
	return true, s.db.Save(&user).Error
}

func (s *UserService) UseGame(id uuid.UUID) (bool, error) {
	var user database.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return false, err
	}

	now := time.Now()
	if user.LastLimitReset == nil || user.LastLimitReset.Day() != now.Day() || user.LastLimitReset.Month() != now.Month() || user.LastLimitReset.Year() != now.Year() {
		user.PuzzlesUsedToday = 0
		user.GamesUsedToday = 0
		user.HintsUsedToday = 0
		user.LastLimitReset = &now
	}

	limit := 1 // Guest
	switch user.SubscriptionType {
	case "free":
		limit = 2
	case "starter":
		limit = 5
	case "pro", "elite", "coach":
		limit = 99999
	}

	if user.GamesUsedToday >= limit {
		return false, nil
	}

	user.GamesUsedToday++
	return true, s.db.Save(&user).Error
}

func GetRank(xp int) string {
	if xp < 100 {
		return "Beginner"
	}
	if xp < 500 {
		return "Learner"
	}
	if xp < 1000 {
		return "Thinker"
	}
	if xp < 2500 {
		return "Strategist"
	}
	return "Master"
}

func (s *UserService) SaveProgress(userID uuid.UUID, categoryID uuid.UUID, categoryName string, progressType string, stepNumber int, status string, xpEarned int) (*database.UserProgress, error) {
	var progress database.UserProgress
	err := s.db.Where("user_id = ? AND category_name = ? AND progress_type = ?", userID, categoryName, progressType).First(&progress).Error
	
	if err == gorm.ErrRecordNotFound {
		progress = database.UserProgress{
			ID:           uuid.New(),
			UserID:       userID,
			CategoryID:   categoryID,
			CategoryName: categoryName,
			ProgressType: progressType,
			StepNumber:   stepNumber,
			Status:       status,
			XPEarned:     xpEarned,
			UpdatedAt:    time.Now(),
		}
		if err := s.db.Create(&progress).Error; err != nil {
			return nil, err
		}
	} else if err == nil {
		progress.CategoryID = categoryID
		if stepNumber > progress.StepNumber {
			progress.StepNumber = stepNumber
		}
		progress.Status = status
		progress.XPEarned += xpEarned
		progress.UpdatedAt = time.Now()
		if err := s.db.Save(&progress).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return &progress, nil
}

func (s *UserService) GetProgressList(userID uuid.UUID) ([]database.UserProgress, error) {
	var progressList []database.UserProgress
	err := s.db.Where("user_id = ?", userID).Order("updated_at desc").Find(&progressList).Error
	return progressList, err
}
