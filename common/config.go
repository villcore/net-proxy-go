package common

import (
	"encoding/json"
	"errors"
	"os"
)

type ClientConfig struct {
	LocalPort  string `json:"local_port"`
	RemoteAddr string `json:"remote_addr"`
	RemotePort string `json:"remote_port"`
	Password   string `json:"password"`
}

type ServerConfig struct {
	PortPair []PortAndPassword `json:"port_pair"`
}

type PortAndPassword struct {
	ListenPort string `json:"listen_port"`
	Password   string `json:"password"`
}

func ReadServerConf(path string) (serverConfig *ServerConfig, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("parse conf error")
	}

	info, err := f.Stat()
	if err != nil {
		return nil, errors.New("parse conf error")
	}

	fileSize := info.Size()

	bytes := make([]byte, fileSize)
	_, _ = f.Read(bytes)

	serverConfig = &ServerConfig{}

	err = json.Unmarshal(bytes, &serverConfig)
	if err != nil {
		return nil, errors.New("parse conf error")
	}
	return serverConfig, nil
}

func ReadClientConf(path string) (clientConfig *ClientConfig, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("parse conf error")
	}

	info, err := f.Stat()
	if err != nil {
		return nil, errors.New("parse conf error")
	}

	fileSize := info.Size()

	bytes := make([]byte, fileSize)
	_, _ = f.Read(bytes)

	clientConfig = &ClientConfig{}

	err = json.Unmarshal(bytes, &clientConfig)
	if err != nil {
		return nil, errors.New("parse conf error")
	}
	return clientConfig, nil
}
