package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/nesiler/cestx/common"
)

func main() {
	godotenv.Load("../.env")
	godotenv.Load(".env")
	// Load service configuration from file
	serviceFile, err := os.ReadFile("service.json")
	common.FailError(err, "error reading service file: %v", err)

	// Register the service with the registry
	common.REGISTRY_HOST = os.Getenv("REGISTRY_HOST")
	common.RegisterService(serviceFile)

	// Setup health check endpoint
	http.HandleFunc("/health", common.HealthHandler())

	// Start the server
	common.Info("Starting ExampleService on port 8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		common.Fatal("Server failed to start: %v\n", err)
	}
}
