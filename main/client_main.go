package main

import (
	"fmt"
	"net"
	"os"

	//"github.com/villcore/net-proxy-go/client"
	"../client"
	"../conf"
	"log"
	"os/exec"
	"os/signal"
)

func main() {
	clientConf, err := conf.ReadClientConf("client.conf")
	if err != nil {
		fmt.Println("can not load conf file ...")
		return
	}

	//set windows proxy
	cmd := exec.Command("win_utils\\sysproxy.exe", "global", "127.0.0.1:50081")
	cmd.Start()

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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	s := <-c
	fmt.Println("Got signal:", s)
	clearProxy := exec.Command("win_utils\\sysproxy.exe", "set", "1")
	clearProxy.Start()
}
