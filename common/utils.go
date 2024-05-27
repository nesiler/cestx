package utils

import (
	"errors"
	"net"
	"os"

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
	Ok                = color.New(color.FgHiGreen).PrintlnFunc()
	downloadErrorsLog *os.File
)

func externalIP() (string, error) {
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
