package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nesiler/cestx/common"
)

// NewMinIOClient creates a new MinIO client and ensures the templates bucket exists.
func NewMinIOClient(cfg *Config) (*minio.Client, error) {
	// Create a new MinIO client instance
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, common.Err("Failed to create MinIO client: %w", err)
	}

	common.Ok("MinIO client created successfully.")

	// Check if the templates bucket exists
	ctx := context.Background() // Use a background context for this operation
	exists, err := client.BucketExists(ctx, cfg.TemplatesBucket)
	if err != nil {
		return nil, common.Err("Failed to check if bucket '%s' exists: %w", cfg.TemplatesBucket, err)
	}

	if !exists {
		// Create the bucket if it doesn't exist (adjust settings as needed)
		err = client.MakeBucket(ctx, cfg.TemplatesBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, common.Err("Failed to create bucket '%s': %w", cfg.TemplatesBucket, err)
		}
		common.Ok("Bucket '%s' created successfully.", cfg.TemplatesBucket)
	} else {
		common.Info("Bucket '%s' already exists.", cfg.TemplatesBucket)
	}

	return client, nil
}
