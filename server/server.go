package server

import (
	"net"
	"strings"
	"sync"
	"errors"
	"strconv"
	"fmt"
	"net/http"
	"bufio"

	//"github.com/villcore/net-proxy-go/common"
	"../common"
	"../encrypt"
	"os"
	"log"
)

const (
	HTTP    = 1
	HTTPS   = 2
	SOCKS_5 = 3
)

func init() {
	log.SetOutput(os.Stdout)
}

var(
	dncryptHandler2 *common.DecryptHandler
)
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
			log.Println("error ... [%v]", err)

			if localConn != nil {
				localConn.Close()
			}

			if remoteConn != nil {
				remoteConn.Close()
			}
		}
	}()

	bytesToPackageHandlers, packageToBytesHandlers := initHandlers(password)

	protocalDetected := false
	interrupt := false
	hasError := false

	var buf []byte
	var protocal int = -1
	var addr string
	var port int

	//

	if !interrupt || !protocalDetected {
		pkg := *common.NewPackage()
		err := pkg.ReadWithHeader(localConn)

		if err != nil {
			log.Printf("read bytes form conn %v failed...\n", localConn.RemoteAddr())
			interrupt = true
		}

		for _, handler := range packageToBytesHandlers {
			pkg = handler.Handle(&pkg)
		}

		//detect protocal
		body := pkg.GetBody()
		buf = body
		//log.Printf("server recv first pkg = %v", string(body))
		protocal = parseProtocal(body, len(body))

		switch protocal {
		case HTTP:
			log.Println("http protocal...")
			break
		case HTTPS:
			log.Println("https protocal...")
			break
		case SOCKS_5:
			log.Println("socks_5 protocal...")
			break
		default:
			log.Println("unrecognized protocal...")
		}

		if protocal > 0 {
			protocalDetected = true
		}
	}

	if !protocalDetected {
		wg.Done()
		wg.Done()

		if localConn != nil {
			localConn.Close()
		}
		hasError = true
	}

	if !hasError {
		parsedAddr, parsedPort, err := parseAddressAndPort(buf, protocal, localConn)
		addr = parsedAddr
		port = parsedPort

		if err != nil {
			wg.Done()
			wg.Done()
			if localConn != nil {
				localConn.Close()
			}
			hasError = true
		}
	}

	log.Printf("need connect to server [%v:%v]", addr, port)

	if !hasError {
		conn, error := common.NewRemoteConn(addr, strconv.Itoa(port))
		remoteConn = conn
		if error != nil {
			wg.Done()
			wg.Done()
			log.Printf("build conn to remote [%v:%v] failed ...", addr, port)
			hasError = true
		} else {
			log.Printf("build conn to remote [%v:%v] success ...", addr, port)
		}
	}

	if !hasError {
		if (protocal == HTTPS) {
			httpsConnectResp := "HTTP/1.0 200 Connection Established\r\n\r\n"
			httpsRespPkg := *common.NewPackage()
			httpsRespPkg.ValueOf(make([]byte, 0), []byte(httpsConnectResp))


			for _, handler := range bytesToPackageHandlers {
				httpsRespPkg = handler.Handle(&httpsRespPkg)
			}
			fmt.Println("https resp ori     = ", len(httpsRespPkg.GetBody()))

			localConn.Write(httpsRespPkg.ToBytes())
		}

		if (protocal == HTTP) {
			remoteConn.Write(buf)
		}

		if (protocal == SOCKS_5) {
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
			localConn.Close()
		}
		if remoteConn != nil {
			remoteConn.Close()
		}
	}()
}
func initHandlers(password string) ([]common.PackageHandler, []common.PackageHandler) {
	//init handlers start ...
	var bytesToPackageHandlers []common.PackageHandler = make([]common.PackageHandler, 0)
	var packageToBytesHandlers []common.PackageHandler = make([]common.PackageHandler, 0)
	//
	cipher, err := encrypt.NewCipher(password)
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
	dncryptHandler2 = decryptHandler
	//init handlers end ...
	return bytesToPackageHandlers, packageToBytesHandlers
}

func parseProtocal(req []byte, len int) int {
	//TODO SOCKS_5

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
		//fmt.Println(val, int(val))
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

func parseAddressAndPort(firstReq []byte, protocal int, localConn net.Conn) (addr string, port int, err error) {
	//log.Printf("read content = \n%v ...\n", string(firstReq))
	switch protocal {
	case HTTP:
		return parseHttpAddress(firstReq)
		break
	case HTTPS:
		return parseHttpsAddress(firstReq)
		break
	case SOCKS_5:
		break
	default:
		log.Println("unrecognized proctol ...")
	}
	return "", -1, errors.New("unrecognized proctol...")
}

func parseHttpsAddress(firstReq []byte) (addr string, port int, err error) {
	//fmt.Printf("first line = \n%v\n", string(firstReq))
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(string(firstReq))))

	if err != nil {
		return "", -1, errors.New("unrecognized proctol...")
	}

	addrAndPort := req.Host
	infos := strings.Split(addrAndPort, ":")

	addr = infos[0]
	port = 443

	if (len(infos) > 1) {
		port, err = strconv.Atoi(infos[1])
	}

	fmt.Printf("addr = %v, port = %v\n", addr, port)
	return addr, port, nil
}

func parseHttpAddress(firstReq []byte) (addr string, port int, err error) {
	//fmt.Printf("first line = \n%v\n", string(firstReq))
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(string(firstReq))))

	if err != nil {
		return "", -1, errors.New("unrecognized proctol...")
	}

	addrAndPort := req.Host
	infos := strings.Split(addrAndPort, ":")

	addr = infos[0]
	port = 80

	if (len(infos) > 1) {
		port, err = strconv.Atoi(infos[1])
	}

	fmt.Printf("addr = %v, port = %v\n", addr, port)
	return addr, port, nil
}
