package main

import (
	"fmt"
	"net"
	"time"
)

var (
	systemID string
)

func init() {
	netInterfaces, err := net.Interfaces()
	if err == nil && len(netInterfaces) > 0 {
		netIf := netInterfaces[0]
		systemID = string(netIf.HardwareAddr)
		if len(systemID) > 0 {
			return
		}
	}

	systemID = fmt.Sprintf("%d", time.Now().Unix())
}
