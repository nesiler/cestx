package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-github/v35/github"
	"github.com/nesiler/cestx/common"
	"golang.org/x/oauth2"
)

const latestCommitFile = "/tmp/latest_commit.txt"

type GitHubClient struct {
	client *github.Client
}

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

func (client *GitHubClient) PullLatest(repoPath string) error {
	// Get the current commit ID
	currentCommit, err := getCurrentCommit(repoPath)
	if err != nil {
		return fmt.Errorf("failed to get current commit ID: %w", err)
	}
	common.Head("Current commit: %s", currentCommit)

	latestCommit, err := client.GetLatestCommit(config.RepoOwner, config.RepoName)
	if err != nil {
		return fmt.Errorf("failed to get latest commit ID: %w", err)
	}
	common.Head("Latest commit: %s", latestCommit)

	if currentCommit == latestCommit {
		common.Info("Already at the latest commit: %s", currentCommit)
		return nil
	}

	common.Info("Pulling latest changes from GitHub...")
	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull latest changes: %w", err)
	}

	// Build new binary
	cmd = exec.Command("go", "build", "-o", "deployer")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build new binary: %w", err)
	}
	common.Out("New binary built successfully")

	common.Info("Restarting deployer service...")
	cmd = exec.Command("systemctl", "restart", "deployer.service")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart deployer service: %w", err)
	}

	os.Exit(0)
	return nil
}

func (c *GitHubClient) GetChangedDirs(repoPath, latestCommit string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "diff", "--name-only", latestCommit+"^!")
	output, err := cmd.Output()
	common.FailError(err, "")

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	dirSet := make(map[string]struct{})
	for _, file := range changedFiles {
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			dirSet[parts[0]] = struct{}{}
		}
	}

	var dirs []string
	for dir := range dirSet {
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

func readLatestCommit() (string, error) {
	data, err := os.ReadFile(latestCommitFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // If the file does not exist, return an empty string
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func writeLatestCommit(commit string) error {
	return os.WriteFile(latestCommitFile, []byte(commit), 0644)
}

// watchForChanges watches for new commits and triggers deployments
func watchForChanges(client *GitHubClient) {
	for {
		latestCommit, err := client.GetLatestCommit(config.RepoOwner, config.RepoName)
		if err != nil {
			common.Err("Error getting latest commit: %v", err)
			time.Sleep(time.Duration(config.CheckInterval) * time.Second)
			continue
		}

		lastDeployedCommit, err := readLatestCommit()
		if err != nil {
			common.Err("Error reading last deployed commit: %v", err)
			time.Sleep(time.Duration(config.CheckInterval) * time.Second)
			continue
		}

		if latestCommit != lastDeployedCommit {
			common.Info("New commit detected: %s", latestCommit)

			changedDirs, err := client.GetChangedDirs(config.RepoPath, latestCommit)
			common.Out("Changed directories: %v", changedDirs)
			if err != nil {
				common.Err("Error getting changed directories: %v", err)
				time.Sleep(time.Duration(config.CheckInterval) * time.Second)
				continue
			}

			for _, dir := range changedDirs {
				if dir == "deployer" {
					common.Info("Deployer code updated, pulling changes...")
					err := client.PullLatest(config.RepoPath)
					if err != nil {
						common.Err("Error updating deployer: %v", err)
					}
					// No need to continue the loop after updating the deployer
					break
				} else {
					common.Out("Deploying service: %s", dir)
					common.SendMessageToTelegram("**DEPLOYER** ::: Deploying service: " + dir)
					err = Deploy(dir)
					if err != nil {
						common.Err("Error deploying service %s: %v", dir, err)
					}
				}
			}

			if err := writeLatestCommit(latestCommit); err != nil {
				common.Err("Error writing latest commit to file: %v", err)
			}
		} else {
			common.Info("No changes detected")
		}

		time.Sleep(time.Duration(config.CheckInterval) * time.Second)
	}
}
