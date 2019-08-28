package main

import (
	"fmt"
	"net"
	"os"

	//"github.com/villcore/net-proxy-go/client"
	"../client"
	"../conf"
	"log"
)

func main() {
	clientConf, err := conf.ReadClientConf("client.conf")
	if err != nil {
		fmt.Println("can not load conf file ...")
		return
	}

	listenPort := clientConf.LocalPort
	remoteAddr := clientConf.RemoteAddr
	remotePort := clientConf.RemotePort
	password := clientConf.Password

	fmt.Print("local client start...\n")
	//
	log.SetOutput(os.Stdout)

	listenAddr := ":" + listenPort
	log.Printf("[%v]", listenAddr)

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Printf("starting listen failed at port [%v] ...\n", listenPort)
		log.Println(err)
		return
	}

	addr := listener.Addr()
	log.Printf("staring listen address : [%v] ...\n", addr.String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept conn [%v] failed ...\n", conn.LocalAddr())
		}
		log.Printf("accept conn [%v] success ...\n", conn.RemoteAddr())

		go client.AcceptConn(conn, remoteAddr, remotePort, password)
	}
}
