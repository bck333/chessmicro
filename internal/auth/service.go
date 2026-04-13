package auth

import (
	"github.com/chess-puzzle-app/backend/internal/config"
	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/chess-puzzle-app/backend/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

func (s *AuthService) GuestLogin(name string) (string, *database.User, error) {
	user := &database.User{
		ID:      uuid.New(),
		Name:    name,
		IsGuest: true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return "", nil, err
	}

	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *AuthService) GoogleLogin(idToken string) (string, *database.User, error) {
	// In Phase 1, we will mock the Google token verification.
	// For production, use "google.golang.org/api/oauth2/v2" to verify the token.
	
	// Mock implementation
	email := "mockuser@gmail.com"
	name := "Mock Google User"

	var user database.User
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			user = database.User{
				ID:    uuid.New(),
				Email: &email,
				Name:  name,
			}
			if err := s.db.Create(&user).Error; err != nil {
				return "", nil, err
			}
		} else {
			return "", nil, err
		}
	}

	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}
