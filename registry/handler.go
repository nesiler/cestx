package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nesiler/cestx/common"
)

func registerServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var service common.Service
	err := json.NewDecoder(r.Body).Decode(&service)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	registerService(service)
	scheduleHealthCheck(service)
	w.WriteHeader(http.StatusOK)
}

func registerService(service common.Service) {
	common.Info("Registering service: %s", service.Name)

	serviceData, err := json.Marshal(service)
	common.FailError(err, "")

	err = rdb.Set(ctx, "service:"+service.ID, serviceData, 0).Err()
	common.FailError(err, "Redis error: %v")
}

func getServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Path[len("/service/"):]

	serviceData, err := rdb.Get(ctx, "service:"+serviceID).Result()
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
	configName := r.URL.Path[len("/config/"):]

	serviceInfo, ok := configData.ExternalServices[configName]
	if !ok {
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}

	serviceInfoData, err := json.Marshal(serviceInfo)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(serviceInfoData)
}

func scheduleHealthCheck(service common.Service) {
	interval, err := time.ParseDuration(service.HealthCheck.Interval)
	if err != nil {
		common.Fatal("Error parsing interval: %v\n", err)
	}

	cronSpec := "@every " + interval.String()
	c.AddFunc(cronSpec, func() {
		monitorService(service)
	})
}

func monitorService(service common.Service) {
	resp, err := http.Get("http://" + service.Address + ":" + strconv.Itoa(service.Port) + service.HealthCheck.Endpoint)
	status := "unhealthy"
	if err == nil && resp.StatusCode == http.StatusOK {
		status = "healthy"
	}

	err = rdb.HSet(ctx, "service:"+service.ID, "status", status).Err()
	if err != nil {
		common.Warn("Error updating status for service %s: %v", service.Name, err)
	}
}
