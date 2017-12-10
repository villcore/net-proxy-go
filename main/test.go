package main

import (
	"../encrypt"
	"fmt"
	"../conf"
	"os"
	"io"
	"crypto/rand"
)

func sliceModify(slice *[]int) {
	// slice[0] = 88
	*slice = append(*slice, 6)
}

func main() {
	slice := []int{1, 2, 3, 4, 5}
	sliceModify(&slice)
	fmt.Println(slice)

	testStr := "CONNECT www.baidu.com:443 HTTP/1.1\r\nHost: www.baidu.com:443\r\nUser-Agent: curl/7.53.1\r\nProxy-Connection: Keep-Alive\r\n\r\n"

	//testStr := "HTTP/1.0 200 Connection Established\r\n\r\n"

	cipher, err := encrypt.NewCipher("villcore")
	iv, err := cipher.InitEncrypt()
	if err != nil {
		fmt.Println("init cipher error ...")
	}

	fmt.Println("iv = ", iv)
	cipher.InitDecrypt(iv)
	if err != nil {
		fmt.Println("init cipher error ...")
	}

	fmt.Println("size = {}", len([]byte(testStr)))
	eBytes := make([]byte, len(testStr))
	dBytes := make([]byte, len(testStr))

	cipher.Encrypt(eBytes, []byte(testStr))
	fmt.Println(string(eBytes))

	cipher.Decrypt(dBytes, eBytes)

	fmt.Println(string(dBytes))

	file, err := os.Create("d://encrypt.dat")
	defer file.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	file2, err2 := os.Create("d://iv.dat")
	defer file2.Close()

	if err2 != nil {
		fmt.Println(err2)
		return
	}

	file.Write(eBytes)
	file2.Write(iv)
	file.Sync()

	//json
	clientConfig, _ := conf.ReadClientConf("client.conf")
	fmt.Println(clientConfig)

	serverConfig, _ := conf.ReadServerConf("server.conf")
	fmt.Println(len(serverConfig.PortPair))

	testDefer(0)
	testDefer(1)

	i := 0
	defer fmt.Println(i)
	i++

	fmt.Println("new iv ...")
	newIv := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, newIv); err != nil {
		fmt.Println(err)
	}

	fmt.Println(newIv)
}

func testDefer(num int) (int) {
	defer func() {
		fmt.Println("end...")
	}()
	if num%2 == 0 {
		return 0
	}

	fmt.Println(num)
	return 1

	return 1
}
