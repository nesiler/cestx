package main

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/nesiler/cestx/common" 
    "gorm.io/gorm"
)

// ... (Your Machine struct from main.go) ...

// CreateMachine creates a new machine record in the database.
func CreateMachine(db *gorm.DB, machine *Machine) error {
    if err := db.Create(machine).Error; err != nil {
        return common.Err("Failed to create machine: %w", err)
    }
    return nil
}

// GetMachineByID retrieves a machine by its ID.
func GetMachineByID(db *gorm.DB, machineID uuid.UUID) (*Machine, error) {
    var machine Machine
    if err := db.First(&machine, "id = ?", machineID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, common.Err("Machine not found: %w", err) 
        }
        return nil, common.Err("Failed to get machine by ID: %w", err)
    }
    return &machine, nil
}

// UpdateMachine updates an existing machine record.
func UpdateMachine(db *gorm.DB, machine *Machine) error { 
    if err := db.Save(machine).Error; err != nil {
        return common.Err("Failed to update machine: %w", err) 
    }
    return nil
}

// DeleteMachine deletes a machine record by its ID.
func DeleteMachine(db *gorm.DB, machineID uuid.UUID) error {
    if err := db.Delete(&Machine{}, "id = ?", machineID).Error; err != nil {
        return common.Err("Failed to delete machine: %w", err) 
    }
    return nil
}