package common

import (
	"container/list"
	"log"
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
	buf := make([]byte, 1024*1024*1) //1mb
	for running {
		read, err := inConn.Read(buf)
		if err != nil {
			log.Printf("read bytes form conn %v failed...\n", inConn.RemoteAddr())
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
		_, err = outConn.Write(pkg.ToBytes())

		if err != nil {
			running = false
		}
	}

	defer func() {
		CloseConn(append(make([]net.Conn, 2), inConn, outConn))
		wg.Done()
	}()
}

func TransferBytes(inConn net.Conn, outConn net.Conn, req []byte, len int, handlers []PackageHandler, wg *sync.WaitGroup) {
	running := true
	buf := make([]byte, 1024*1024*1) //1mb

	var header, body []byte
	body = make([]byte, len)
	copy(body[:], req[:len])
	pkg := handleBytes(make([]byte, 0), body, handlers)
	if _, err := writePkgToConn(outConn, pkg); err != nil {
		running = false
		return
	}

	for running {
		var read int
		var err error
		if read, err = inConn.Read(buf); err != nil {
			log.Printf("read bytes form conn %v failed...\n", inConn.RemoteAddr())
			running = false
		}

		header = make([]byte, 0)
		body = make([]byte, read)
		copy(body[:], buf[:read])
		pkg := handleBytes(header, body, handlers)

		if _, err = writePkgToConn(outConn, pkg); err != nil {
			running = false
		}
	}

	defer func() {
		CloseConn(append(make([]net.Conn, 2), inConn, outConn))
		wg.Done()
	}()
}

func writePkgToConn(outConn net.Conn, pkg Package) (bool, error) {
	if _, err := outConn.Write(pkg.ToBytes()); err != nil {
		//log.Printf("write bytes to conn %v failed...\n", outConn.RemoteAddr())
		return false, err
	}
	return true, nil
}

func handleBytes(header []byte, body []byte, handlers []PackageHandler) Package {
	pkg := *NewPackage()
	pkg.ValueOf(header, body)
	for _, handler := range handlers {
		pkg = handler.Handle(&pkg)
	}
	return pkg
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
		_, err = outConn.Write(pkg.body)
		if err != nil {
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
			_ = conn.Close()
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
