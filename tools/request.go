package tools

import (
	"fmt"
	"net"
)

func HandleRequest(src net.Conn, dest net.Conn, errMsg chan bool) {
	buf := make([]byte, 0xffff)
	for {
		n, err := src.Read(buf)
		if err != nil {
			fmt.Println("User Closed")
			break
		}

		b := buf[:n]

		n, err = dest.Write(b)
		if err != nil {
			fmt.Printf("dest err %v", err)
			break
		}
	}

	errMsg <- true
}
