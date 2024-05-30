package common

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	Head  = color.New(color.FgHiMagenta).Add(color.Bold).Add(color.Underline).Add(color.BgHiWhite).PrintlnFunc()
	Out   = color.New(color.FgHiWhite).PrintlnFunc()
	Info  = color.New(color.FgHiCyan).PrintlnFunc()
	Warn  = color.New(color.FgHiYellow).Add(color.Bold).PrintlnFunc()
	Err   = color.New(color.FgHiRed).Add(color.Bold).PrintlnFunc()
	Fatal = func(format string, args ...interface{}) error {
		msg := fmt.Sprintf(format, args...)
		color.New(color.FgHiRed).Add(color.Bold).Println(msg)
		return fmt.Errorf(msg) // Return a formatted error
	}
	Ok          = color.New(color.FgHiGreen).PrintlnFunc()
	foundEnvVar map[string]string
)

// SendMessageToChat sends a message to the chat using the Python API.
func SendMessageToTelegram(message string) {
	time.Sleep(1 * time.Second)
	// Create the JSON payload
	payload := map[string]string{"message": message}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		Fatal("Error marshalling JSON data: %v\n", err)
	}

	// Get the Python API host from environment variables
	apiHost, err := FindEnvVar("PYTHON_API_HOST")
	if err != nil {
		Fatal("PYTHON_API_HOST environment variable not set\n")
	}

	// Send the POST request to the Python API
	resp, err := http.Post("http://"+apiHost+"/send", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		Fatal("Error sending message: %v\n", err)
	}
	defer resp.Body.Close()

	// Check for success response
	if resp.StatusCode != http.StatusOK {
		Fatal("Failed to send message, received status code: %d\n", resp.StatusCode)
	}
}

// RegisterService sends a registration request to the registry with the given JSON data.
func RegisterService(jsonData []byte) {
	var service map[string]interface{}
	err := json.Unmarshal(jsonData, &service)
	if err != nil {
		Fatal("Error unmarshalling JSON data: %v\n", err)
	}

	// Check if the service JSON has an IP address, if not get it manually
	address, ok := service["address"].(string)
	if !ok || strings.TrimSpace(address) == "" {
		ip, err := ExternalIP()
		if err != nil {
			Fatal("Error getting external IP address: %v\n", err)
		}
		service["address"] = ip
	}

	// Marshal the updated service JSON data
	updatedJsonData, err := json.Marshal(service)
	if err != nil {
		Fatal("Error marshalling updated JSON data: %v\n", err)
	}

	// Get the registry host from environment variables
	registryHost := os.Getenv("REGISTRY_HOST")
	if registryHost == "" {
		Fatal("REGISTRY_HOST environment variable not set\n")
	}

	// Send the registration request to the registry
	resp, err := http.Post("http://"+registryHost+"/register", "application/json", bytes.NewBuffer(updatedJsonData))
	if err != nil {
		Fatal("Error sending registration request: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		Fatal("Failed to register service, received status code: %d\n", resp.StatusCode)
	}

	Ok("Service registered successfully")
}

// FindEnvVar searches for an environment variable in various locations recursively.
func FindEnvVar(varName string) (string, error) {
	// 1. Check system environment variables
	if val, exists := os.LookupEnv(varName); exists {
		return val, nil
	}

	// Check if the variable was already found
	if foundVal, ok := foundEnvVar[varName]; ok {
		return foundVal, nil
	}

	// 2. Check .env and config.json files recursively (2 levels up and down)
	Info("Searching for environment variable '" + varName + "' in .env and config.json files...\n")
	Info("Current directory: ", getCurrentDir(0))
	var foundVal string
	for level := -2; level <= 2; level++ {
		startDir := getCurrentDir(level)
		err := filepath.Walk(startDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip the starting directory itself
			if path == startDir {
				return nil
			}

			// Check for .env files
			if !info.IsDir() && info.Name() == ".env" {
				foundVal, err = readEnvFile(path, varName)
				if err != nil {
					Warn("Error reading .env file:", err)
					return nil // Continue searching
				}
				if foundVal != "" {
					return fmt.Errorf("found") // Signal that the variable was found
				}
			}

			// Check for config.json files
			if !info.IsDir() && info.Name() == "config.json" {
				foundVal, err = readConfigJSON(path, varName)
				if err != nil {
					Warn("Error reading config.json file:", err)
					return nil // Continue searching
				}
				if foundVal != "" {
					return fmt.Errorf("found") // Signal that the variable was found
				}
			}
			return nil
		})

		// If the variable was found during the walk, return it
		if err != nil && err.Error() == "found" {
			Info("Found environment variable '" + varName + "' = " + foundVal + "\n")
			foundEnvVar = map[string]string{varName: foundVal}
			return foundVal, nil
		}
	}

	return "", fmt.Errorf("environment variable " + varName + " not found")
}

func getCurrentDir(level int) string {
	dir, _ := os.Getwd() // Ignore error; default to current directory
	for i := 0; i < abs(level); i++ {
		if level > 0 {
			dir = filepath.Join(dir, "..") // Move up
		} else {
			files, _ := os.ReadDir(dir) // Ignore error; default to empty slice
			if len(files) > 0 {
				dir = filepath.Join(dir, files[0].Name()) // Move down (choose first subdirectory)
			}
		}
	}
	return dir
}

// readEnvFile reads a specific variable from a .env file
func readEnvFile(filename, varName string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, varName+"=") {
			return strings.TrimPrefix(line, varName+"="), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil // Variable not found
}

// readConfigJSON reads a specific variable from a config.json file
func readConfigJSON(filename, varName string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	var configMap map[string]interface{}
	err = json.Unmarshal(data, &configMap)
	if err != nil {
		return "", err
	}

	if val, ok := configMap[varName].(string); ok {
		return val, nil
	}

	return "", nil // Variable not found or not a string
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// HealthHandler returns an HTTP handler function for the health check endpoint.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}
}

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", Fatal("Error getting network interfaces: %v\n", err)
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", Fatal("Error getting addresses for interface %v: %v\n", iface.Name, err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", Fatal("Error getting external IP address: %v\n", errors.New("no IP address found"))
}
