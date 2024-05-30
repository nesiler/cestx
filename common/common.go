package common

import (
	"bufio"
	"bytes"
	"encoding/json"
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
	Head  = newlinePrintfFunc(color.New(color.FgHiMagenta).Add(color.Bold).Add(color.Underline).Add(color.BgHiWhite).PrintfFunc())
	Out   = newlinePrintfFunc(color.New(color.FgHiWhite).PrintfFunc())
	Info  = newlinePrintfFunc(color.New(color.FgCyan).PrintfFunc())
	Warn  = newlinePrintfFunc(color.New(color.FgHiYellow).Add(color.Bold).PrintfFunc())
	Err   = errorPrintfFunc(color.New(color.FgHiRed).Add(color.Bold).PrintfFunc())
	Fatal = func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		color.New(color.FgHiRed).Add(color.Bold).Add(color.BgBlack).Println(msg)
		os.Exit(1) // Terminate the program
	}
	Ok          = newlinePrintfFunc(color.New(color.FgHiGreen).PrintfFunc())
	foundEnvVar map[string]string
)

func newlinePrintfFunc(f func(format string, a ...interface{})) func(format string, a ...interface{}) {
	return func(format string, a ...interface{}) {
		f(format+"\n", a...)
	}
}

func errorPrintfFunc(f func(format string, a ...interface{})) func(format string, a ...interface{}) error {
	return func(format string, a ...interface{}) error {
		f(format+"\n", a...)
		return fmt.Errorf(format, a...)
	}
}

func FailError(err error, format string, args ...interface{}) {
	if err != nil {
		Fatal(format, err, args[0])
	}
}

// SendMessageToChat sends a message to the chat using the Python API.
func SendMessageToTelegram(message string) {
	time.Sleep(1 * time.Second)
	// Create the JSON payload
	payload := map[string]string{"message": message}
	jsonData, err := json.Marshal(payload)
	FailError(err, "Error marshalling JSON data: %v\n")

	// Get the Python API host from environment variables
	apiHost, err := FindEnvVar("PYTHON_API_HOST")
	if err != nil {
		Fatal("PYTHON_API_HOST environment variable not set\n")
	}

	// Send the POST request to the Python API
	resp, err := http.Post("http://"+apiHost+"/send", "application/json", bytes.NewBuffer(jsonData))
	FailError(err, "send message error: %v\n")
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
	FailError(err, "Error unmarshalling JSON data: %v\n")

	// Check if the service JSON has an IP address, if not get it manually
	address, ok := service["address"].(string)
	if !ok || strings.TrimSpace(address) == "" {
		ip, err := ExternalIP()
		FailError(err, "Error getting external IP: %v\n")
		service["address"] = ip
	}

	// Marshal the updated service JSON data
	updatedJsonData, err := json.Marshal(service)
	FailError(err, "Error marshalling JSON data: %v\n")

	// Get the registry host from environment variables
	registryHost := os.Getenv("REGISTRY_HOST")
	if registryHost == "" {
		Fatal("REGISTRY_HOST environment variable not set\n")
	}

	// Send the registration request to the registry
	resp, err := http.Post("http://"+registryHost+"/register", "application/json", bytes.NewBuffer(updatedJsonData))
	FailError(err, "")

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
	Warn("\nSearching for environment variable '" + varName + "' in .env and config.json files...\n")
	// Info("\nCurrent directory: ", getCurrentDir(0))
	var foundVal string
	for level := -2; level <= 2; level++ {
		startDir := getCurrentDir(level)
		err := filepath.Walk(startDir, func(path string, info os.FileInfo, err error) error {
			FailError(err, "")

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
					return Err("found") // Signal that the variable was found
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
					return Err("found") // Signal that the variable was found
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
	FailError(err, "")

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
	FailError(err, "")

	var configMap map[string]interface{}
	err = json.Unmarshal(data, &configMap)
	FailError(err, "")

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
	FailError(err, "Interfaces error: %v\n")

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		FailError(err, "")

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
	return "", Err("No network connection found")
}
