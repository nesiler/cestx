package postgresql

import (
	"github.com/nesiler/cestx/common"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// LoadConfig loads the PostgreSQL database configuration from environment variables.
func LoadConfig() *Config {
	return &Config{
		Host:     common.GetEnv("DB_HOST", "localhost"),
		Port:     common.GetEnv("DB_PORT", "5432"),
		User:     common.GetEnv("DB_USER", "postgres"),
		Password: common.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   common.GetEnv("DB_NAME", "cestx"),
	}
}
