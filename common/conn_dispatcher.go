package common

import (
	"log"
	"net"
	"os"
	"strconv"
	"sync"
)

func init() {
	log.SetOutput(os.Stdout)
}

var (
	Dns = NewAddress()
)

func AcceptConn(conn net.Conn) {
	b := make([]byte, 4096)
	var n, e = conn.Read(b)
	if e != nil {
		_ = conn.Close()
		return
	}

	addr, port, e := ParseHttpOrHttps(b, n)
	if e != nil {
		_ = conn.Close()
	}

	shutdownGroup := sync.WaitGroup{}
	shutdownGroup.Add(1)

	var fromConn = conn
	var toConn net.Conn
	var err interface{}

	if Dns.IsAccessible(addr, port) {
		toConn, err = net.Dial("tcp", addr+":"+strconv.Itoa(port))
		if err != nil {
			shutdownGroup.Done()
			return
		}

		log.Printf("conn to %v:%v is accessible. \n", addr, port)
		if port == HTTPS {
			_, e := conn.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))
			if e != nil {
				_ = conn.Close()
			}
		} else {
			if _, err := toConn.Write(b[0:n]); err != nil {
				shutdownGroup.Done()
				_ = conn.Close()
			}
		}

		go forward(fromConn, toConn, shutdownGroup)
	} else {
		log.Printf("conn to %v:%v is unccessible. \n", addr, port)
		// remote proxy
	}

	shutdownGroup.Wait()
	defer func() {
		closeConn(fromConn)
		closeConn(toConn)
	}()

}

func closeConn(conn net.Conn) {
	if conn != nil {
		_ = conn.Close()
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
