package main

import (
	"context"
	"fmt"

	"github.com/nesiler/cestx/common"
)

// HandleMachineCreate processes messages from the machine.create queue.
func HandleMachineCreate(msg common.RabbitMQMessage) error {
	var machineMsg common.MachineMessage
	if err := msg.Unmarshal(machineMsg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	common.Head("Received a message: %s", machineMsg)

	ctx := context.Background()

	// Call the createMachine function (which you already defined in main.go)
	if err := createMachine(db, dockerClient, minioClient, minioCfg.TemplatesBucket, &machineMsg); err != nil {
		return fmt.Errorf("failed to create machine: %w", err)
	}

	return nil // Indicate successful processing
}
