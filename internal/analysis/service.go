package analysis

import (
	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnalysisService struct {
	db *gorm.DB
}

func NewAnalysisService(db *gorm.DB) *AnalysisService {
	return &AnalysisService{db: db}
}

// ListSessions returns all analysis sessions for a user, newest first.
func (s *AnalysisService) ListSessions(userID uuid.UUID) ([]database.AnalysisSession, error) {
	var sessions []database.AnalysisSession
	err := s.db.Where("user_id = ?", userID).Order("updated_at desc").Find(&sessions).Error
	return sessions, err
}

// GetSession returns a single session owned by userID.
func (s *AnalysisService) GetSession(userID, sessionID uuid.UUID) (*database.AnalysisSession, error) {
	var session database.AnalysisSession
	err := s.db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// CreateSessionInput is the payload accepted when creating a new session.
type CreateSessionInput struct {
	Title         string `json:"title"`
	RootFEN       string `json:"root_fen" binding:"required"`
	CurrentNodeID string `json:"current_node_id"`
	TreeJSON      string `json:"tree_json" binding:"required"`
}

// CreateSession persists a new analysis session for a user.
func (s *AnalysisService) CreateSession(userID uuid.UUID, input CreateSessionInput) (*database.AnalysisSession, error) {
	currentNode := input.CurrentNodeID
	if currentNode == "" {
		currentNode = "root"
	}
	session := database.AnalysisSession{
		ID:            uuid.New(),
		UserID:        userID,
		Title:         input.Title,
		RootFEN:       input.RootFEN,
		CurrentNodeID: currentNode,
		TreeJSON:      input.TreeJSON,
	}
	if err := s.db.Create(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSessionInput is the payload accepted when updating an existing session.
type UpdateSessionInput struct {
	Title         *string `json:"title"`
	CurrentNodeID *string `json:"current_node_id"`
	TreeJSON      *string `json:"tree_json"`
}

// UpdateSession overwrites mutable fields of a session owned by userID.
func (s *AnalysisService) UpdateSession(userID, sessionID uuid.UUID, input UpdateSessionInput) error {
	updates := map[string]interface{}{}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.CurrentNodeID != nil {
		updates["current_node_id"] = *input.CurrentNodeID
	}
	if input.TreeJSON != nil {
		updates["tree_json"] = *input.TreeJSON
	}
	if len(updates) == 0 {
		return nil
	}
	result := s.db.Model(&database.AnalysisSession{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteSession removes a session owned by userID.
func (s *AnalysisService) DeleteSession(userID, sessionID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", sessionID, userID).Delete(&database.AnalysisSession{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
