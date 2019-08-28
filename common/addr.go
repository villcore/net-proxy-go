package common

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type Address struct {
	M          sync.Mutex
	Accessible map[string]bool
}

func NewAddress() *Address {
	return &Address{Accessible: make(map[string]bool)}
}

func (addr *Address) setAccessible(address string, accessible bool) {
	addr.M.Lock()
	defer addr.M.Unlock()
	addr.Accessible[address] = accessible
}

func (addr *Address) getAccessible(address string) (bool, bool2 bool) {
	addr.M.Lock()
	defer addr.M.Unlock()

	host, _ := addr.hostAndPort(address)
	accessible, ok := addr.Accessible[host]
	return accessible, ok
}

func (addr *Address) clear() {
	addr.M.Lock()
	defer addr.M.Unlock()
	addr.Accessible = make(map[string]bool)
}

func (addr *Address) IsAccessible(address string) bool {
	host, port := addr.hostAndPort(address)
	accessible, ok := addr.getAccessible(address)
	if !ok {
		accessible = connect(host + ":" + port)
		addr.setAccessible(host, accessible)
	}
	return accessible
}

func (addr *Address) hostAndPort(address string) (string, string) {
	hostAndPort := strings.Split(address, ":")
	if len(hostAndPort) == 1 {
		return hostAndPort[0], string(80)
	}
	var host, port string
	host = hostAndPort[0]
	port = hostAndPort[1]
	return host, port
}

func connect(address string) bool {
	conn, err := net.Dial("tcp", address) //创建套接字,连接服务器,设置超时时间
	if err != nil {
		fmt.Println("x {}", err)
		return false
	}
	defer conn.Close()
	return true
}
