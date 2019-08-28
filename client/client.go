package client

import (
	"net"
	"os"
	"sync"

	//"github.com/villcore/net-proxy-go/common"
	"../common"
	"../encrypt"
	"fmt"
	"log"
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
	cipher, err := encrypt.NewCipher(password)
	if err != nil {
		fmt.Println("init cipher error ...")
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
	remoteConn, error := common.GetRemoteConn(remoteAddr, remotePort)

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
		go common.TransferBytesToPackage(localConn, remoteConn, bytesToPackageHandlers, &wg)

		//transfer
		go common.TransferPackageToBytes(remoteConn, localConn, packageToBytesHandlers, &wg)
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
