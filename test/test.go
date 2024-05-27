package main

import (
	nsd "github.com/nesiler/cestx/common"
)

func main() {

	nsd.Head("Test")
	nsd.Out("This is a test.")
	nsd.Info("This is an info message.")

	ip, err := nsd.ExternalIP()
	if err != nil {
		nsd.Fatal("%s", "Error: ", err.Error())
		return
	}
	nsd.Info("External IP: ", ip)
}
