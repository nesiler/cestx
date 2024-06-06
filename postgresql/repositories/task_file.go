package postgresql


import (
	"context"

	"github.com/google/uuid"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/postgresql/models"
	"gorm.io/gorm"
)

// TaskFileRepository defines methods for interacting with Task and File entities.
type TaskFileRepository interface {
	CreateTask(ctx context.Context, task *models.Task) error
	GetTaskByID(ctx context.Context, taskID uuid.UUID) (*models.Task, error)
	CreateFile(ctx context.Context, file *models.File) error
	GetFileByID(ctx context.Context, fileID uuid.UUID) (*models.File, error)
}

type taskFileRepository struct {
	db *gorm.DB
}

// NewTaskFileRepository creates a new instance of TaskFileRepository.
func NewTaskFileRepository(db *gorm.DB) TaskFileRepository {
	return &taskFileRepository{db: db}
}

// CreateTask creates a new task record in the database.
func (r *taskFileRepository) CreateTask(ctx context.Context, task *models.Task) error {
	result := r.db.WithContext(ctx).Create(task)
	if result.Error != nil {
		return common.Err("Failed to create task: %v", result.Error)
	}
	return nil
}

// GetTaskByID retrieves a task by its ID from the database.
func (r *taskFileRepository) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	result := r.db.WithContext(ctx).First(&task, "id = ?", taskID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, common.Err("Task not found: %v", result.Error)
		}
		return nil, common.Err("Failed to get task by ID: %v", result.Error)
	}
	return &task, nil
}

// CreateFile creates a new file record in the database.
func (r *taskFileRepository) CreateFile(ctx context.Context, file *models.File) error {
	result := r.db.WithContext(ctx).Create(file)
	if result.Error != nil {
		return common.Err("Failed to create file: %v", result.Error)
	}
	return nil
}

// GetFileByID retrieves a file by its ID from the database.
func (r *taskFileRepository) GetFileByID(ctx context.Context, fileID uuid.UUID) (*models.File, error) {
	var file models.File
	result := r.db.WithContext(ctx).First(&file, "id = ?", fileID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, common.Err("File not found: %v", result.Error)
		}
		return nil, common.Err("Failed to get file by ID: %v", result.Error)
	}
	return &file, nil
}