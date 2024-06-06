package redis

import (
	"time"

	"github.com/google/uuid"
)

// ------------------------------------
// Constants for Keys (Promotes Consistency)
// ------------------------------------

const (
	KeyServicePrefix = "service:" // Use a prefix for service keys
	KeySessionPrefix = "session:" // Use a prefix for session keys
)

// ------------------------------------
// Machine Session
// ------------------------------------

// MachineSession represents an active machine session
type MachineSession struct {
	SessionID uuid.UUID `json:"session_id"`
	MachineID uuid.UUID `json:"machine_id"`
	UserID    uuid.UUID `json:"user_id"`
	URL       string    `json:"url"`
	Port      int       `json:"port"`
	StartedAt time.Time `json:"started_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ------------------------------------
// Service Registry
// ------------------------------------

// HealthCheck represents a service's health check configuration.
type HealthCheck struct {
	Endpoint string `json:"endpoint"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}

// Service represents a registered microservice
type Service struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Address     string      `json:"address"`
	Port        int         `json:"port"`
	HealthCheck HealthCheck `json:"healthCheck"`
}

// ExternalService represents configuration for an external service
type ExternalService struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user,omitempty"` // Optional fields for credentials
	Password string `json:"password,omitempty"`
	DBName   string `json:"dbname,omitempty"`
}

// ExternalServicesConfig holds configurations for external services
type ExternalServicesConfig struct {
	Registry   ExternalService `json:"registry"`
	Redis      ExternalService `json:"redis"`
	PostgreSQL ExternalService `json:"postgresql"`
	RabbitMQ   ExternalService `json:"rabbitmq"`
}
