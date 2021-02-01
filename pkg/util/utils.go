package util

import (
	"fmt"
	"net"
	"os"
	"crypto/rand"
)

func GenerateUUID() string{
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
  	//   log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
	    b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	// fmt.Println(uuid)

	return uuid
}


func GetLocalIP() string{
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