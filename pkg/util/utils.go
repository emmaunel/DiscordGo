package util

import (
	"fmt"
	"net"
	"os"
)

// DEBUG is set to true, lots of print statement
// comes alive
var DEBUG bool = false

// GenerateUUID returns a 8 bit(is that right?) UUID
func GenerateUUID() string {
	b := make([]byte, 16)
	// _, err := rand.Read(b)
	// if err != nil {
	//   log.Fatal(err)
	// }
	// This is cool but it's really long

	// uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
	// 	b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	uuid := fmt.Sprintf("%x",
		b[0:4])
	// fmt.Println(uuid)

	return uuid
}

// GetLocalIP return their IP
// I say local because the agent might be behind a NAT network
// And their external IP is gonna be different.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "nil"
}
