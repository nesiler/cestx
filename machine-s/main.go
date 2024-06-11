package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	rc "github.com/redis/go-redis/v9"
	"github.com/joho/godotenv"
	mc "github.com/minio/minio-go/v7"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/minio"
	"github.com/nesiler/cestx/postgresql"
	"github.com/nesiler/cestx/rabbitmq"
	"github.com/nesiler/cestx/redis"
	amqp "github.com/streadway/amqp"
	gc "gorm.io/gorm"
)

// Declare global variables for clients
var (
	minioClient    *mc.Client
	postgresClient *gc.DB
	redisClient    *rc.Client
	amqpConn       *amqp.Connection
)

func main() {
	// 1. Load Environment Variables
	godotenv.Load("../.env")
	godotenv.Load(".env")

	common.PYTHON_API_HOST = common.GetEnv("PYTHON_API_HOST", "")
	common.TELEGRAM_TOKEN = common.GetEnv("TELEGRAM_TOKEN", "")
	common.CHAT_ID = common.GetEnv("CHAT_ID", "")
	common.REGISTRY_HOST = common.GetEnv("REGISTRY_HOST", "")

	// 2. Initialize Clients
	InitializeClients()
	defer closeClients() // Ensure graceful shutdown

	// 3. Register Service
	go registerService()

	// 4. Start Consuming Messages
	common.Head("Starting Machine Service...")
	common.SendMessageToTelegram("**MACHINE SERVICE** ::: Service starting...")

	if err := consumeMessages(); err != nil {
		common.Fatal("Error consuming messages: %v", err)
	}

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) // Capture interrupt signals
	<-quit                                             // Block until an interrupt is received

	common.Info("Machine service stopping...")
	common.SendMessageToTelegram("**MACHINE SERVICE** ::: Service stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-ctx.Done()
	common.Ok("Machine service stopped successfully")
}

// InitializeClients sets up connections to external services.
func InitializeClients() {
	var err error

	// Initialize Minio client
	minioCfg := common.LoadMinIOConfig()
	minioClient, err = minio.NewMinIOClient(minioCfg)
	common.FailError(err, "Failed to initialize Minio client: %v\n", err)

	// Initialize PostgreSQL client
	dbCfg := common.LoadPostgreSQLConfig()
	postgresClient, err = postgresql.NewPostgreSQLDB(dbCfg)
	common.FailError(err, "Failed to initialize PostgreSQL client: %v\n", err)

	// Initialize Redis client
	redisCfg := common.LoadRedisConfig()
	redisClient, err = redis.NewRedisClient(redisCfg)
	common.FailError(err, "Failed to initialize Redis client: %v\n", err)

	// Initialize RabbitMQ connection
	rabbitCfg := common.LoadRabbitMQConfig()
	amqpConn, err = rabbitmq.NewConnection(rabbitCfg)
	common.FailError(err, "Failed to initialize RabbitMQ connection: %v\n", err)
}

// closeClients closes connections to external services.
func closeClients() {
	// Close RabbitMQ connection gracefully
	if amqpConn != nil {
		if err := amqpConn.Close(); err != nil {
			common.Err("Error closing RabbitMQ connection: %v", err)
		}
	}

	// Close Redis connection
	if redisClient != nil && redisClient != nil {
		if err := redisClient.Close(); err != nil {
			common.Err("Failed to close Redis connection: %v", err)
		}
	}

	// (Optional) Close other client connections if needed
}

// registerService registers the machine-s with the registry service.
func registerService() {
	// Read the service configuration from service.json
	serviceData, err := os.ReadFile("service.json")
	common.FailError(err, "Failed to read service configuration: %v\n", err)

	// Load service configuration
	service, err := common.LoadServiceConfig(serviceData)
	common.FailError(err, "Failed to load service configuration: %v\n", err)

	// Set the service address
	service.Address, err = common.ExternalIP()
	common.FailError(err, "Failed to get external IP: %v\n", err)

	// Register the service with the registry
	err = common.RegisterService(service)
	if err != nil {
		// Log the error, retry registration after a delay
		common.Warn("Failed to register service: %v", err)
		time.Sleep(5 * time.Second)
		registerService() // Retry registration
		return
	}

	// Send a Telegram message on successful registration
	common.SendMessageToTelegram(fmt.Sprintf("**%s** ::: Service registered successfully!", service.Name))
}

// consumeMessages sets up consumers for different machine events.
func consumeMessages() error {
	// 1. Create a channel for message consumption
	ch, err := amqpConn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close() // Close the channel when the function exits

	// 2. Define consumer functions for different events (using closures to capture amqpConn)
	handleMachineCreate := func(delivery amqp.Delivery) error {
		return handleMessage(delivery, amqpConn)
	}

	handleMachineStart := func(delivery amqp.Delivery) error {
		return handleMessage(delivery, amqpConn)
	}

	// ... Define similar consumer functions for handleMachineStop, handleMachineDelete, etc.

	// 3. Start consuming messages from the respective queues
	if err := rabbitmq.Consume(ch, rabbitmq.QueueMachineCreate, handleMachineCreate); err != nil {
		return fmt.Errorf("failed to consume from 'machine.create' queue: %w", err)
	}

	if err := rabbitmq.Consume(ch, rabbitmq.QueueMachineStart, handleMachineStart); err != nil {
		return fmt.Errorf("failed to consume from 'machine.start' queue: %w", err)
	}

	// ... Start consuming from other queues: QueueMachineStop, QueueMachineDelete, etc.

	// 4. Keep the service running to listen for messages
	forever := make(chan bool)
	<-forever
	return nil
}
