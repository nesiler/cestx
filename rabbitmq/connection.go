package rabbitmq

import (
	"fmt"
	"time"

	"github.com/nesiler/cestx/common"
	"github.com/streadway/amqp"
)

// RabbitMQConnection represents a connection to a RabbitMQ server
type RabbitMQConnection struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Config     *common.RabbitMQConfig
}

// NewRabbitMQConnection creates a new connection to the RabbitMQ server
func NewRabbitMQConnection(cfg *common.RabbitMQConfig) (*RabbitMQConnection, error) {
	conn, err := connect(cfg)
	if err != nil {
		common.Err("Failed to connect to RabbitMQ: %v", err) // Use common.Err
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	common.Info("Connected to RabbitMQ!")

	ch, err := conn.Channel()
	if err != nil {
		common.Err("Failed to open a channel: %v", err)
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	common.Info("RabbitMQ channel opened.")

	return &RabbitMQConnection{
		Connection: conn,
		Channel:    ch,
		Config:     cfg,
	}, nil
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
		conn, err = amqp.Dial(cfg.URL)
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

// Close closes the RabbitMQ connection and channel.
func (c *RabbitMQConnection) Close() error {
	common.Info("Closing RabbitMQ connection...")

	if c.Channel != nil {
		if err := c.Channel.Close(); err != nil {
			common.Err("Error closing RabbitMQ channel: %v", err)
			return fmt.Errorf("error closing channel: %w", err)
		}
		common.Info("RabbitMQ channel closed.")
	}

	if c.Connection != nil {
		if err := c.Connection.Close(); err != nil {
			common.Err("Error closing RabbitMQ connection: %v", err)
			return fmt.Errorf("error closing connection: %w", err)
		}
		common.Ok("RabbitMQ connection closed successfully.")
	}

	return nil
}
