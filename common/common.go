package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
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
		color.Unset()
		os.Exit(1) // Terminate the program
	}
	Ok              = newlinePrintfFunc(color.New(color.FgHiGreen).PrintfFunc())
	TELEGRAM_TOKEN  string
	CHAT_ID         string
	PYTHON_API_HOST string
	REGISTRY_HOST   string
)

type Service struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Address     string      `json:"address"`
	Port        int         `json:"port"`
	HealthCheck HealthCheck `json:"healthCheck"`
}

type HealthCheck struct {
	Endpoint string `json:"endpoint"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}

func newlinePrintfFunc(f func(format string, a ...interface{})) func(format string, a ...interface{}) {
	return func(format string, a ...interface{}) {
		f(format+"\n", a...)
		color.Unset() // Reset the color settings
	}
}

func errorPrintfFunc(f func(format string, a ...interface{})) func(format string, a ...interface{}) error {
	return func(format string, a ...interface{}) error {
		fmt.Println()
		f(format+"\n", a...)
		color.Unset() // Reset the color settings
		return fmt.Errorf(format, a...)
	}
}

func FailError(err error, format string, args ...interface{}) {
	if err != nil {
		Err(format, args...)
		Fatal(format, err)
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
	if PYTHON_API_HOST == "" {
		FailError(err, "Error finding PYTHON_API_HOST: %v\n")
	}

	// Send the POST request to the Python API
	resp, err := http.Post("http://"+PYTHON_API_HOST+":5005/send", "application/json", bytes.NewBuffer(jsonData))
	FailError(err, "send message error: %v\n")
	defer resp.Body.Close()

	// Check for success response
	if resp.StatusCode != http.StatusOK {
		Fatal("Failed to send message, received status code: %d\n", resp.StatusCode)
	}
}

func JsonToService(jsonData []byte) (*Service, error) {
	var service Service
	err := json.Unmarshal(jsonData, &service)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON data: %v", err)
	}
	return &service, nil
}

func RegisterService(service *Service) error {

	// Marshal the updated service data
	updatedJsonData, err := json.Marshal(service)
	FailError(err, "Error marshalling Service JSON data: %v\n")

	// Check if the registry host is set
	if REGISTRY_HOST == "" {
		Fatal("Error finding REGISTRY_HOST: %v\n")
	}

	// Send the registration request to the registry
	resp, err := http.Post("http://"+REGISTRY_HOST+":3434/register", "application/json", bytes.NewBuffer(updatedJsonData))
	FailError(err, "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		Fatal("Failed to register service, received status code: %d\n", resp.StatusCode)
	}

	return nil
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
