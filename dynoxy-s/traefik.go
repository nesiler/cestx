package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nesiler/cestx/common"
)

func configureTraefik(subdomain, containerIP string, port int) error {
	// 1. Construct the Traefik API payload
	// Assuming you are using docker provider, you can use labels for dynamic configuration.
	// This configuration creates a new service and a new router, and links them together.

	// Service configuration
	service := map[string]interface{}{
		"loadBalancer": map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"url": fmt.Sprintf("http://%s:%d", containerIP, port),
				},
			},
		},
	}

	// Router configuration
	router := map[string]interface{}{
		"rule":        fmt.Sprintf("Host(`%s`)", subdomain),
		"service":     subdomain,
		"middlewares": []string{"auth"}, // Assuming you have an "auth" middleware
	}

	// Wrap service and router configuration in a JSON object
	config := map[string]interface{}{
		"http": map[string]interface{}{
			"services": map[string]interface{}{
				subdomain: service,
			},
			"routers": map[string]interface{}{
				subdomain: router,
			},
		},
	}

	// Convert the config map to JSON
	jsonConfig, err := json.Marshal(config)
	if err != nil {
		return common.Err("Failed to marshal Traefik config: %w", err)
	}

	// 2. Make the API request to Traefik
	traefikURL := "http://localhost:8080/api/providers/docker/configure" // Update with your Traefik API endpoint
	req, err := http.NewRequest(http.MethodPost, traefikURL, bytes.NewBuffer(jsonConfig))
	if err != nil {
		return common.Err("Failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return common.Err("Failed to update Traefik config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return common.Err("Traefik API returned non-OK status: %s", resp.Status)
	}

	common.Ok("Traefik config updated successfully for subdomain: %s", subdomain)
	return nil
}

func removeSubdomain(subdomain string) error {
	// 1. Make a DELETE request to remove the router
	routerURL := fmt.Sprintf("http://localhost:8080/api/providers/docker/routers/%s", subdomain)
	reqRouter, err := http.NewRequest(http.MethodDelete, routerURL, nil)
	if err != nil {
		return common.Err("Failed to create router delete request: %w", err)
	}

	client := &http.Client{}
	respRouter, err := client.Do(reqRouter)
	if err != nil {
		return common.Err("Failed to delete router from Traefik: %w", err)
	}
	defer respRouter.Body.Close()

	if respRouter.StatusCode != http.StatusOK {
		// Log the error but don't fail the operation (the router might not exist)
		common.Warn("Traefik API returned non-OK status for router deletion: %s", respRouter.Status)
	} else {
		common.Ok("Traefik router deleted successfully for subdomain: %s", subdomain)
	}

	// 2. Make a DELETE request to remove the service
	serviceURL := fmt.Sprintf("http://localhost:8080/api/providers/docker/services/%s", subdomain)
	reqService, err := http.NewRequest(http.MethodDelete, serviceURL, nil)
	if err != nil {
		return common.Err("Failed to create service delete request: %w", err)
	}

	respService, err := client.Do(reqService)
	if err != nil {
		return common.Err("Failed to delete service from Traefik: %w", err)
	}
	defer respService.Body.Close()

	if respService.StatusCode != http.StatusOK {
		// Log the error but don't fail the operation (the service might not exist)
		common.Warn("Traefik API returned non-OK status for service deletion: %s", respService.Status)
	} else {
		common.Ok("Traefik service deleted successfully for subdomain: %s", subdomain)
	}

	return nil
}
