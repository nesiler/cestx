#!/bin/bash

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go and try again."
    exit 1
fi

# Tidy up the Go file
go mod tidy

# Build the Go file
go build -o setup setup.go

# get argument from the command line: setup or remove
if [ $? -eq 0 ]; then
    echo "Build successful. Running setup..."
    if [ "$1" == "setup" ]; then
        ./setup --insecure  --file setup.json setup
    elif [ "$1" == "remove" ]; then
        # Run the remove script
        ./setup --insecure remove
    else
        echo "Invalid argument. Please use 'setup' or 'remove'."
        exit 1
    fi
else
    echo "Build failed. Please fix any errors and try again."
fi