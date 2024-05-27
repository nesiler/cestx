package main

import (
	"fmt"
	"github.com/nesiler/cestx/common"
)

func main() {
	ip := externalIP()
	fmt.Println("IP address:", ip)
}
