package main

import (
	"fmt"
	"log"
	"net"
	"net-proxy-go/common"
	"net-proxy-go/server"
	"os"
	"sync"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	serverConfig, err := common.ReadServerConf("server.conf")
	if err != nil {
		fmt.Println("can not load conf file ...")
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(serverConfig.PortPair))

	for _, portAndPair := range serverConfig.PortPair {
		go startListen(portAndPair)
	}
	wg.Wait()
}

func startListen(portAndPassword common.PortAndPassword) {
	fmt.Println(portAndPassword)
	listenPort := portAndPassword.ListenPort
	listenAddrAndPort := ":" + listenPort
	password := portAndPassword.Password
	fmt.Println("start listen port ", listenPort)
	log.Println("server start...")
	listener, err := net.Listen("tcp", listenAddrAndPort)
	if err != nil {
		log.Printf("erver starting listen failed at port [%v] ...\n", listenPort)
	}

	addr := listener.Addr()
	log.Printf("server staring listen address : [%v] ...\n", addr.String())

	for {
		localConn, err := listener.Accept()
		if err != nil {
			log.Printf("accept conn [%v] failed ...\n\n", localConn.LocalAddr())
		}
		log.Printf("accept conn [%v] success ...\n\n", localConn.RemoteAddr())

		go server.AcceptConn(localConn, password)
	}
}
