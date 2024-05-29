package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	Head  = color.New(color.FgHiMagenta).Add(color.Bold).Add(color.Underline).Add(color.BgHiWhite).PrintlnFunc()
	Out   = color.New(color.FgHiWhite).PrintlnFunc()
	Info  = color.New(color.FgHiCyan).PrintlnFunc()
	Warn  = color.New(color.FgHiYellow).Add(color.Bold).PrintlnFunc()
	Err   = color.New(color.FgHiRed).Add(color.Bold).FprintfFunc()
	Fatal = func(format string, args ...interface{}) {
		color.New(color.FgHiRed).Add(color.Bold).Printf(format, args...)
		os.Exit(1)
	}
	Ok = color.New(color.FgHiGreen).PrintlnFunc()
)

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
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
			return "", err
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
	return "", errors.New("Error: No network connection found.")
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

// HealthHandler returns an HTTP handler function for the health check endpoint.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}
}