package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base contains common columns for other models.
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedBy string         `gorm:"default:'System'"`
	UpdatedBy string
	DeletedBy string
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()
	return nil
}
func (b *Base) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}

type User struct {
	Base
	Username     string `gorm:"uniqueIndex"`
	Email        string `gorm:"uniqueIndex"`
	PasswordHash string
}

type Machine struct {
	Base
	Name       string    `gorm:"uniqueIndex"`
	UserID     uuid.UUID `gorm:"type:uuid"`
	TemplateID uuid.UUID `gorm:"type:uuid"`
	Status     bool
	Password   string
	ExpiresAt  time.Time
	URL        string
	// Tasks       []Task
}

type Template struct {
	Base
	Name        string `gorm:"uniqueIndex"`
	Description string
	UserID      uuid.UUID `gorm:"type:uuid"`
	// Files []File
}

type Task struct {
	Base
	MachineID uuid.UUID `gorm:"type:uuid"`
	Type      string
	FileID    uuid.UUID `gorm:"type:uuid"`
	Status    string
	Message   string
}

type File struct {
	Base
	Name       string `gorm:"uniqueIndex"`
	Path       string
	Size       int64
	Type       string
	TemplateID uuid.UUID `gorm:"type:uuid"`
	UserID     uuid.UUID `gorm:"type:uuid"`
}

/*
the following SQL command creates the required extensions in your PostgreSQL database:
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
*/
