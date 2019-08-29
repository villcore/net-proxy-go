package common

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"../encrypt"
)

func init() {
	log.SetOutput(os.Stdout)
}

var (
	Addr = NewAddress()
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

	var fromConn = conn
	var toConn net.Conn
	defer func() {
		closeConn(fromConn)
		closeConn(toConn)
		log.Printf("conn local [%v] to [%v] is closed. \n", fromConn.LocalAddr(), toConn.RemoteAddr())
	}()

	shutdownGroup := sync.WaitGroup{}
	shutdownGroup.Add(2)
	var err interface{}

	if Addr.IsAccessible(addr, port) {
		log.Printf("conn to %v:%v is acessible ✅. \n", addr, port)

		toConn, err = net.Dial("tcp", addr+":"+strconv.Itoa(port))
		if err != nil {
			shutdownGroup.Add(-2)
			return
		}
		if port == 443 {
			connectResp := "HTTP/1.0 200 Connection Established\r\n\r\n"
			_, e := fromConn.Write([]byte(connectResp))
			if e != nil {
				_ = conn.Close()
				shutdownGroup.Add(-2)
				return
			}
		} else {
			if _, err := toConn.Write(b[0:n]); err != nil {
				_ = conn.Close()
				shutdownGroup.Add(-2)
				return
			}
		}
		go forward(fromConn, toConn, shutdownGroup)
		go forward(toConn, fromConn, shutdownGroup)
	} else {
		log.Printf("conn to %v:%v is unacessible ❌. \n", addr, port)
		remoteAddr := "207.246.108.224"
		remotePort := "20081"
		password := "villcore2"
		proxyRemoteConn(fromConn, remoteAddr, remotePort, password)
		shutdownGroup.Add(-2)
	}
	shutdownGroup.Wait()
}

func closeConn(conn net.Conn) {
	if conn != nil {
		_ = conn.Close()
	}
}

func forward(fromConn, toConn net.Conn, shutdownGroup sync.WaitGroup) {
	b := make([]byte, 4096)
	var n int
	var err error

	for true {
		if n, err = fromConn.Read(b); err != nil {
			shutdownGroup.Done()
			return
		}
		// log.Printf("read from %d bytes.\n %s \n", n, string(b[0:n]))
		log.Printf("🚀 ⬇️ %-5d bytes. \n", n)
		if n, err = toConn.Write(b[0:n]); err != nil {
			shutdownGroup.Done()
			return
		}
		log.Printf("🚀 ⬆️ %-5d bytes. \n", n)
	}
}

func proxyRemoteConn(localConn net.Conn, remoteAddr string, remotePort string, password string) {
	var bytesToPackageHandlers = make([]PackageHandler, 0)
	var packageToBytesHandlers = make([]PackageHandler, 0)
	//
	cipher, err := encrypt.NewCipher(password)
	if err != nil {
		fmt.Println("init cipher error ...")
	}

	encryptHandler := NewEncryptHandler(cipher)
	bytesToPackageHandlers = append(bytesToPackageHandlers, encryptHandler)

	//
	decryptHandler := NewDecryptHandler(cipher)
	packageToBytesHandlers = append(packageToBytesHandlers, decryptHandler)

	encryptHandler.SetInitPostHook(func() {
		decryptHandler.SetIv(encryptHandler.GetIv())
		decryptHandler.Init()
	})

	var wg sync.WaitGroup
	wg.Add(2)
	remoteConn, error := GetRemoteConn(remoteAddr, remotePort)
	if error != nil {
		if localConn != nil {
			localConn.Close()
		}
		if remoteConn != nil {
			remoteConn.Close()
		}
		log.Printf("build conn to remote [%v:%v] failed ...", remoteAddr, remotePort)
		wg.Done()
		wg.Done()
	} else {
		//transfer
		go TransferBytesToPackage(localConn, remoteConn, bytesToPackageHandlers, &wg)

		//transfer
		go TransferPackageToBytes(remoteConn, localConn, packageToBytesHandlers, &wg)
	}

	wg.Wait()

	defer func() {
		if localConn != nil {
			localConn.Close()
		}

		if remoteConn != nil {
			remoteConn.Close()
		}
	}()
}
