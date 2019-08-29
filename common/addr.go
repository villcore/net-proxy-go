package common

import (
	"log"
	"net"
	"os"
	"strconv"
	"sync"
)

func init() {
	log.SetOutput(os.Stdout)
}

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
	accessible, ok := addr.Accessible[address]
	return accessible, ok
}

func (addr *Address) clear() {
	addr.M.Lock()
	defer addr.M.Unlock()
	addr.Accessible = make(map[string]bool)
}

func (addr *Address) IsAccessible(address string, port int) bool {
	accessible, ok := addr.getAccessible(address)
	if !ok {
		accessible = connect(address + ":" + strconv.Itoa(port))
		addr.setAccessible(address, accessible)
	}
	return accessible
}

func connect(address string) bool {
	conn, err := net.Dial("tcp", address) //创建套接字,连接服务器,设置超时时间
	if err != nil {
		log.Printf("connect %v failed. \n", address)
		return false
	}
	defer func() {
		_ = conn.Close()
	}()
	log.Printf("connect %v succeed. \n", address)
	return true
}
