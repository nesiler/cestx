package rabbitmq

import (
	"github.com/nesiler/cestx/common"
)

// Config holds the configuration for RabbitMQ
type Config struct {
	URL      string
	Username string
	Password string
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		URL:      common.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Username: common.GetEnv("RABBITMQ_USERNAME", "guest"),
		Password: common.GetEnv("RABBITMQ_PASSWORD", "guest"),
	}
}
