package rabbitmq

import (
	"github.com/nesiler/cestx/common"
	"github.com/streadway/amqp"
)

// Consume consumes messages from the specified queue.
// It takes a RabbitMQConnection, the queue name, and a handler function
// to process each received message.
func Consume(conn *RabbitMQConnection, queueName string, handler func(amqp.Delivery) error) error {
	if conn == nil || conn.Channel == nil {
		return common.Err("RabbitMQ connection or channel is nil")
	}

	// Declare the queue (makes the consumer idempotent)
	_, err := conn.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return common.Err("Failed to declare queue '%s': %v", queueName, err)
	}

	// Register the consumer
	msgs, err := conn.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (set to false to manually ack/nack)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return common.Err("Failed to register a consumer: %w", err)
	}

	common.Ok("Consumer started successfully for queue '%s'", queueName)

	// Process messages asynchronously
	go func() {
		for d := range msgs {
			// Process the message in the handler function
			err := handler(d)
			if err != nil {
				common.Err("Error processing message: %v", err)

				// Example: Negatively acknowledge the message with requeue
				if err := d.Nack(false, true); err != nil {
					common.Err("Failed to Nack the message: %v", err)
				}
			} else {
				// Acknowledge the message if processed successfully
				if err := d.Ack(false); err != nil {
					common.Err("Failed to acknowledge message: %v", err)
				}
			}
		}
	}()

	return nil
}
