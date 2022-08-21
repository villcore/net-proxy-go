package common

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	HTTP  = 1
	HTTPS = 2
)

func ParseHttpOrHttps(b []byte, len int) (string, int, error) {
	switch parseProtocol(b, len) {
	case HTTP:
		return parseHttpAddress(b)
	case HTTPS:
		return parseHttpsAddress(b)
	}
	return "", 0, errors.New("unknown protocol")
}

func parseProtocol(req []byte, len int) int {
	// HTTPS
	if len >= 7 {
		headerInfo := string(req[:7])
		if strings.EqualFold("CONNECT", headerInfo) {
			return HTTPS
		}
	}

	//
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
		// TODO: http 其他方法
		if strings.EqualFold(requestMethod, "GET") || strings.EqualFold(requestMethod, "POST") {
			return HTTP
		}
	}

	// log.Printf("unknown protocol \n %v \n", string(req[0:len]))
	return -1
}

func parseHttpsAddress(firstReq []byte) (addr string, port int, err error) {
	//fmt.Printf("first line = \n%v\n", string(firstReq))
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(string(firstReq))))

	if err != nil {
		return "", -1, errors.New("unreconized proctocol")
	}

	addrAndPort := req.Host
	infos := strings.Split(addrAndPort, ":")

	addr = infos[0]
	port = 443

	if len(infos) > 1 {
		port, err = strconv.Atoi(infos[1])
	}
	return addr, port, nil
}

func parseHttpAddress(firstReq []byte) (addr string, port int, err error) {
	//fmt.Printf("first line = \n%v\n", string(firstReq))
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(string(firstReq))))

	if err != nil {
		return "", -1, errors.New("unrecognized proctol")
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
