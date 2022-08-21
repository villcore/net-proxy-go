package common

import (
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
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
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		log.Printf("connect %v failed. %v \n", address, err)
		return false
	}
	defer func() {
		_ = conn.Close()
	}()
	log.Printf("connect %v succeed. \n", address)
	return true
}
