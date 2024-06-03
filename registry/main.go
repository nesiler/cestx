package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/nesiler/cestx/common"
	"github.com/robfig/cron/v3"
)

type Config struct {
	ExternalServices map[string]ServiceInfo `json:"externalServices"`
}

type ServiceInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	DBName   string `json:"dbname,omitempty"`
}

var (
	rdb        *redis.Client
	ctx        = context.Background()
	configData Config
	c          = cron.New()
)

func main() {
	// Load configuration file
	configFile, err := os.ReadFile("config.json")
	common.FailError(err, "error reading config file")

	err = json.Unmarshal(configFile, &configData)
	common.FailError(err, "error parsing config file")

	// Initialize Redis client using config data
	redisConfig, ok := configData.ExternalServices["redis"]
	if !ok {
		common.Fatal("Redis configuration not found in config file\n")
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: redisConfig.Host + ":" + strconv.Itoa(redisConfig.Port),
	})

	http.HandleFunc("/register", registerServiceHandler)
	http.HandleFunc("/service/", getServiceHandler)
	http.HandleFunc("/config/", getConfigHandler)

	// Start the cron scheduler
	c.Start()

	currentHost, err := common.ExternalIP()
	common.FailError(err, "")

	common.Info("Server started on: ", currentHost)
	err = http.ListenAndServe(":3434", nil)
	common.FailError(err, "")
}
