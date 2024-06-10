package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/nesiler/cestx/common"
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
	common.FailError(err, "")

	err = rdb.HSet(ctx, "service:"+service.ID, map[string]interface{}{
		"data":   serviceData,
		"status": "unknown",
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
	// TODO 1: Get the configurations with common package
	// TODO 2: Use these functions: common.LoadPostgreSQLConfig, common.LoadRabbitMQConfig, common.LoadRedisConfig
	// TODO 3: Return the configurations as JSON, use switch case for different configurations
}

func scheduleHealthCheck(service common.ServiceConfig) {
	// TODO: Schedule a health check for the service
}

func monitorService(service common.ServiceConfig) {
	// TODO: Monitor the service and update the status in Redis
}

func checkUp() {
	// TODO: Get all status of services and send a message to telegram
}
