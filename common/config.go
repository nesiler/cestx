package common

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// It returns the value if found, otherwise the provided default value.
func GetEnv(key string, defaultValue string) string {
	godotenv.Load("../.env")
	godotenv.Load(".env")
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// It returns the value if found, otherwise the provided default value.
func GetEnvAsBool(key string, defaultValue bool) bool {
	godotenv.Load("../.env")
	godotenv.Load(".env")
	if value, exists := os.LookupEnv(key); exists {
		return value == "true"
	}
	return defaultValue
}

// It returns the value if found, otherwise the provided default value.
func GetEnvAsInt(key string, defaultValue int) int {
	godotenv.Load("../.env")
	godotenv.Load(".env")
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Config holds the configuration for MinIO.
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	TemplatesBucket string // Add bucket name to config
}

// LoadConfig loads the configuration from environment variables.
func LoadMinIOConfig() *MinIOConfig {
	return &MinIOConfig{
		Endpoint:        GetEnv("MINIO_ENDPOINT", "192.168.4.70:9000"),
		AccessKeyID:     GetEnv("MINIO_ACCESS_KEY_ID", "admin"),
		SecretAccessKey: GetEnv("MINIO_SECRET_ACCESS_KEY", "password"),
		UseSSL:          GetEnvAsBool("MINIO_USE_SSL", false),
		TemplatesBucket: GetEnv("MINIO_TEMPLATES_BUCKET", "templates"), // Load bucket name
	}
}

type ServiceConfig struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Address     string      `json:"address"`
	Port        int         `json:"port"`
	HealthCheck HealthCheck `json:"healthCheck"`
}

func LoadServiceConfig(jsonData []byte) (*ServiceConfig, error) {
	var service ServiceConfig
	err := json.Unmarshal(jsonData, &service)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON data: %v", err)
	}
	return &service, nil
}

type HealthCheck struct {
	Endpoint string `json:"endpoint"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}

type PostgreSQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// LoadConfig loads the PostgreSQL database configuration from environment variables.
func LoadPostgreSQLConfig() *PostgreSQLConfig {
	return &PostgreSQLConfig{
		Host:     GetEnv("DB_HOST", "192.168.4.61"),
		Port:     GetEnv("DB_PORT", "5432"),
		User:     GetEnv("DB_USER", "postgres"),
		Password: GetEnv("DB_PASSWORD", "postgres"),
		DBName:   GetEnv("DB_NAME", "postgres"),
	}
}

// Config holds the configuration for RabbitMQ
type RabbitMQConfig struct {
	Host     string
	Username string
	Password string
}

// LoadConfig loads the configuration from environment variables
func LoadRabbitMQConfig() *RabbitMQConfig {
	return &RabbitMQConfig{
		Host:     GetEnv("RABBITMQ_URL", "192.168.4.62"),
		Username: GetEnv("RABBITMQ_USERNAME", "admin"),
		Password: GetEnv("RABBITMQ_PASSWORD", "password"),
	}
}

type RedisConfig struct {
	Host string
	Port string
}

// LoadConfig loads the Redis configuration from environment variables.
func LoadRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host: GetEnv("REDIS_HOST", "192.168.4.60"),
		Port: GetEnv("REDIS_PORT", "6379"),
	}
}
