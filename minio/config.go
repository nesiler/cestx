package minio

import (
	"github.com/nesiler/cestx/common"
)

// Config holds the configuration for MinIO.
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	TemplatesBucket string // Add bucket name to config
}

// LoadConfig loads the configuration from environment variables.
func LoadConfig() *Config {
	return &Config{
		Endpoint:        common.GetEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKeyID:     common.GetEnv("MINIO_ACCESS_KEY_ID", "admin"),
		SecretAccessKey: common.GetEnv("MINIO_SECRET_ACCESS_KEY", "password"),
		UseSSL:          common.GetEnvAsBool("MINIO_USE_SSL", false),
		TemplatesBucket: common.GetEnv("MINIO_TEMPLATES_BUCKET", "templates"), // Load bucket name
	}
}
