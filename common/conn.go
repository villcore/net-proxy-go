package common

import (
	"container/list"
	"net"
	"sync"
)

type Connection struct {
	fromConn     *net.Conn
	toConn       *net.Conn
	sendHandlers *list.List
	recvHandlers *list.List
}

func TransferBytesToPackage(inConn net.Conn, outConn net.Conn, handlers []PackageHandler, wg *sync.WaitGroup) {
	running := true
	buf := make([]byte, 1024*100) //100kb

	for running {
		read, err := inConn.Read(buf)
		if err != nil {
			//log.Printf("read bytes form conn %v failed...\n", inConn.RemoteAddr())
			running = false
		}

		header := make([]byte, 0)
		body := make([]byte, read)

		copy(body[:], buf[:read])

		pkg := *NewPackage()
		pkg.ValueOf(header, body)

		for _, handler := range handlers {
			pkg = handler.Handle(&pkg)
		}
		//write一定是全部写入
		_, error := outConn.Write(pkg.ToBytes())
		if error != nil {
			//log.Printf("write bytes to conn %v failed...\n", outConn.RemoteAddr())
			running = false
		}
		//log.Printf("client write %v bytes to remote ...", write, outConn.RemoteAddr())

	}

	defer func() {
		CloseConn(append(make([]net.Conn, 2), inConn, outConn))
		wg.Done()
	}()
}

func TransferPackageToBytes(inConn net.Conn, outConn net.Conn, handlers []PackageHandler, wg *sync.WaitGroup) {
	running := true
	for running {
		pkg := *NewPackage()
		err := pkg.ReadWithHeader(inConn)

		if err != nil {
			running = false
		}

		for _, handler := range handlers {
			pkg = handler.Handle(&pkg)
		}
		//write一定是全部写入
		_, error := outConn.Write(pkg.body)
		if error != nil {
			running = false
		}
	}

	defer func() {
		CloseConn(append(make([]net.Conn, 2), inConn, outConn))
		wg.Done()
	}()
}

func CloseConn(conns []net.Conn) {
	for _, conn := range conns {
		if conn != nil {
			conn.Close()
		}
	}
}
func GetRemoteConn(addr string, port string) (net.Conn, error) {
	return NewRemoteConn(addr, port)
}

func NewRemoteConn(addr string, port string) (net.Conn, error) {
	addrAndPort := addr + ":" + port
	return net.Dial("tcp", addrAndPort)
}
