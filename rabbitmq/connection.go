package rabbitmq

import (
	"fmt"
	"time"

	"github.com/nesiler/cestx/common"
	"github.com/streadway/amqp"
)

// rmqClient creates a new connection to the RabbitMQ server
func NewConnection(cfg *common.RabbitMQConfig) (*amqp.Connection, error) {
	conn, err := connect(cfg)
	if err != nil {
		common.Err("Failed to connect to RabbitMQ: %v", err) // Use common.Err
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	common.Info("Connected to RabbitMQ!")
	return conn, nil
}

// connect attempts to establish a connection to the RabbitMQ server
// with retries and exponential backoff.
func connect(cfg *common.RabbitMQConfig) (*amqp.Connection, error) {
	var err error
	var conn *amqp.Connection
	retries := 3                     // Total connection attempts
	retryInterval := 2 * time.Second // Initial retry interval

	for i := 1; i <= retries; i++ {
		common.Out("Attempting to connect to RabbitMQ (attempt %d/%d)", i, retries)
		conn, err = amqp.Dial(cfg.Host)
		if err == nil {
			common.Ok("Successfully connected to RabbitMQ!")
			return conn, nil // Successful connection
		}

		common.Warn("Connection attempt %d failed: %v", i, err)
		if i < retries {
			common.Info("Retrying in %v...", retryInterval)
			time.Sleep(retryInterval)
			retryInterval *= 2 // Exponential backoff
		}
	}

	return nil, fmt.Errorf("failed to connect to RabbitMQ after %d retries: %w", retries, err)
}
