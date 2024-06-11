package postgresql

import (
	"context"

	"github.com/google/uuid"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/postgresql/models"
	"gorm.io/gorm"
)

// TemplateRepository defines methods for interacting with Template entities.
type TemplateRepository interface {
	CreateTemplate(ctx context.Context, template *models.Template) error
	GetTemplateByID(ctx context.Context, templateID uuid.UUID) (*models.Template, error)
	GetTemplateByName(ctx context.Context, templateName string) (*models.Template, error)
	DeleteTemplate(ctx context.Context, templateID uuid.UUID) error
}

type templateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository creates a new instance of TemplateRepository.
func NewTemplateRepository(db *gorm.DB) TemplateRepository {
	return &templateRepository{db: db}
}

// CreateTemplate creates a new template record in the database.
func (r *templateRepository) CreateTemplate(ctx context.Context, template *models.Template) error {
	result := r.db.WithContext(ctx).Create(template)
	if result.Error != nil {
		return common.Err("Failed to create template: %v", result.Error)
	}
	return nil
}

// GetTemplateByName retrieves a template by its name from the database.
func (r *templateRepository) GetTemplateByName(ctx context.Context, templateName string) (*models.Template, error) {
	var template models.Template
	result := r.db.WithContext(ctx).First(&template, "name = ?", templateName)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, common.Err("Template not found: %v", result.Error)
		}
		return nil, common.Err("Failed to get template by name: %v", result.Error)
	}
	return &template, nil
}

// GetTemplateByID retrieves a template by its ID from the database.
func (r *templateRepository) GetTemplateByID(ctx context.Context, templateID uuid.UUID) (*models.Template, error) {
	var template models.Template
	result := r.db.WithContext(ctx).First(&template, "id = ?", templateID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, common.Err("Template not found: %v", result.Error)
		}
		return nil, common.Err("Failed to get template by ID: %v", result.Error)
	}
	return &template, nil
}

// DeleteTemplate deletes a template record from the database.
func (r *templateRepository) DeleteTemplate(ctx context.Context, templateID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Template{}, "id = ?", templateID)
	if result.Error != nil {
		return common.Err("Failed to delete template: %v", result.Error)
	}
	return nil
}
