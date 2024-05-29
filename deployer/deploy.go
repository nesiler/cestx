package main

import (
    "log"
    "os/exec"
)

func Deploy(config *Config, serviceName string) error {
    cmd := exec.Command("ansible-playbook", "-i", config.AnsiblePath+"/inventory.ini", config.AnsiblePath+"/deploy.yml", "-e", "service="+serviceName)
    cmd.Stdout = log.Writer()
    cmd.Stderr = log.Writer()
    return cmd.Run()
}
