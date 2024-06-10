package main

import (
	"context"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/redis"
	"github.com/robfig/cron/v3"
)

var (
	rdb *redis.Client
	ctx = context.Background()
	c   = cron.New()
)

func main() {
	// TODO 1: Load environment variables
	// TODO 2: Initialize the Redis client
	// TODO 3: Start api server for the registry operations
	// TODO 4: Start cron job to check the services' health
	// TODO 5: Send a message to the Telegram bot when the server starts
	// TODO 6: Implement the registerServiceHandler, getServiceHandler, and getConfigHandler functions

	godotenv.Load("../.env")
	godotenv.Load(".env")

	common.PYTHON_API_HOST = common.GetEnv("PYTHON_API_HOST", "http://192.168.4.99")

	common.SendMessageToTelegram("**REGISTRY** ::: Service started")

	rdb, err := redis.NewRedisClient(common.LoadRedisConfig())
	common.SendMessageToTelegram("**REGISTRY** ::: Redis client initialized")
	common.FailError(err, "")

	defer redis.Close(rdb)

	http.HandleFunc("/register", registerServiceHandler)
	http.HandleFunc("/service/", getServiceHandler)
	http.HandleFunc("/config/", getConfigHandler)

	c.AddFunc("@every 15s", checkUp)
	c.Start()

	currentHost, err := common.ExternalIP()
	common.FailError(err, "")

	common.Ok("Server started on: %s", currentHost)
	err = http.ListenAndServe(":3434", nil)
	common.FailError(err, "")
	common.SendMessageToTelegram("**REGISTRY** ::: Server started on: " + currentHost)
}
