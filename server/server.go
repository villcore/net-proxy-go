package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"log"
	"net-proxy-go/common"
	"os"
)

const (
	HTTP   = 1
	HTTPS  = 2
	Socks5 = 3
)

func init() {
	log.SetOutput(os.Stdout)
}

//1.接受本地连接
//2.解析包, 解析协议, 解析目的地址
//3.构建远程连接
//4.循环转发(接受包 -> handler处理 -> 发送)
//5.错误处理
//TODO defer recover panic 处理
func AcceptConn(localConn net.Conn, password string) {
	var remoteConn net.Conn
	var wg sync.WaitGroup
	wg.Add(2)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("error, [%v]\n", err)

			if localConn != nil {
				_ = localConn.Close()
			}

			if remoteConn != nil {
				_ = remoteConn.Close()
			}
		}
	}()

	bytesToPackageHandlers, packageToBytesHandlers := initHandlers(password)

	protocolDetected := false
	interrupt := false
	hasError := false

	var buf []byte
	var protocol = -1
	var addr string
	var port int

	//

	if !interrupt || !protocolDetected {
		pkg := *common.NewPackage()
		err := pkg.ReadWithHeader(localConn)

		if err != nil {
			log.Printf("read bytes form conn %v failed...\n", localConn.RemoteAddr())
			interrupt = true
		}

		for _, handler := range packageToBytesHandlers {
			pkg = handler.Handle(&pkg)
		}

		//detect protocol
		body := pkg.GetBody()
		buf = body
		log.Printf("server recv first pkg = %v", string(body))
		protocol = parseProtocol(body, len(body))

		switch protocol {
		case HTTP:
			log.Println("http protocol...")
			break
		case HTTPS:
			log.Println("https protocol...")
			break
		case Socks5:
			log.Println("socks_5 protocol...")
			break
		default:
			log.Println("unrecognized protocol...")
		}

		if protocol > 0 {
			protocolDetected = true
		}
	}

	if !protocolDetected {
		wg.Done()
		wg.Done()

		if localConn != nil {
			_ = localConn.Close()
		}
		hasError = true
	}

	if !hasError {
		parsedAddr, parsedPort, err := parseAddressAndPort(buf, protocol, localConn)
		addr = parsedAddr
		port = parsedPort

		if err != nil {
			wg.Done()
			wg.Done()
			if localConn != nil {
				_ = localConn.Close()
			}
			hasError = true
		}
	}

	log.Printf("need connect to server [%v:%v]", addr, port)

	if !hasError {
		conn, err := common.NewRemoteConn(addr, strconv.Itoa(port))
		remoteConn = conn
		if err != nil {
			wg.Done()
			wg.Done()
			log.Printf("build conn to remote [%v:%v] failed ...", addr, port)
			hasError = true
		} else {
			log.Printf("build conn to remote [%v:%v] success ...", addr, port)
		}
	}

	if !hasError {
		if protocol == HTTPS {
			httpsConnectResp := "HTTP/1.0 200 Connection Established\r\n\r\n"
			httpsRespPkg := *common.NewPackage()
			httpsRespPkg.ValueOf(make([]byte, 0), []byte(httpsConnectResp))

			for _, handler := range bytesToPackageHandlers {
				httpsRespPkg = handler.Handle(&httpsRespPkg)
			}

			_, _ = localConn.Write(httpsRespPkg.ToBytes())
		}

		if protocol == HTTP {
			_, _ = remoteConn.Write(buf)
		}

		if protocol == Socks5 {
		}
	}

	if !hasError {
		//transfer
		go common.TransferPackageToBytes(localConn, remoteConn, packageToBytesHandlers, &wg)

		//transfer
		go common.TransferBytesToPackage(remoteConn, localConn, bytesToPackageHandlers, &wg)
	}

	wg.Wait()
	defer func() {
		if localConn != nil {
			_ = localConn.Close()
		}
		if remoteConn != nil {
			_ = remoteConn.Close()
		}
	}()
}
func initHandlers(password string) ([]common.PackageHandler, []common.PackageHandler) {
	var bytesToPackageHandlers = make([]common.PackageHandler, 0)
	var packageToBytesHandlers = make([]common.PackageHandler, 0)
	//
	cipher, err := common.NewCipher(password)
	if err != nil {
		fmt.Println("init cipher error ...")
	}
	encryptHandler := common.NewEncryptHandler(cipher)
	bytesToPackageHandlers = append(bytesToPackageHandlers, encryptHandler)
	//
	decryptHandler := common.NewDecryptHandler(cipher)
	packageToBytesHandlers = append(packageToBytesHandlers, decryptHandler)
	decryptHandler.SetInitPostHook(func() {
		encryptHandler.SetIv(decryptHandler.GetIv())
		encryptHandler.Init()
	})
	_ = decryptHandler
	return bytesToPackageHandlers, packageToBytesHandlers
}

func parseProtocol(req []byte, len int) int {
	//HTTPS
	if len >= 7 {
		headerInfo := string(req[:7])
		if strings.EqualFold("CONNECT", headerInfo) {
			return HTTPS
		}
	}

	//HTTP
	httpOpePos := -1
	for index, val := range req {
		if index >= len {
			break
		}
		if int(val) == 32 {
			httpOpePos = index
			break
		}
	}

	if httpOpePos != -1 {
		requestMethod := string(req[:httpOpePos])
		log.Println("requestMethod = ", requestMethod)
		if strings.EqualFold(requestMethod, "GET") || strings.EqualFold(requestMethod, "POST") {
			return HTTP
		}
	}
	return -1
}

func parseAddressAndPort(firstReq []byte, protocol int, localConn net.Conn) (addr string, port int, err error) {
	switch protocol {
	case HTTP:
		return parseHttpAddress(firstReq)
		break
	case HTTPS:
		return parseHttpsAddress(firstReq)
		break
	case Socks5:
		break
	default:
		log.Println("unrecognized protocol")
	}
	return "", -1, errors.New("unrecognized protocol")
}

func parseHttpsAddress(firstReq []byte) (addr string, port int, err error) {
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(string(firstReq))))

	if err != nil {
		return "", -1, errors.New("unrecognized proctol")
	}

	addrAndPort := req.Host
	infos := strings.Split(addrAndPort, ":")

	addr = infos[0]
	port = 443

	if len(infos) > 1 {
		port, err = strconv.Atoi(infos[1])
	}

	fmt.Printf("addr = %v, port = %v\n", addr, port)
	return addr, port, nil
}

func parseHttpAddress(firstReq []byte) (addr string, port int, err error) {
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(string(firstReq))))

	if err != nil {
		return "", -1, errors.New("unrecognized protocol")
	}

	addrAndPort := req.Host
	infos := strings.Split(addrAndPort, ":")

	addr = infos[0]
	port = 80

	if len(infos) > 1 {
		port, err = strconv.Atoi(infos[1])
	}

	fmt.Printf("addr = %v, port = %v\n", addr, port)
	return addr, port, nil
}
