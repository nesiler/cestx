package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/minio"
	"github.com/nesiler/cestx/postgresql"
	"github.com/nesiler/cestx/postgresql/models"
	"github.com/nesiler/cestx/rabbitmq"
	"github.com/nesiler/cestx/redis"
	amqp "github.com/streadway/amqp"
)

// CreateMachine handles the creation of a new machine.
func CreateMachine(msg rabbitmq.MachineMessage, amqpConn *amqp.Connection) error {
	ch, err := amqpConn.Channel()
	ctx := context.Background()
	// 1. Generate a unique machine name
	machineName := fmt.Sprintf("%s-%s", msg.TemplateID, uuid.New().String()[:8])

	// 2. Fetch Dockerfile path from Redis or PostgreSQL
	// Assuming you have a Redis key pattern like "template:{templateID}:filepath"
	redisKey := fmt.Sprintf("template:%s:filepath", msg.TemplateID)

	var dockerfilePath string
	err = redis.Get(ctx, redisClient, redisKey, &dockerfilePath)

	if err != nil {
		// If not found in Redis, fetch from PostgreSQL
		templateRepo := postgresql.NewTemplateRepository(postgresClient)
		template, err := templateRepo.GetTemplateByID(ctx, msg.TemplateID)
		if err != nil {
			return fmt.Errorf("failed to get template from PostgreSQL: %w", err)
		}

		dockerfilePath = template.Name
	}

	// 3. Download Dockerfile from Minio
	minioBucket := "templates"
	localDockerfilePath := fmt.Sprintf("/tmp/%s", dockerfilePath)

	if err := minio.DownloadTemplate(
		ctx,
		minioClient,
		dockerfilePath,
		localDockerfilePath,
		minioBucket,
	); err != nil {
		return fmt.Errorf("failed to download Dockerfile from Minio: %w", err)
	}

	// Ensure the file is removed after the function completes
	defer os.Remove(localDockerfilePath)

	// 4. Build Docker image
	imageName := fmt.Sprintf("cestx/%s", machineName)
	_, err = buildImage(localDockerfilePath, imageName)
	if err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	// 5. Run Docker container
	containerName := fmt.Sprintf("%s-%s", machineName, uuid.New().String()[:8])
	containerID, err := runContainer(imageName, containerName, []string{"80:80"}, "cpu=1", "memory=512m")
	if err != nil {
		return fmt.Errorf("failed to run Docker container: %w", err)
	}

	common.Info("Container ID: %s", containerID) // Log the container ID for reference

	randomPassword := "generated-password"

	// 7. Store machine details in PostgreSQL
	newMachine := &models.Machine{
		Name:       machineName,
		UserID:     msg.UserID,
		TemplateID: msg.TemplateID,
		Status:     true,
		Password:   randomPassword,
		ExpiresAt:  time.Now().Add(time.Hour * 1),                  // Default expiration: 1 hour
		URL:        fmt.Sprintf("%s.%s", containerID, "cestx.com"), // Update with your domain
	}

	machineRepo := postgresql.NewMachineRepository(postgresClient)
	if err := machineRepo.CreateMachine(ctx, newMachine); err != nil {
		return fmt.Errorf("failed to create machine record in PostgreSQL: %w", err)
	}

	// 8. Publish dynoxy.create message
	dynoxyMessage := rabbitmq.DynoxyMessage{
		Event:     rabbitmq.DynoxyCreate,
		RouteID:   uuid.New(),
		MachineID: newMachine.ID,
		UserID:    msg.UserID,
		Port:      80, // Assuming your app inside the container runs on port 80
	}

	// Publish the message to RabbitMQ for Dynoxy to create a route
	if err := rabbitmq.Publish(ch, rabbitmq.ExchangeDynoxy, rabbitmq.QueueDynoxyCreate, dynoxyMessage); err != nil {
		return fmt.Errorf("failed to publish dynoxy.create message: %w", err)
	}

	return nil
}

// StartMachine starts a stopped machine.
func StartMachine(machineID uuid.UUID, amqpConn *amqp.Connection) error {
	ctx := context.Background()
	// Retrieve machine details from PostgreSQL
	machineRepo := postgresql.NewMachineRepository(postgresClient)
	machine, err := machineRepo.GetMachineByID(ctx, machineID)
	if err != nil {
		return fmt.Errorf("failed to get machine from PostgreSQL: %w", err)
	}

	// Start the Docker container
	err = startContainer(machine.Name) // You'll likely use the container name or ID
	if err != nil {
		return fmt.Errorf("failed to start Docker container: %w", err)
	}

	// Update machine status in PostgreSQL
	machine.Status = true
	if err := machineRepo.UpdateMachine(ctx, machine); err != nil { // Assuming you have an UpdateMachine method
		return fmt.Errorf("failed to update machine status in PostgreSQL: %w", err)
	}

	return nil
}

// StopMachine stops a running machine.
func StopMachine(machineID uuid.UUID, amqpConn *amqp.Connection) error {
	ctx := context.Background()
	// Retrieve machine details from PostgreSQL
	machineRepo := postgresql.NewMachineRepository(postgresClient)
	machine, err := machineRepo.GetMachineByID(ctx, machineID)
	if err != nil {
		return fmt.Errorf("failed to get machine from PostgreSQL: %w", err)
	}

	// Stop the Docker container
	err = stopContainer(machine.Name)
	if err != nil {
		return fmt.Errorf("failed to stop Docker container: %w", err)
	}

	// Update machine status in PostgreSQL
	machine.Status = false
	if err := machineRepo.UpdateMachine(ctx, machine); err != nil {
		return fmt.Errorf("failed to update machine status in PostgreSQL: %w", err)
	}

	return nil
}

// DeleteMachine completely removes a machine.
func DeleteMachine(machineID uuid.UUID, amqpConn *amqp.Connection) error {
	ctx := context.Background()
	ch, err := amqpConn.Channel()
	// 1. Retrieve machine details from PostgreSQL
	machineRepo := postgresql.NewMachineRepository(postgresClient)
	machine, err := machineRepo.GetMachineByID(ctx, machineID)
	if err != nil {
		return fmt.Errorf("failed to get machine from PostgreSQL: %w", err)
	}

	// 2. Stop the Docker container (if it's running)
	if machine.Status {
		if err := stopContainer(machine.Name); err != nil {
			return fmt.Errorf("failed to stop Docker container: %w", err)
		}
	}

	// 3. Remove the Docker container
	if err := removeContainer(machine.Name); err != nil {
		return fmt.Errorf("failed to remove Docker container: %w", err)
	}

	// 4. Delete the machine record from PostgreSQL
	if err := machineRepo.DeleteMachine(ctx, machineID); err != nil {
		return fmt.Errorf("failed to delete machine record from PostgreSQL: %w", err)
	}

	// 5. Publish dynoxy.delete message
	dynoxyMessage := rabbitmq.DynoxyMessage{
		Event:     rabbitmq.DynoxyDelete,
		MachineID: machineID,
		// ... other necessary fields ...
	}

	if err := rabbitmq.Publish(ch, rabbitmq.ExchangeDynoxy, rabbitmq.QueueDynoxyDelete, dynoxyMessage); err != nil {
		// Handle the error (maybe log and proceed, or retry publishing)
		return fmt.Errorf("failed to publish dynoxy.delete message: %w", err)
	}

	return nil
}
