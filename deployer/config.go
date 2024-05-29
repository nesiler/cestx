package main

import (
    "encoding/json"
    "io/ioutil"
)

type Config struct {
    GitHubToken   string `json:"github_token"`
    RepoOwner     string `json:"repo_owner"`
    RepoName      string `json:"repo_name"`
    RepoPath      string `json:"repo_path"`
    CheckInterval int    `json:"check_interval"`
    AnsiblePath   string `json:"ansible_path"`
}

func LoadConfig(filename string) (*Config, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var config Config
    err = json.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }

    return &config, nil
}
