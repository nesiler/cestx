package main

import (
	"context"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/redis"
	rc "github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

var (
	rdb *rc.Client
	ctx = context.Background()
	c   = cron.New()
)

func main() {
	godotenv.Load("../.env")
	godotenv.Load(".env")

	common.PYTHON_API_HOST = common.GetEnv("PYTHON_API_HOST", "http://192.168.4.99") // default value is your local IP
	common.TELEGRAM_TOKEN = common.GetEnv("TELEGRAM_TOKEN", "")
	common.CHAT_ID = common.GetEnv("CHAT_ID", "")

	common.SendMessageToTelegram("**REGISTRY** ::: Service starting...")

	// Initialize Redis client
	var err error
	cfg := common.LoadRedisConfig()
	rdb, err = redis.NewRedisClient(cfg)
	if err != nil {
		common.Fatal("Failed to connect to Redis: %v", err)
	}
	defer redis.Close(rdb)

	common.SendMessageToTelegram("**REGISTRY** ::: Redis client initialized")

	// Register routes and start server
	http.HandleFunc("/register", registerServiceHandler)
	http.HandleFunc("/service/", getServiceHandler)
	http.HandleFunc("/config/", getConfigHandler)
	http.HandleFunc("/health", common.HealthHandler()) // Health check endpoint

	go func() {
		c.AddFunc("@every 15s", checkUp) // Check service health every 15 seconds
		c.Start()
	}()

	currentHost, err := common.ExternalIP()
	common.FailError(err, "")

	common.Ok("Registry server started on: %s", currentHost)
	common.SendMessageToTelegram("**REGISTRY** ::: Server started on: " + currentHost)

	err = http.ListenAndServe(":3434", nil)
	common.FailError(err, "")

}
