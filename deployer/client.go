package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v35/github"
	"github.com/nesiler/cestx/common"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	client *github.Client
}

var client *GitHubClient

func NewGitHubClient(token string) *GitHubClient {
	common.Info("Creating GitHub client")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	common.Ok("GitHub client created: %v", client.BaseURL.String())
	return &GitHubClient{client: client}
}

func (c *GitHubClient) GetLatestCommit(owner, repo string) (string, error) {
	ctx := context.Background()
	commits, _, err := c.client.Repositories.ListCommits(ctx, owner, repo, nil)
	if err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", nil
	}
	return *commits[0].SHA, nil
}

func (c *GitHubClient) GetChangedDirs(repoPath, latestCommit, lastKnownCommit string) ([]string, error) {
	common.Warn("Getting changed directories between %s and %s", lastKnownCommit, latestCommit)
	common.Out("Checking for changes in %s", repoPath)

	cmd := exec.Command("git", "-C", repoPath, "diff", "--name-only", lastKnownCommit, latestCommit)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	changedFiles := strings.Split(out.String(), "\n")
	changedDirs := make(map[string]bool)
	for _, file := range changedFiles {
		if dir := filepath.Dir(file); dir != "." {
			changedDirs[dir] = true
		}
	}

	var dirs []string
	for dir := range changedDirs {
		dirs = append(dirs, dir)
	}

	return dirs, nil
}

func getCurrentCommit(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func pullLatestChanges(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		common.Err("Failed to pull latest changes: %s, %v", out.String(), err)
		return fmt.Errorf("failed to pull latest changes: %w, output: %s", err, out.String())
	}
	common.Info("Latest changes pulled successfully: %s", out.String())
	return nil
}

func watchForChanges() {
	latestCommit, err := client.GetLatestCommit(config.RepoOwner, config.RepoName)
	if err != nil {
		common.Err("Failed to fetch latest commit: %v", err)
		return
	}

	// Load last known commit to prevent reprocessing the same commit on restart
	lastKnownCommit := loadLastKnownCommit()
	updateDeployer := false

	// If the last known commit is the same as the latest commit, no changes have been made
	if lastKnownCommit == latestCommit {
		common.Out("No changes detected")
		return
	}

	// Pull latest changes regardless, to keep deployer updated
	if err := pullLatestChanges(config.RepoPath); err != nil {
		return
	}

	// Fetch and compute diff
	changedDirs, err := client.GetChangedDirs(config.RepoPath, latestCommit, lastKnownCommit)
	if err != nil {
		common.Err("Failed to get changed directories: %v", err)
		return
	}

	// Handle other service updates
	for _, dir := range changedDirs {
		if dir != "deployer" {
			common.Info("Deploying changes for directory: %s", dir)
			err := Deploy(dir) // Ensure Deploy checks the host and service status

			if err != nil {
				common.Err("Failed to deploy service %s: %v", dir, err)
			}
		} else if dir == "deployer" {
			updateDeployer = true
		}
	}

	// Restart deployer if changes were detected
	if updateDeployer {
		restartService()
	}

	saveLastKnownCommit(latestCommit)
}

func saveLastKnownCommit(commit string) {
	err := os.WriteFile("/tmp/last_known_commit", []byte(commit), 0644)
	if err != nil {
		common.Err("Failed to save last known commit: %v", err)
	}
}

func loadLastKnownCommit() string {
	data, err := os.ReadFile("/tmp/last_known_commit")
	if err != nil {
		common.Err("Failed to read last known commit: %v", err)
		return ""
	}
	return string(data)
}

func restartService() {
	common.Info("Deployer changes detected; pulling updates and restarting...")

	// Restart the starter.service to apply changes
	// cmd := exec.Command("systemctl", "restart", "starter.service")
	cmd := exec.Command("systemctl", "restart", "deployer.service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		common.Err("Failed to restart deployer.service: %v", err)
	}

	// it should already be restarted, but just in case, sleep for a few seconds
	time.Sleep(3 * time.Second)
	return
}
