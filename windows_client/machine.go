package main

import (
	"fmt"
	"os"
	"time"
)

var (
	systemID string
)

func init() {
	name, err := os.Hostname()
	if err == nil {
		systemID = name
	} else {
		systemID = fmt.Sprintf("%d", time.Now().Unix())
	}
}
