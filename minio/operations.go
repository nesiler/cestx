package minio

import (
	"context"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/nesiler/cestx/common"
)

// UploadTemplate uploads a template file to MinIO.
func UploadTemplate(ctx context.Context, client *minio.Client, filePath, templateName, bucketName string) (string, error) {
	// Open the template file
	file, err := os.Open(filePath)
	if err != nil {
		return "", common.Err("Failed to open template file '%s': %w", filePath, err)
	}
	defer file.Close()

	// Get file stat for size
	fileInfo, err := file.Stat()
	if err != nil {
		return "", common.Err("Failed to get file info for '%s': %w", filePath, err)
	}

	// Create upload info (Corrected - only 2 return values)
	uploadInfo, err := client.PutObject(ctx, bucketName, templateName, file, fileInfo.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return "", common.Err("Failed to upload template '%s': %w", templateName, err)
	}

	common.Ok("Template '%s' uploaded successfully. Upload Info: %s", templateName, uploadInfo)
	return templateName, nil // Return the object name for reference
}

// DownloadTemplate downloads a template file from MinIO.
func DownloadTemplate(ctx context.Context, client *minio.Client, objectName, localFilePath, bucketName string) error {
	// Create a local file to write to
	file, err := os.Create(localFilePath)
	if err != nil {
		return common.Err("Failed to create local file '%s': %w", localFilePath, err)
	}
	defer file.Close()

	// Get object from MinIO
	object, err := client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return common.Err("Failed to get object '%s' from MinIO: %w", objectName, err)
	}
	defer object.Close()

	// Copy the object content to the local file
	stat, err := object.Stat()
	if err != nil {
		return common.Err("Failed to get object stat: %w", err)
	}
	if _, err := io.CopyN(file, object, stat.Size); err != nil {
		return common.Err("Failed to download object '%s': %w", objectName, err)
	}

	common.Ok("Template '%s' downloaded successfully to '%s'", objectName, localFilePath)
	return nil
}
