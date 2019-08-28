package common

import (
	"net"
	"net/http"
)

func init() {

}

func AcceptConn(conn net.Conn) {
	b := make([]byte, 4096)
	n, error := conn.Read(b)

	if (error != nil) {
		conn.Close()
		return
	}

	http.ReadRequest(bufio.)

	string(b[0:n]);

	// read request address
	// check direct or proxy
	// go proxy or direct
}