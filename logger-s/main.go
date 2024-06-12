package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/nesiler/cestx/common"
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

func healthCheck(service *common.ServiceConfig) {
	// Setup health check endpoint
	http.HandleFunc("/health", common.HealthHandler())

	// Start the server
	common.Info("Starting %v on port %d", service.Name, service.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", service.Port), nil); err != nil {
		common.Fatal("Server failed to start: %v\n", err)
	}
}