package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
)

var (
	Head  = newlinePrintfFunc(color.New(color.FgHiMagenta).Add(color.Bold).Add(color.Underline).PrintfFunc())
	Out   = newlinePrintfFunc(color.New(color.FgHiWhite).PrintfFunc())
	Info  = newlinePrintfFunc(color.New(color.FgCyan).PrintfFunc())
	Warn  = newlinePrintfFunc(color.New(color.FgHiYellow).Add(color.Bold).PrintfFunc())
	Err   = errorPrintfFunc(color.New(color.FgHiRed).Add(color.Bold).PrintfFunc())
	Fatal = func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		color.New(color.FgHiRed).Add(color.Bold).Add(color.BgBlack).Println(msg)
		os.Exit(1) // Terminate the program
	}
	Ok              = newlinePrintfFunc(color.New(color.FgHiGreen).PrintfFunc())
	TELEGRAM_TOKEN  string
	CHAT_ID         string
	PYTHON_API_HOST string
	REGISTRY_HOST   string
)

func newlinePrintfFunc(f func(format string, a ...interface{})) func(format string, a ...interface{}) {
	return func(format string, a ...interface{}) {
		f(format+"\n", a...)
	}
}

func errorPrintfFunc(f func(format string, a ...interface{})) func(format string, a ...interface{}) error {
	return func(format string, a ...interface{}) error {
		fmt.Println()
		f(format+"\n", a...)
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
	if err != nil {
		Warn("Error marshalling JSON data: %v", err)
	}

	// Get the Python API host from environment variables
	if PYTHON_API_HOST == "" {
		FailError(err, "Error finding PYTHON_API_HOST: %v\n")
	}

	// Send the POST request to the Python API
	resp, err := http.Post("http://"+PYTHON_API_HOST+":5005/send", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		Warn("send message error: %v", err)
	}

	defer resp.Body.Close()

	// Check for success response
	if resp.StatusCode != http.StatusOK {
		Warn("Failed to send message, received status code: %d\n", resp.StatusCode)
	}
}

func RegisterService(service *ServiceConfig) error {
	// Marshal the updated service data
	updatedJsonData, err := json.Marshal(service)
	FailError(err, "error marshalling updated service data: %v")

	// Check if the registry host is set
	if REGISTRY_HOST == "" {
		return Err("REGISTRY_HOST environment variable not set")
	}

	// Send the registration request to the registry
	resp, err := http.Post("http://"+REGISTRY_HOST+":3434/register", "application/json", bytes.NewBuffer(updatedJsonData))
	FailError(err, "error sending registration request: %v")
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Read the response body for more details
		return fmt.Errorf("failed to register service, received status code: %d, response: %s", resp.StatusCode, body)
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
