package client

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"net-proxy-go/common"
)

func init() {
	log.SetOutput(os.Stdout)
}

//1.接受本地连接
//2.构建远程连接,(可用连接连接池)
//3.循环转发(接受包 -> handler处理 -> 发送)
//4.错误处理
//TODO defer panic recover 处理
func AcceptConn(localConn net.Conn, remoteAddr string, remotePort string, password string) {

	var bytesToPackageHandlers = make([]common.PackageHandler, 0)
	var packageToBytesHandlers = make([]common.PackageHandler, 0)
	//
	cipher, err := common.NewCipher(password)
	if err != nil {
		fmt.Println("init cipher err")
	}

	encryptHandler := common.NewEncryptHandler(cipher)
	bytesToPackageHandlers = append(bytesToPackageHandlers, encryptHandler)

	//
	decryptHandler := common.NewDecryptHandler(cipher)
	packageToBytesHandlers = append(packageToBytesHandlers, decryptHandler)

	encryptHandler.SetInitPostHook(func() {
		decryptHandler.SetIv(encryptHandler.GetIv())
		decryptHandler.Init()
	})

	var wg sync.WaitGroup
	wg.Add(2)
	remoteConn, err := common.GetRemoteConn(remoteAddr, remotePort)

	if err != nil {
		if localConn != nil {
			_ = localConn.Close()
		}
		if remoteConn != nil {
			_ = remoteConn.Close()
		}
		log.Printf("build conn to remote [%v:%v] failed ...", remoteAddr, remotePort)
		wg.Done()
		wg.Done()
	} else {
		//transfer
		go common.TransferBytesToPackage(localConn, remoteConn, bytesToPackageHandlers, &wg)

		//transfer
		go common.TransferPackageToBytes(remoteConn, localConn, packageToBytesHandlers, &wg)
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
