package common

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"strconv"
	"sync"
)

func init() {
}

var (
	Dns = NewAddress()
)

func AcceptConn(conn net.Conn) {
	b := make([]byte, 4096)
	n, error := conn.Read(b)
	if error != nil {
	_:
		conn.Close()
		return
	}

	requestReader := bufio.NewReader(bytes.NewReader(b[0:n]))
	request, error := http.ReadRequest(requestReader)
	address := request.Host
	_, port := Dns.hostAndPort(address)
	print(address)

	shutdownGroup := sync.WaitGroup{}
	shutdownGroup.Add(1)
	var fromConn = conn
	var toConn net.Conn
	var err interface{}

	if Dns.IsAccessible(address) {
		toConn, err = net.Dial("tcp", address)
		if err != nil {
			shutdownGroup.Done()
			return
		}

		print("+accessible")
		if portNumber, _ := strconv.Atoi(port); portNumber == 443 {
			_, error := conn.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))
			if error != nil {
				conn.Close()
			}
		} else {
			if _, err := toConn.Write(b[0:n]); err != nil {
				shutdownGroup.Done()
				conn.Close()
			}
		}

		go forward(fromConn, toConn, shutdownGroup)
	} else {
		print("-accessible")
		// remote proxy
	}

	// read request address
	// check direct or proxy
	// go proxy or direct
	shutdownGroup.Wait()
	defer func() {
		closeConn(fromConn)
		closeConn(toConn)
	}()

}

func closeConn(conn net.Conn) {
	if conn != nil {
		conn.Close()
	}
}

func forward(fromConn, toConn net.Conn, shutdownGroup sync.WaitGroup) {
	var error interface{}
	var n int

	b := make([]byte, 4096)
	n, error = fromConn.Read(b)
	if error != nil {
		shutdownGroup.Done()
		return
	}
	_, error = toConn.Write(b[0:n])
	if error != nil {
		shutdownGroup.Done()
		return
	}
}
