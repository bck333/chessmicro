package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex" json:"name"`
	Description string    `json:"description"`
	Group       string    `gorm:"default:'practice'" json:"group"` // "learn" (for basics/pieces), "practice" (for motifs)
	FEN         string    `gorm:"default:''" json:"fen"`
	Targets     string    `gorm:"default:''" json:"targets"` // comma-separated list of target squares, e.g. "e1,e2,e3"
	XP          int       `gorm:"default:0" json:"xp"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Difficulty struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex" json:"name"` // Easy, Medium, Hard, Master, etc.
	Level       int       `gorm:"default:0" json:"level"`    // Order/Weight
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Setting struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Key         string    `gorm:"uniqueIndex"`
	Value       string
	Description string
	UpdatedAt   time.Time
}

type User struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Email             *string    `gorm:"uniqueIndex"`
	Name              string
	IsGuest           bool       `gorm:"default:false"`
	SubscriptionType  string     `gorm:"default:'free'"` // guest, free, starter, pro, elite, coach
	SubscriptionExpiry *time.Time
	SolvedCount       int        `gorm:"default:0"`
	XP                int        `gorm:"default:0"`
	StreakCount       int        `gorm:"default:0"`
	LastActiveDate    *time.Time
	
	// Daily Limit Tracking
	PuzzlesUsedToday  int        `gorm:"default:0"`
	GamesUsedToday    int        `gorm:"default:0"`
	HintsUsedToday    int        `gorm:"default:0"`
	LastLimitReset    *time.Time
	
	Rating            int        `gorm:"default:1200"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Puzzle struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Title         string
	Description   string
	FEN           string
	SolutionMoves string     `gorm:"type:text"`
	Difficulty    string     `gorm:"index"` // Legacy / Main difficulty
	Rating        int        `gorm:"index;default:1500"` // New: Rating (e.g. 1980)
	Plays         int        `gorm:"default:0"`           // New: Number of times played
	LastMove      string                                  // New: The move leading to this position
	Opening       string                                  // New: Opening name (e.g. Sicilian)
	ECO           string                                  // New: ECO code (e.g. B40)
	White         string                                  // New: White player name
	Black         string                                  // New: Black player name
	Year          int                                     // New: Year of the game
	Source        string                                  // New: Tournament or source URL
	Categories    []Category `gorm:"many2many:puzzle_categories;"`
	CreatedBy     uuid.UUID  `gorm:"type:uuid"`
	CreatedAt     time.Time  `gorm:"index"`
	UpdatedAt     time.Time
}

type UserProgress struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	CategoryID   uuid.UUID `gorm:"type:uuid;index" json:"category_id"`
	CategoryName string    `json:"category_name"`
	ProgressType string    `json:"progress_type"` // "learn", "practice"
	StepNumber   int       `json:"step_number"`   // current step (for basics/pieces) or solved count (for motifs)
	Status       string    `json:"status"`        // "started", "completed"
	XPEarned     int       `json:"xp_earned"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LearningCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string    `gorm:"uniqueIndex" json:"title"`
	Icon        string    `json:"icon"` // e.g. "castle", "award", "book"
	Description string    `json:"description"`
	Lessons     []Lesson  `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"lessons,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Lesson struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	CategoryID  uuid.UUID    `gorm:"type:uuid;index" json:"category_id"`
	Title       string       `json:"title"`
	Difficulty  string       `json:"difficulty"` // e.g. "Beginner", "Intermediate"
	Thumbnail   string       `json:"thumbnail"`  // e.g. "rook_thumb"
	Steps       []LessonStep `gorm:"foreignKey:LessonID;constraint:OnDelete:CASCADE" json:"steps,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type LessonStep struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	LessonID           uuid.UUID `gorm:"type:uuid;index" json:"lesson_id"`
	StepOrder          int       `gorm:"default:0" json:"step_order"`
	Title              string    `json:"title"`
	Description        string    `gorm:"type:text" json:"description"` // supports rich text explanation
	FEN                string    `json:"fen"`         // starting chessboard state
	HighlightedSquares string    `json:"highlighted_squares"` // e.g. "e4,d5"
	MoveArrows         string    `json:"move_arrows"`         // e.g. "g1f3,e2e4"
	ImageURL           string    `json:"image_url"`           // optional
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// AnalysisSession stores a user's saved analysis move tree as JSON so it can
// be resumed later. The tree is stored as a single JSON blob to avoid
// over-normalizing until there is a clear product need for node-level queries.
type AnalysisSession struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Title         string    `json:"title"`
	RootFEN       string    `json:"root_fen"`
	CurrentNodeID string    `gorm:"default:'root'" json:"current_node_id"`
	TreeJSON      string    `gorm:"type:text" json:"tree_json"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&User{}, &Puzzle{}, &Category{}, &Setting{}, &Difficulty{}, &UserProgress{}, &LearningCategory{}, &Lesson{}, &LessonStep{}, &AnalysisSession{}); err != nil {
		return err
	}
	return Seed(db)
}

func Seed(db *gorm.DB) error {
	// Seed function is now empty to allow all content to be managed via API/Admin only.
	return nil
}
