package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	mc "github.com/minio/minio-go/v7"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/minio"
	"github.com/nesiler/cestx/postgresql"
	"github.com/nesiler/cestx/postgresql/models"
	"github.com/nesiler/cestx/rabbitmq"
	amqp "github.com/streadway/amqp"
	gc "gorm.io/gorm"
)

type Template struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"uniqueIndex"`
	Description string    `gorm:"type:text"`
	Type        string    `gorm:"type:varchar(20);default:'dockerfile'"`
	File        string    `gorm:"type:text"`
	UserID      uuid.UUID `gorm:"type:uuid"`
}

var (
	minioClient    *mc.Client
	postgresClient *gc.DB
	amqpConn       *amqp.Connection
	amqpChannel    *amqp.Channel
)

// InitializeClients initializes the Minio, PostgreSQL and RabbitMQ clients.
func InitializeClients() {
	// Initialize Minio client
	minioCfg := common.LoadMinIOConfig()
	minioClient, _ = minio.NewMinIOClient(minioCfg)

	// Initialize PostgreSQL client
	dbCfg := common.LoadPostgreSQLConfig()
	postgresClient, _ = postgresql.NewPostgreSQLDB(dbCfg)

	// Initialize RabbitMQ connection
	rabbitCfg := common.LoadRabbitMQConfig()
	amqpConn, _ = rabbitmq.NewConnection(rabbitCfg)

	// Initialize RabbitMQ channel
	amqpChannel, _ = amqpConn.Channel()
}

// UploadTemplate uploads a Dockerfile to Minio and creates a template record in PostgreSQL.
func UploadTemplate(templateName, filePath string, userID uuid.UUID) error {
	ctx := context.Background()
	bucketName := "templates"

	// 1. Upload Dockerfile to Minio
	objectName, err := minio.UploadTemplate(ctx, minioClient, filePath, templateName, bucketName)
	if err != nil {
		return fmt.Errorf("failed to upload Dockerfile to Minio: %w", err)
	}

	// 2. Create template record in PostgreSQL
	newTemplate := &models.Template{
		Name:        templateName,
		Description: "Uploaded by template service",
		Type:        "dockerfile",
		File:        objectName,
		UserID:      userID,
	}

	templateRepo := postgresql.NewTemplateRepository(postgresClient)
	if err := templateRepo.CreateTemplate(ctx, newTemplate); err != nil {
		return fmt.Errorf("failed to create template record: %w", err)
	}

	common.Ok("Template uploaded and registered successfully: %s", templateName)
	return nil
}

// DeleteTemplate deletes a template from both Minio and PostgreSQL.
func DeleteTemplate(templateName string) error {
	ctx := context.Background()
	bucketName := "templates"

	// 1. Get template information from PostgreSQL
	templateRepo := postgresql.NewTemplateRepository(postgresClient)
	template, err := templateRepo.GetTemplateByName(ctx, templateName)
	if err != nil {
		return fmt.Errorf("failed to get template from PostgreSQL: %w", err)
	}

	// 2. Delete from Minio
	if err := minio.DeleteTemplate(ctx, minioClient, template.Name, bucketName); err != nil {
		return fmt.Errorf("failed to delete template from Minio: %w", err)
	}

	// 3. Delete from PostgreSQL
	if err := templateRepo.DeleteTemplate(ctx, template.ID); err != nil {
		return fmt.Errorf("failed to delete template record: %w", err)
	}

	common.Ok("Template deleted successfully: %s", templateName)
	return nil
}

// handleMessage processes RabbitMQ messages.
func handleMessage(delivery amqp.Delivery) {
	var templateMessage rabbitmq.TemplateMessage
	if err := json.Unmarshal(delivery.Body, &templateMessage); err != nil {
		common.Err("Error unmarshalling message: %v", err)
		// Handle the error (e.g., nack the message)
		return
	}

	common.Info("Received message: %+v", templateMessage)

	// TODO: You'll likely need a TemplateID in the message to identify the template

	// Implement logic for different template events (create, delete, ...)
	switch templateMessage.Event {
	case rabbitmq.TemplateCreate:
		// Call UploadTemplate function here
	case rabbitmq.TemplateDelete:
		// Call DeleteTemplate function here
	// ... handle other events
	default:
		common.Warn("Unknown template event: %s", templateMessage.Event)
	}
}

// ConsumeMessages starts consuming RabbitMQ messages for template operations.
func ConsumeMessages() {
	// Start consuming messages
	err := rabbitmq.Consume(amqpChannel, rabbitmq.QueueTemplateCreate, func(delivery amqp.Delivery) error {
		handleMessage(delivery)
		return nil // Acknowledge message if processed successfully; otherwise return an error
	})
	if err != nil {
		common.Fatal("Error consuming messages: %v", err)
	}
}
