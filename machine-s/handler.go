package main

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/rabbitmq"
	"github.com/streadway/amqp"
)

// handleMessage processes RabbitMQ messages for machine operations.
func handleMessage(delivery amqp.Delivery, amqpConn *amqp.Connection) error {
	// 1. Unmarshal the message
	var machineMessage rabbitmq.MachineMessage
	if err := json.Unmarshal(delivery.Body, &machineMessage); err != nil {
		// Log the error, but don't return it (Nack below)
		common.Err("Error unmarshalling machine message: %v", err)
		return delivery.Nack(false, true) // Nack to requeue
	}

	common.Info("Received message: %+v", machineMessage)

	// 2. Handle different machine events
	switch machineMessage.Event {
	case rabbitmq.MachineCreate:
		return handleCreateMachine(machineMessage, amqpConn)
	case rabbitmq.MachineStart:
		return handleStartMachine(machineMessage, amqpConn)
	case rabbitmq.MachineStop:
		return handleStopMachine(machineMessage, amqpConn)
	case rabbitmq.MachineDelete:
		return handleDeleteMachine(machineMessage, amqpConn)
	default:
		// Handle unknown events (you might choose to Nack instead)
		common.Warn("Unknown machine event: %s", machineMessage.Event)
		return delivery.Nack(false, false) // Nack without requeue (dead-letter queue might be better)
	}
}

// Individual handler functions for each machine event:

func handleCreateMachine(msg rabbitmq.MachineMessage, amqpConn *amqp.Connection) error {
	// Implement machine creation logic
	if err := CreateMachine(msg, amqpConn); err != nil {
		// Handle errors (e.g., log, Nack the message, etc.)
		return fmt.Errorf("failed to create machine: %w", err)
	}
	return nil
}

func handleStartMachine(msg rabbitmq.MachineMessage, amqpConn *amqp.Connection) error {
	// You'll likely need the machine ID to start a specific machine
	machineID, err := uuid.Parse(msg.MachineID.String())
	if err != nil {
		return fmt.Errorf("invalid machine ID: %w", err)
	}

	// Implement machine start logic
	if err := StartMachine(machineID, amqpConn); err != nil {
		// Handle errors (e.g., log, Nack the message, etc.)
		return fmt.Errorf("failed to start machine: %w", err)
	}
	return nil
}

func handleStopMachine(msg rabbitmq.MachineMessage, amqpConn *amqp.Connection) error {
	// You'll likely need the machine ID to start a specific machine
	machineID, err := uuid.Parse(msg.MachineID.String())
	if err != nil {
		return fmt.Errorf("invalid machine ID: %w", err)
	}

	// Implement machine stop logic
	if err := StopMachine(machineID, amqpConn); err != nil {
		// Handle errors (e.g., log, Nack the message, etc.)
		return fmt.Errorf("failed to stop machine: %w", err)
	}
	return nil
}

func handleDeleteMachine(msg rabbitmq.MachineMessage, amqpConn *amqp.Connection) error {
	// You'll likely need the machine ID to start a specific machine
	machineID, err := uuid.Parse(msg.MachineID.String())
	if err != nil {
		return fmt.Errorf("invalid machine ID: %w", err)
	}

	// Implement machine deletion logic
	if err := DeleteMachine(machineID, amqpConn); err != nil {
		// Handle errors (e.g., log, Nack the message, etc.)
		return fmt.Errorf("failed to delete machine: %w", err)
	}
	return nil
}
