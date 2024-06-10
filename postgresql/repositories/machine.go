package postgresql

import (
	"context"

	"github.com/google/uuid"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/postgresql/models"
	"gorm.io/gorm"
)

// MachineRepository defines methods for interacting with Machine entities.
type MachineRepository interface {
	CreateMachine(ctx context.Context, machine *models.Machine) error
	GetMachineByID(ctx context.Context, machineID uuid.UUID) (*models.Machine, error)
}

type machineRepository struct {
	db *gorm.DB
}

// NewMachineRepository creates a new instance of MachineRepository.
func NewMachineRepository(db *gorm.DB) MachineRepository {
	return &machineRepository{db: db}
}

// CreateMachine creates a new machine record in the database.
func (r *machineRepository) CreateMachine(ctx context.Context, machine *models.Machine) error {
	result := r.db.WithContext(ctx).Create(machine)
	if result.Error != nil {
		return common.Err("Failed to create machine: %v", result.Error)
	}
	return nil
}

// GetMachineByID retrieves a machine by its ID from the database.
func (r *machineRepository) GetMachineByID(ctx context.Context, machineID uuid.UUID) (*models.Machine, error) {
	var machine models.Machine
	result := r.db.WithContext(ctx).First(&machine, "id = ?", machineID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, common.Err("Machine not found: %v", result.Error)
		}
		return nil, common.Err("Failed to get machine by ID: %v", result.Error)
	}
	return &machine, nil
}