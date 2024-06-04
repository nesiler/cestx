package main

import (
	"context"
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
	common.SendMessageToTelegram("**DEPLOYER** ::: Deployer service updating itself")
	cmd := exec.Command("git", "-C", repoPath, "pull")
	output, err := cmd.CombinedOutput()
	common.FailError(err, "output: %s", err, string(output))

	// build and run this code again
	cmd = exec.Command("go", "build", "-o", "deployer")
	output, err = cmd.CombinedOutput()
	common.FailError(err, "output: %s", err, string(output))
	common.Ok("Built new binary: %s", string(output))

	latestCommit, err := client.GetLatestCommit(config.RepoOwner, config.RepoName)
	common.FailError(err, "Error getting latest commit: %v\n")
	writeLatestCommit(latestCommit)

	common.SendMessageToTelegram("**DEPLOYER** ::: Trying to restart deployer service ...")
	cmd = exec.Command("systemctl", "restart", "deployer.service")
	output, err = cmd.CombinedOutput()
	common.FailError(err, "output: %s", err, string(output))

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
	latestCommit, err := readLatestCommit()
	common.FailError(err, "Error reading latest commit from file")

	for {
		// Get the latest commit from GitHub
		commit, err := client.GetLatestCommit(config.RepoOwner, config.RepoName)
		common.FailError(err, "Error getting latest commit: %v\n")

		if commit != latestCommit {
			common.Ok("New commit detected: %s", commit)
			common.SendMessageToTelegram("New commit detected: " + commit)

			// Get the list of changed directories in the repository
			changedDirs, err := client.GetChangedDirs(config.RepoPath, commit)
			common.FailError(err, "Error getting changed directories")

			common.Info("Changed directories: %v", changedDirs)
			common.SendMessageToTelegram("**DEPLOYER** ::: Changed directories: " + strings.Join(changedDirs, ", "))

			// For each changed directory, deploy the corresponding service
			for _, dir := range changedDirs {
				// Check if the directory is "deployer", skip deployment and pull locally and run this program again
				if dir == "deployer" {
					common.Ok("Pulling latest changes for directory: %s", dir)
					if err := client.PullLatest(config.RepoPath); err != nil {
						common.Err("Error pulling latest changes for directory: %s: %v", dir, err)
						common.SendMessageToTelegram("**DEPLOYER** ::: Error pulling latest changes for directory: " + dir)
					}
					continue
				}
				common.Info("Trying to deploy: %s", dir)
				if err := Deploy(dir); err != nil {
					common.Warn("%v: %s", err, dir)
				} else {
					common.Ok("Successfully deployed: %s", dir)
					common.SendMessageToTelegram("**DEPLOYER** ::: Successfully deployed: " + dir)
				}
			}
			latestCommit = commit // Update latest commit hash
			err = writeLatestCommit(latestCommit)
			common.FailError(err, "Error writing latest commit to file")
		}

		time.Sleep(time.Second * time.Duration(config.CheckInterval))
	}
}
