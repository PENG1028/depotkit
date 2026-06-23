package utils

import (
	"fmt"
	"net"
	"time"
)

// PortAvailable checks if a TCP port is available.
func PortAvailable(port int) bool {
	target := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", target, 2*time.Second)
	if err != nil {
		return true
	}
	conn.Close()
	return false
}

// CheckPorts checks multiple ports and returns a list of occupied ones.
func CheckPorts(ports []int) []int {
	var occupied []int
	for _, port := range ports {
		if !PortAvailable(port) {
			occupied = append(occupied, port)
		}
	}
	return occupied
}

// RequiredPorts returns the default ports for all services.
func RequiredPorts() []int {
	return []int{5432, 6379, 9000, 9001, 27017}
}
