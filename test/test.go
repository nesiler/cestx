package main

import (
	nsd "github.com/nesiler/cestx/common"
)

func main() {

	ip, err := nsd.ExternalIP()
	if err != nil {
		nsd.Err(err, "Error: ", err.Error())
		return
	}
	nsd.Info("External IP: ", ip)
}
