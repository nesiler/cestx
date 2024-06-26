package rabbitmq

import (
	"encoding/json"

	"github.com/nesiler/cestx/common"
	"github.com/streadway/amqp"
)

// Publish publishes a message to the specified exchange and routing key.
// It handles common publishing tasks and error scenarios.
func Publish(ch *amqp.Channel, exchange, routingKey string, message interface{}) error {
	// Input validation
	if ch == nil {
		return common.Err("Channel is required")
	}

	body, err := json.Marshal(message)
	if err != nil {
		return common.Err("Failed to marshal message: %v", err)
	}

	// Publish the message
	err = ch.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return common.Err("Failed to publish message: %v", err)
	}

	common.Ok("Message published successfully to exchange '%s' with routing key '%s'", exchange, routingKey)
	return nil
}
