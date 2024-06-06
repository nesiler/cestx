package rabbitmq

import (
	"time"

	"github.com/google/uuid"
)

// ------------------------------------
// Common Events (Used across services)
// ------------------------------------

// CommonEvent represents common events in the system
type CommonEvent string

// Constants for different common events.
const (
	EventCreate CommonEvent = "create"
	EventDelete CommonEvent = "delete"
	EventUpdate CommonEvent = "update"
	EventStart  CommonEvent = "start"
	EventStop   CommonEvent = "stop"
	// TODO add more fields as needed
)

// ------------------------------------
// Machine Events and Messages
// ------------------------------------

// MachineEvent represents the type of event related to a machine.
type MachineEvent string

// Constants for different machine events.
const (
	MachineCreate MachineEvent = "machine.create"
	MachineDelete MachineEvent = "machine.delete"
	MachineStart  MachineEvent = "machine.start"
	MachineStop   MachineEvent = "machine.stop"
	MachineUpdate MachineEvent = "machine.update"
)

// MachineMessage represents a message related to a machine.
type MachineMessage struct {
	Event      MachineEvent `json:"event"`
	MachineID  uuid.UUID    `json:"machine_id"`
	TemplateID uuid.UUID    `json:"template_id,omitempty"`
	UserID     uuid.UUID    `json:"user_id,omitempty"`
	// TODO add more fields as needed
}

// ------------------------------------
// Template Events and Messages
// ------------------------------------

// TemplateEvent represents an event related to a template
type TemplateEvent string

// Constants for template events
const (
	TemplateCreate TemplateEvent = "template.create"
	TemplateDelete TemplateEvent = "template.delete"
	TemplateUpdate TemplateEvent = "template.update"
)

// TemplateMessage represents a message for template operations
type TemplateMessage struct {
	Event      TemplateEvent `json:"event"`
	TemplateID uuid.UUID     `json:"template_id"`
	Name       string        `json:"name,omitempty"`
	// TODO add more fields as needed
}

// ------------------------------------
// Dynoxy Events and Messages
// ------------------------------------

// DynoxyEvent represents an event for Dynoxy (adjust fields as needed)
type DynoxyEvent string

// Constants for dynoxy events
const (
	DynoxyCreate DynoxyEvent = "dynoxy.create"
	DynoxyDelete DynoxyEvent = "dynoxy.delete"
)

// DynoxyMessage represents a message for Dynoxy operations
type DynoxyMessage struct {
	Event     DynoxyEvent `json:"event"`
	RouteID   uuid.UUID   `json:"route_id"`
	MachineID uuid.UUID   `json:"machine_id"`
	UserID    uuid.UUID   `json:"user_id"`
	Port      int         `json:"port"`
	// TODO add more fields as needed
}

// ------------------------------------
// Taskmaster Events and Messages
// ------------------------------------

// TaskmasterTaskType represents the type of task for Taskmaster
type TaskmasterTaskType string

// Constants for Taskmaster task types
const (
	TaskmasterTaskAnsible TaskmasterTaskType = "ansible"
	TaskmasterTaskSSH     TaskmasterTaskType = "ssh"
	TaskmasterTaskScript  TaskmasterTaskType = "script"
	// TODO add more task types as needed
)

// TaskmasterMessage represents a message for Taskmaster
type TaskmasterMessage struct {
	TaskType  TaskmasterTaskType `json:"task_type"`
	MachineID uuid.UUID          `json:"machine_id"`
	// TODO: Add more fields as needed for different task types
}

// ------------------------------------
// Logger Service Message
// ------------------------------------

// LogMessage represents a log message
type LogMessage struct {
	Service   string    `json:"service"`
	Level     string    `json:"level"` // "info", "error", "debug"
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
