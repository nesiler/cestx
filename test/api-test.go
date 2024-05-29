package main

import (
	"time"

	"github.com/nesiler/cestx/common"
)

func main() {
	//get time and send to telegram as message
	for i := 0; i < 50; i++ {
		message := time.Now().Format("2006-01-02 15:04:05")
		common.SendMessageToTelegram(message)

		// wait for 1 second
		time.Sleep(100 * time.Microsecond)
	}
}
