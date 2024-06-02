package main

import (
	"net/http"
	"os"
	"time"

	"github.com/nesiler/cestx/common"
)

func serviceTest() {
	// Load service configuration from file
	serviceFile, err := os.ReadFile("service.json")
	if err != nil {
		common.Fatal("Error reading service configuration file: %v\n", err)
	}

	// Register the service with the registry
	service, err := common.JsonToService(serviceFile)
	common.FailError(err, "")
	common.RegisterService(service)

	// Setup health check endpoint
	http.HandleFunc("/health", common.HealthHandler())
	ip, err := common.ExternalIP()
	if err != nil {
		common.Err("Error getting external IP: %v\n", err)
	}

	common.Info("Health check service running on http://%s:3333/health\n", ip)

	if err := http.ListenAndServe(":3333", nil); err != nil {
		common.Fatal("Error starting health check service: %v\n", err)
	}
}

func telegramTest() {
	//get time and send to telegram as message
	for i := 0; i < 5; i++ {
		message := time.Now().Format("2006-01-02 15:04:05")
		common.SendMessageToTelegram(message)
		// wait for 1 second
		time.Sleep(100 * time.Microsecond)
	}
}

func main() {
	// serviceTest()
	// envFinderTest()
	// // telegramTest()
	// keyPath := fmt.Sprintf("%s/.ssh/%s", os.Getenv("HOME"), "master")
	// if _, err := os.Stat(keyPath); os.IsNotExist(err) {
	// 	// print error message
	// 	fmt.Println("SSH key does not exist")
	// }
	// println(keyPath)

	// ip := "127.0.0.1"
	// common.Info("Health check service running on http://%s:3333/health\n", ip)

	// common.Head("--TEST STARTS--")
	// common.Out("This is a test message")
	// common.Info("This is an info message")
	// common.Warn("This is a warning message")
	// common.Err("This is an error message")
	// common.Ok("This is a success message")
	// common.FailError(fmt.Errorf("this is a test error"), "this is a test error message")
}
