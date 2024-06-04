#!/bin/bash

# Change to the desired directory
cd /home/cestx

# Pull the latest commit from the Git repository
git pull

# Change to the deployer directory
cd /home/cestx/deployer

# Build the Go file
go build -o deployer

# Restart the restart.service
systemctl restart deployer.service

# Finish the script
exit 0