package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/minio/minio-go"
	"github.com/nesiler/cestx/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Machine struct representing a machine instance.
type Machine struct {
	common.Base
	UserID      uuid.UUID `gorm:"type:uuid"`
	TemplateID  uuid.UUID `gorm:"type:uuid"`
	Name        string    `gorm:"uniqueIndex"`
	Status      string
	ContainerID string
	IP          string
	URL         string
	Port        int
}

func main() {
	// Load configurations from environment variables
	common.TELEGRAM_TOKEN = common.GetEnv("TELEGRAM_TOKEN", "")
	common.CHAT_ID = common.GetEnv("CHAT_ID", "")
	common.PYTHON_API_HOST = common.GetEnv("PYTHON_API_HOST", "localhost")
	common.REGISTRY_HOST = common.GetEnv("REGISTRY_HOST", "localhost:3434")
	rabbitMQConfig := &common.RabbitMQConfig{
		URL:      common.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Username: common.GetEnv("RABBITMQ_USERNAME", "guest"),
		Password: common.GetEnv("RABBITMQ_PASSWORD", "guest"),
	}
	postgresConfig := &common.PostgresConfig{
		Host:     common.GetEnv("DB_HOST", "localhost"),
		Port:     common.GetEnv("DB_PORT", "5432"),
		User:     common.GetEnv("DB_USER", "postgres"),
		Password: common.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   common.GetEnv("DB_NAME", "cestx"),
	}
	minioCfg := &common.MinIOConfig{
		Endpoint:        common.GetEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKeyID:     common.GetEnv("MINIO_ACCESS_KEY_ID", "admin"),
		SecretAccessKey: common.GetEnv("MINIO_SECRET_ACCESS_KEY", "password"),
		UseSSL:          common.GetEnvAsBool("MINIO_USE_SSL", false),
		TemplatesBucket: common.GetEnv("MINIO_TEMPLATES_BUCKET", "templates"),
	}

	// Initialize database connection
	db, err := initDatabase(postgresConfig)
	if err != nil {
		common.Fatal("Failed to connect to database: %s", err)
	}
	defer func(db *gorm.DB) {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}(db)

	// Initialize MinIO client
	minioClient, err := common.NewMinIOClient(minioCfg)
	if err != nil {
		common.Fatal("Failed to initialize MinIO client: %s", err)
	}

	// Initialize Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		common.Fatal("Failed to initialize Docker client: %s", err)
	}

	// Initialize RabbitMQ connection
	rabbitMQConn, err := common.NewRabbitMQConnection(rabbitMQConfig)
	if err != nil {
		common.Fatal("Failed to connect to RabbitMQ: %s", err)
	}
	defer func(conn *common.RabbitMQConnection) {
		_ = conn.Close()
	}(rabbitMQConn)

	// Start consuming from the machine.create queue
	messages, err := rabbitMQConn.Channel.Consume(
		"machine.tasks.create", // queue name
		"",                     // consumer
		false,                  // auto-ack
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		common.Fatal("Failed to register a consumer: %s", err)
	}

	forever := make(chan bool)

	go func() {
		for msg := range messages {
			var machineMsg common.MachineMessage
			if err := msg.Unmarshal(machineMsg); err != nil {
				// Handle unmarshaling error (log, nack, etc.)
				common.Err("Failed to unmarshal message: %s", err)
				continue // Skip to the next message
			}
			common.Head("Received a message: %s", machineMsg)

			// Create the machine
			if err := createMachine(db, dockerClient, minioClient, minioCfg.TemplatesBucket, &machineMsg); err != nil {
				// Handle machine creation error (log, nack, etc.)
				common.Err("Failed to create machine: %s", err)
				continue // Skip to the next message
			}

			if err := msg.Ack(false); err != nil {
				common.Err("Failed to acknowledge message: %s", err)
			}
		}
	}()

	common.Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func createMachine(db *gorm.DB, dockerClient *client.Client, minioClient *minio.Client, minioBucket string, msg *common.MachineMessage) error {
	ctx := context.Background()

	// 1. Get the template file from MinIO
	templateFileName := msg.TemplateID.String() + ".zip"
	templateFilePath := "/tmp/" + templateFileName // Or a suitable temporary location
	err := common.DownloadTemplate(ctx, minioClient, templateFileName, templateFilePath, minioBucket)
	if err != nil {
		return fmt.Errorf("failed to download template: %w", err)
	}

	// 2. Create the machine (Implement container creation logic using dockerClient)
	containerID, machineIP, err := createContainerFromTemplate(ctx, dockerClient, templateFilePath)
	if err != nil {
		return fmt.Errorf("failed to create container from template: %w", err)
	}

	// 3. Store machine information in the database
	newMachine := &Machine{
		UserID:      msg.UserID,
		TemplateID:  msg.TemplateID,
		Name:        "Machine-" + msg.MachineID.String(), // You might want to generate this differently
		Status:      "running",                           // Or another appropriate initial status
		ContainerID: containerID,
		IP:          machineIP,
		URL:         "test",
		Port:        22, // Default SSH port, adjust as needed
	}
	if err := db.Create(&newMachine).Error; err != nil {
		return fmt.Errorf("failed to save machine to database: %w", err)
	}

	// 4. Send message to Dynoxy to set up the proxy (implementation will be added later)
	if err := sendDynoxySetupMessage(msg.MachineID); err != nil {
		return fmt.Errorf("failed to send Dynoxy setup message: %w", err)
	}

	return nil
}

