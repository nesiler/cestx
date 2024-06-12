package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nesiler/cestx/common"
	"github.com/redis/go-redis/v9"
)

func registerServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var service common.ServiceConfig
	err := json.NewDecoder(r.Body).Decode(&service)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	registerService(service)
	scheduleHealthCheck(service)

	w.WriteHeader(http.StatusOK)
}

func registerService(service common.ServiceConfig) {
	common.Info("Registering service: %s", service.Name)
	common.SendMessageToTelegram("**REGISTRY** ::: Registering service: " + service.Name)

	serviceData, err := json.Marshal(service)
	if err != nil {
		common.Err("Failed to marshal service data: %v", err)
		return
	}

	common.Warn("Service data: %s", serviceData)

	err = rdb.HSet(ctx, "service:"+service.ID, map[string]interface{}{
		"data":   serviceData,
		"status": "unknown", // Initial status is unknown
	}).Err()
	common.FailError(err, "Redis error: %v")
}

func getServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Path[len("/service/"):]

	serviceData, err := rdb.HGet(ctx, "service:"+serviceID, "data").Result()
	if err == redis.Nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(serviceData))
}

func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	configType := r.URL.Path[len("/config/"):]

	var configData []byte
	var err error

	switch configType {
	case "postgresql":
		configData, err = json.Marshal(common.LoadPostgreSQLConfig())
	case "rabbitmq":
		configData, err = json.Marshal(common.LoadRabbitMQConfig())
	case "redis":
		configData, err = json.Marshal(common.LoadRedisConfig())
	default:
		http.Error(w, "Invalid config type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Error fetching config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(configData)
}

func scheduleHealthCheck(service common.ServiceConfig) {
	common.Info("Scheduling health check for service: %s", service.Name)

	// Define the health check function
	healthCheckFunc := func() {
		// Construct the health check URL
		healthCheckURL := fmt.Sprintf("http://%s:%d%s", service.Address, service.Port, service.HealthCheck.Endpoint)
		common.Warn("Health check URL: %s", healthCheckURL)

		// Perform the health check request
		resp, err := http.Get(healthCheckURL)
		if err != nil {
			common.Warn("Health check failed for service %s: %v", service.Name, err)
			// Update service status in Redis to "unhealthy"
			updateServiceStatus(service.ID, "unhealthy")
			return
		}
		defer resp.Body.Close()

		// Check for a successful status code (200-299)
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			common.Info("Health check successful for service: %s", service.Name)
			updateServiceStatus(service.ID, "healthy")
		} else {
			common.Warn("Health check failed for service %s: Status Code %d", service.Name, resp.StatusCode)
			updateServiceStatus(service.ID, "unhealthy")
		}
	}

	// Schedule the health check to run every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for {
			<-ticker.C
			healthCheckFunc()
		}
	}()
}

func updateServiceStatus(serviceID, status string) {
	err := rdb.HSet(ctx, "service:"+serviceID, "status", status).Err()
	if err != nil {
		common.Err("Failed to update service status in Redis: %v", err)
	}
}

func checkUp() {
	common.Info("Checking up services...")

	// Iterate through all registered services in Redis
	iter := rdb.Scan(ctx, 0, "service:*", 0).Iterator()
	for iter.Next(ctx) {
		serviceKey := iter.Val()

		// Retrieve service data from Redis
		serviceData, err := rdb.HGet(ctx, serviceKey, "data").Result()
		if err != nil {
			common.Err("Failed to get service data from Redis: %v", err)
			continue // Skip to the next service
		}

		var service common.ServiceConfig
		err = json.Unmarshal([]byte(serviceData), &service)
		if err != nil {
			common.Err("Failed to unmarshal service data: %v", err)
			continue
		}

		// Get service status
		status, err := rdb.HGet(ctx, serviceKey, "status").Result()
		if err != nil {
			common.Err("Failed to get service status: %v", err)
			continue
		}

		// Log the status of each service
		common.Out("Service: %s (%s) - Status: %s", service.Name, service.ID, status)

		// Optionally, send a Telegram message if a service is unhealthy
		if status == "unhealthy" {
			message := fmt.Sprintf("**REGISTRY WARNING**\nService: %s (%s) is unhealthy!", service.Name, service.ID)
			common.SendMessageToTelegram(message)
		}
	}

	if err := iter.Err(); err != nil {
		common.Err("Error iterating through services in Redis: %v", err)
	}
}
