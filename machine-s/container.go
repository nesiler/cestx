package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/nesiler/cestx/common"
)

// buildImage builds a Docker image from the specified Dockerfile.
func buildImage(dockerfilePath, imageName string) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("Error creating Docker client: %v", err)
	}

	dockerfile, err := os.Open(dockerfilePath)
	if err != nil {
		return "", fmt.Errorf("Error opening Dockerfile: %v", err)
	}
	defer dockerfile.Close()

	response, err := cli.ImageBuild(
		ctx,
		dockerfile,
		types.ImageBuildOptions{
			Dockerfile: "Dockerfile", // Assuming the Dockerfile name is "Dockerfile"
			Tags:       []string{imageName},
		},
	)
	if err != nil {
		return "", fmt.Errorf("Error building image: %v", err)
	}
	defer response.Body.Close()

	// Read the build output from the response body
	buildOutput, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading build output: %v", err)
	}

	common.Info("Docker Build Output:\n%s", string(buildOutput))
	return imageName, nil
}

// runContainer runs a Docker container with the specified image and settings.
func runContainer(imageName, containerName string, ports []string, resources ...string) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("Error creating Docker client: %v", err)
	}

	portBindings := nat.PortMap{}
	for _, portMapping := range ports {
		parts := strings.Split(portMapping, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("Invalid port mapping: %s", portMapping)
		}

		hostPort, containerPort := parts[0], parts[1]

		portBindings[nat.Port(containerPort+"/tcp")] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	var cpuShares int64
	var memory int64
	for _, resource := range resources {
		parts := strings.Split(resource, "=")
		if len(parts) != 2 {
			return "", fmt.Errorf("Invalid resource limit: %s", resource)
		}

		key, value := parts[0], parts[1]

		switch key {
		case "cpu":
			cpuShares = parseQuantity(value) // Assuming parseQuantity parses CPU units correctly
		case "memory":
			memory = parseQuantity(value) // Assuming parseQuantity parses memory units correctly
		default:
			return "", fmt.Errorf("Unsupported resource limit: %s", key)
		}
	}

	containerConfig := &container.Config{
		Image: imageName,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Resources: container.Resources{
			CPUShares: cpuShares,
			Memory:    memory,
		},
		// NetworkMode: "host", // If you want to run in host network mode
	}

	networkingConfig := &network.NetworkingConfig{}
	containerResponse, err := cli.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkingConfig,
		nil, // platform
		containerName,
	)

	if err != nil {
		return "", fmt.Errorf("Error creating container: %v", err)
	}

	if err := cli.ContainerStart(ctx, containerResponse.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("Error starting container: %v", err)
	}

	common.Ok("Container %s started successfully with ID: %s", containerName, containerResponse.ID)

	return containerResponse.ID, nil
}

// startContainer starts a stopped Docker container.
func startContainer(containerID string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("Error creating Docker client: %v", err)
	}
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("Error starting container: %v", err)
	}
	common.Ok("Container %s started successfully.", containerID)
	return nil
}

// stopContainer stops a running Docker container.
func stopContainer(containerID string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("Error creating Docker client: %v", err)
	}
	if err := cli.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("Error stopping container: %v", err)
	}
	common.Ok("Container %s stopped successfully.", containerID)
	return nil
}

// removeContainer removes a Docker container.
func removeContainer(containerID string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("Error creating Docker client: %v", err)
	}
	if err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("Error removing container: %v", err)
	}
	common.Ok("Container %s removed successfully.", containerID)
	return nil
}

// Simple helper function to parse resource quantities (e.g., "512m", "1.5g")
func parseQuantity(quantityStr string) int64 {
	// For simplicity, assuming bytes for now
	var quantity int64
	fmt.Sscanf(quantityStr, "%d", &quantity)
	return quantity
}
