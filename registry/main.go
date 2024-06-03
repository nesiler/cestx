package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

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

	common.PYTHON_API_HOST = os.Getenv("PYTHON_API_HOST")
	if common.PYTHON_API_HOST == "" {
		common.Warn("PYTHON_API_HOST not set, using default value")
		common.PYTHON_API_HOST = "http://192.168.4.99"
	}
	common.SendMessageToTelegram("**REGISTRY** ::: Service started at " + time.Now().String())

	// Initialize Redis client using config data
	redisConfig, ok := configData.ExternalServices["redis"]
	if !ok {
		common.Fatal("Redis configuration not found in config file\n")
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: redisConfig.Host + ":" + strconv.Itoa(redisConfig.Port),
	})
	common.SendMessageToTelegram("**REGISTRY** ::: Redis client initialized")

	http.HandleFunc("/register", registerServiceHandler)
	http.HandleFunc("/service/", getServiceHandler)
	http.HandleFunc("/config/", getConfigHandler)

	// Start the cron scheduler
	c.Start()

	// Schedule checkUp function to run periodically
	c.AddFunc("@every 1m", checkUp)

	currentHost, err := common.ExternalIP()
	common.FailError(err, "")

	common.Info("Server started on: ", currentHost)
	err = http.ListenAndServe(":3434", nil)
	common.FailError(err, "")
	common.SendMessageToTelegram("**REGISTRY** ::: Server started on: " + currentHost)
}
