package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/nesiler/cestx/common"
	"github.com/nesiler/cestx/rabbitmq"
	"github.com/streadway/amqp"
)

// handleDynoxyCreate handles messages for creating subdomains
func handleDynoxyCreate(delivery amqp.Delivery) error {
	// Unmarshal the message
	var message rabbitmq.DynoxyMessage
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		common.Err("Error unmarshalling dynoxy.create message: %v", err)
		return delivery.Nack(false, true) // Nack to requeue
	}
	common.Info("Received message: %+v", message)

	// Generate the subdomain
	subdomain := generateSubdomain(message.MachineID, message.UserID, message.Port)
	common.Info("Generated subdomain: %s", subdomain)

	// Get the container IP address
	containerIP, err := getContainerIP(message.MachineID.String()) // Assuming MachineID is the container ID
	if err != nil {
		common.Err("Error getting container IP: %v", err)
		return delivery.Nack(false, true) // Nack to requeue
	}
	common.Info("Container IP: %s", containerIP)

	// Configure Traefik
	if err := configureTraefik(subdomain, containerIP, message.Port); err != nil {
		common.Err("Error configuring Traefik: %v", err)
		return delivery.Nack(false, true) // Nack to requeue
	}

	// Acknowledge the message after successful processing
	return delivery.Ack(false)
}

// handleDynoxyDelete handles messages for deleting subdomains
func handleDynoxyDelete(delivery amqp.Delivery) error {
	// Unmarshal the message
	var message rabbitmq.DynoxyMessage
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		common.Err("Error unmarshalling dynoxy.delete message: %v", err)
		return delivery.Nack(false, true) // Nack to requeue
	}
	common.Info("Received message: %+v", message)

	// Generate the subdomain (to be removed)
	subdomain := generateSubdomain(message.MachineID, message.UserID, message.Port)

	// Remove the subdomain from Traefik
	if err := removeSubdomain(subdomain); err != nil {
		// It's likely okay to log the error and acknowledge the message,
		// even if subdomain removal fails. The container is likely already gone.
		common.Err("Error removing subdomain from Traefik: %v", err)
	}

	// Acknowledge the message
	return delivery.Ack(false)
}

// generateSubdomain creates a subdomain string
func generateSubdomain(machineID, userID uuid.UUID, port int) string {
	return fmt.Sprintf("%d-%s-%s.%s", port, machineID.String()[:8], userID.String()[:8], "yourdomain.com") // Replace with your actual domain
}

// getContainerIP retrieves the IP address of a container
func getContainerIP(containerID string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("error creating Docker client: %w", err)
	}

	containerJSON, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return "", fmt.Errorf("error inspecting container: %w", err)
	}

	// Access the container's IP address
	containerIP := containerJSON.NetworkSettings.IPAddress

	return containerIP, nil
}