func createContainerFromTemplate(ctx context.Context, dockerClient *client.Client, templatePath string) (string, string, error) {

	// 1. Extract the template (assuming it's a zip file)
	tempDir, err := os.MkdirTemp("", "cestx-template-*")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up the temporary directory

	if err := common.Unzip(templatePath, tempDir); err != nil {
		return "", "", fmt.Errorf("failed to unzip template: %w", err)
	}

	// 2. Build Docker image (if a Dockerfile is present)
	dockerfilePath := filepath.Join(tempDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil { // Check if Dockerfile exists
		// Image build context is the directory containing the Dockerfile
		buildCtx, err := archive.TarWithOptions(tempDir, &archive.TarOptions{})
		if err != nil {
			return "", "", fmt.Errorf("failed to create build context: %w", err)
		}
		defer buildCtx.Close()

		// Build the image
		imageBuildResponse, err := dockerClient.ImageBuild(
			ctx,
			buildCtx,
			types.ImageBuildOptions{
				Dockerfile: "Dockerfile",
				Tags:       []string{"cestx-machine-image:" + uuid.New().String()}, // Tag the image
				Remove:     true,                                                   // Remove intermediate containers
			},
		)
		if err != nil {
			return "", "", fmt.Errorf("failed to build Docker image: %w", err)
		}
		defer imageBuildResponse.Body.Close()
	}

	// 3. Create the container (using Sysbox as the runtime)
	containerConfig := &container.Config{
		Image:        "cestx-machine-image:" + uuid.New().String(), // Use the newly built image
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		OpenStdin:    true,
	}

	hostConfig := &container.HostConfig{
		// Port bindings, volume mounts, and other host-level configurations can go here.
		// ...
		Runtime: "sysbox-runc", // Specify Sysbox as the runtime
	}

	containerCreateResponse, err := dockerClient.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		nil,
		nil,
		"cestx-machine-"+uuid.New().String(), // Unique container name
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to create container: %w", err)
	}

	// 4. Start the container
	if err := dockerClient.ContainerStart(ctx, containerCreateResponse.ID, types.ContainerStartOptions{}); err != nil {
		return "", "", fmt.Errorf("failed to start container: %w", err)
	}

	// 5. Get the container's IP address (example for a bridge network)
	containerJSON, err := dockerClient.ContainerInspect(ctx, containerCreateResponse.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to inspect container: %w", err)
	}

	containerIP := containerJSON.NetworkSettings.Networks["bridge"].IPAddress // Adjust network name as needed

	return containerCreateResponse.ID, containerIP, nil
}

func sendDynoxySetupMessage(machineID uuid.UUID) error {
	// TODO: Implement RabbitMQ message sending logic to notify Dynoxy about the new machine.
	return nil
}

func initDatabase(cfg *common.PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
