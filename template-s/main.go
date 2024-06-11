package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/nesiler/cestx/common"
)

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

func main() {
	godotenv.Load("../.env")
	godotenv.Load(".env")

	// Initialize clients and other setup
	InitializeClients()
	defer amqpConn.Close()

	// Register the service with the service registry
	go registerService()

	common.Head("Starting Template Service...")
	common.SendMessageToTelegram("**TEMPLATE SERVICE** ::: Service starting...")

	// Example usage (replace with your actual logic)
	if len(os.Args) > 2 {
		templateName := os.Args[1]
		filePath := os.Args[2]

		// Assuming you have a way to get the current user's ID
		userID, _ := uuid.NewRandom() // Replace with actual logic

		if err := UploadTemplate(templateName, filePath, userID); err != nil {
			log.Fatal(err)
		}
	} else {
		ConsumeMessages()
	}
}
