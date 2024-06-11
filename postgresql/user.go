package postgresql

import (
	"context"

	"github.com/google/uuid"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/postgresql/models"
	"gorm.io/gorm"
)

// UserRepository defines methods for interacting with User entities.
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

// userRepository implements the UserRepository interface.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser creates a new user in the database.
func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return common.Err("Failed to create user: %v", result.Error)
	}
	return nil
}

// GetUserByID retrieves a user by their ID.
func (r *userRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, "id = ?", userID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, common.Err("User not found: %v", result.Error)
		}
		return nil, common.Err("Failed to get user by ID: %v", result.Error)
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by their Username.
func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, "username = ?", username)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, common.Err("User not found: %v", result.Error)
		}
		return nil, common.Err("Failed to get user by username: %v", result.Error)
	}
	return &user, nil
}