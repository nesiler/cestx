package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/nesiler/cestx/common"
)

func main() {
	godotenv.Load("../.env")
	godotenv.Load(".env")

	common.PYTHON_API_HOST = os.Getenv("PYTHON_API_HOST")
	if common.PYTHON_API_HOST == "" {
		common.Warn("PYTHON_API_HOST not set, using default value")
		common.PYTHON_API_HOST = "http://192.168.4.99"
	}
	common.SendMessageToTelegram("**MACHINE** ::: Service started")

	// Load service configuration from file
	serviceFile, err := os.ReadFile("service.json")
	common.FailError(err, "error reading service file: %v", err)
	service, err := common.JsonToService(serviceFile)
	common.FailError(err, "error converting JSON to Service: %v", err)

	if service.Address == "" {
		service.Address, err = common.ExternalIP()
		common.FailError(err, "error finding external IP: %v", err)
	}

	// Register the service with the registry
	common.REGISTRY_HOST = os.Getenv("REGISTRY_HOST")
	err = common.RegisterService(service)
	common.Warn("Failed to register service: %v\n", err)

	// Setup health check endpoint
	http.HandleFunc("/health", common.HealthHandler())

	// Start the server
	common.Info("Starting %v on port %d", service.Name, service.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", service.Port), nil); err != nil {
		common.Fatal("Server failed to start: %v\n", err)
	}
}
