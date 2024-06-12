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
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/rabbitmq"
	"github.com/streadway/amqp"
)

var (
	amqpConn *amqp.Connection
	ctx      = context.Background()
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

	// 2. Initialize Clients (RabbitMQ)
	initClients()
	defer closeClients()

	// 4. Start Consuming Messages
	common.Head("Starting Dynoxy Service...")
	common.SendMessageToTelegram("**DYNOXY SERVICE** ::: Service starting...")
	consumeMessages()

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) // Capture interrupt signals
	<-quit                                             // Block until an interrupt is received

	common.Info("Dynoxy service stopping...")
	common.SendMessageToTelegram("**DYNOXY SERVICE** ::: Service stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-ctx.Done()
	common.Ok("Dynoxy service stopped successfully")
}

// initClients initializes the RabbitMQ client
func initClients() {
	var err error

	// Initialize RabbitMQ connection
	rabbitCfg := common.LoadRabbitMQConfig()
	amqpConn, err = rabbitmq.NewConnection(rabbitCfg)
	common.FailError(err, "Failed to initialize RabbitMQ connection: %v\n", err)
}

// closeClients closes the RabbitMQ connection
func closeClients() {
	// Close RabbitMQ connection gracefully
	if amqpConn != nil {
		if err := amqpConn.Close(); err != nil {
			common.Err("Error closing RabbitMQ connection: %v", err)
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

func consumeMessages() {
	// 1. Create a channel for message consumption
	ch, err := amqpConn.Channel()
	common.FailError(err, "Failed to open a channel: %v\n", err)
	defer ch.Close()

	// 2. Consume messages from the dynoxy queues
	err = rabbitmq.Consume(ch, rabbitmq.QueueDynoxyCreate, handleDynoxyCreate)
	common.FailError(err, "Failed to consume from 'dynoxy.create' queue: %w", err)

	err = rabbitmq.Consume(ch, rabbitmq.QueueDynoxyDelete, handleDynoxyDelete)
	common.FailError(err, "Failed to consume from 'dynoxy.delete' queue: %w", err)

	// Keep the service running to listen for messages
	forever := make(chan bool)
	<-forever
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
