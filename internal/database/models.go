package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"uniqueIndex"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Setting struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Key         string    `gorm:"uniqueIndex"`
	Value       string
	Description string
	UpdatedAt   time.Time
}

type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email       *string   `gorm:"uniqueIndex"`
	Name        string
	IsGuest     bool      `gorm:"default:false"`
	SolvedCount int       `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
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

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Puzzle{}, &Category{}, &Setting{})
}
