package main

import (
	"net/http"
	"os"

	"github.com/nesiler/cestx/common"
)

func main() {
	// Load service configuration from file
	serviceFile, err := os.ReadFile("service.json")
	if err != nil {
		common.Fatal("Error reading service configuration file: %v\n", err)
	}

	// Register the service with the registry
	common.RegisterService(serviceFile)

	// Setup health check endpoint
	http.HandleFunc("/health", common.HealthHandler())
	ip, err := common.ExternalIP()
	if err != nil {
		common.Err(os.Stderr, "Error getting external IP address: %v\n", err)
	}

	common.Info("Health check service running on http://%s:3333/health\n", ip)

	if err := http.ListenAndServe(":3333", nil); err != nil {
		common.Fatal("Error starting health check service: %v\n", err)
	}
}
