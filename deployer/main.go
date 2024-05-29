package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/ini.v1"
)

func readInventory(filePath string) ([]string, error) {
	cfg, err := ini.Load(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read inventory file: %w", err)
	}

	hosts := []string{}
	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}
		for _, key := range section.Keys() {
			if ansibleHost := key.String(); ansibleHost != "" {
				hosts = append(hosts, ansibleHost)
			}
		}
	}

	return hosts, nil
}

func main() {

	inventoryPath := "./ansible/inventory.ini"
	hosts, err := readInventory(inventoryPath)
	if err != nil {
		log.Fatalf("Error reading inventory: %v", err)
	}

	if err := setupSSHKeysForHosts("master", "root", hosts); err != nil {
		log.Fatalf("Error setting up SSH keys: %v", err)
	}

	// Load configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize GitHub client
	client := NewGitHubClient(config.GitHubToken)

	// Store the latest commit hash
	var latestCommit string

	for {
		commit, err := client.GetLatestCommit(config.RepoOwner, config.RepoName)
		if err != nil {
			log.Printf("Error getting latest commit: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}

		if commit != latestCommit {
			log.Printf("New commit detected: %s", commit)
			latestCommit = commit
			err := client.PullLatest(config.RepoPath)
			if err != nil {
				log.Printf("Error pulling latest changes: %v", err)
				continue
			}

			changedDirs, err := client.GetChangedDirs(config.RepoPath, latestCommit)
			if err != nil {
				log.Printf("Error getting changed directories: %v", err)
				continue
			}

			for _, dir := range changedDirs {
				err := Deploy(config, dir)
				if err != nil {
					log.Printf("Error deploying %s: %v", dir, err)
				}
			}
		}

		time.Sleep(time.Second * time.Duration(config.CheckInterval))
	}
}
