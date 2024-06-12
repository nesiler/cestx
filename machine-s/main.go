package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	mc "github.com/minio/minio-go/v7"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/minio"
	"github.com/nesiler/cestx/postgresql"
	"github.com/nesiler/cestx/rabbitmq"
	"github.com/nesiler/cestx/redis"
	rc "github.com/redis/go-redis/v9"
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

	// Read the service configuration from service.json
	serviceData, err := os.ReadFile("service.json")
	common.FailError(err, "Failed to read service configuration: %v\n", err)

	// Load service configuration
	service, err := common.LoadServiceConfig(serviceData)
	common.FailError(err, "Failed to load service configuration: %v\n", err)

	// 3. Register Service
	go registerService(service)
	go healthCheck(service)

	// 2. Initialize Clients
	InitializeClients()
	defer closeClients() // Ensure graceful shutdown

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
	// Initialize Minio client
	minioCfg := common.LoadMinIOConfig()
	minioClient, _ = minio.NewMinIOClient(minioCfg)

	// Initialize PostgreSQL client
	dbCfg := common.LoadPostgreSQLConfig()
	postgresClient, _ = postgresql.NewPostgreSQLDB(dbCfg)

	// Initialize Redis client
	redisCfg := common.LoadRedisConfig()
	redisClient, _ = redis.NewRedisClient(redisCfg)

	// Initialize RabbitMQ connection
	rabbitCfg := common.LoadRabbitMQConfig()
	amqpConn, _ = rabbitmq.NewConnection(rabbitCfg)
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
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			common.Err("Failed to close Redis connection: %v", err)
		}
	}

}

// registerService registers the machine-s with the registry service.
func registerService(service *common.ServiceConfig) {
	var err error
	// Set the service address
	service.Address, err = common.ExternalIP()
	common.FailError(err, "Failed to get external IP: %v\n", err)

	common.REGISTRY_HOST = common.GetEnv("REGISTRY_HOST", "192.168.4.63")
	common.PYTHON_API_HOST = common.GetEnv("PYTHON_API_HOST", "192.168.4.99")

	// Register the service with the registry
	err = common.RegisterService(service)
	if err != nil {
		// Log the error, retry registration after a delay
		common.Warn("Failed to register service: %v", err)
		time.Sleep(5 * time.Second)
		registerService(service) // Retry registration
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

	// 3. Start consuming messages from the respective queues
	if err := rabbitmq.Consume(ch, rabbitmq.QueueMachineCreate, handleMachineCreate); err != nil {
		return fmt.Errorf("failed to consume from 'machine.create' queue: %w", err)
	}

	if err := rabbitmq.Consume(ch, rabbitmq.QueueMachineStart, handleMachineStart); err != nil {
		return fmt.Errorf("failed to consume from 'machine.start' queue: %w", err)
	}

	// 4. Keep the service running to listen for messages
	forever := make(chan bool)
	<-forever
	return nil
}

func healthCheck(service *common.ServiceConfig) {
	// Setup health check endpoint
	http.HandleFunc("/health", common.HealthHandler())

	// Start the server
	common.Info("Starting %v on port %d", service.Name, service.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", service.Port), nil); err != nil {
		common.Fatal("Server failed to start: %v\n", err)
	}
}
