package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
)

const (
	IvLen = 16
)

type Cipher struct {
	enc cipher.Stream
	dec cipher.Stream
	key []byte
	iv  []byte
}

func NewCipher(password string) (cipher *Cipher, err error) {
	return &Cipher{key: generateKeyUsePassword(password, IvLen)}, nil
}

func md5Sum(src []byte) []byte {
	h := md5.New()
	h.Write(src)
	return h.Sum(nil)
}

func generateKeyUsePassword(password string, keyLen int) (key []byte) {
	const md5Len = 16

	cnt := (keyLen-1)/md5Len + 1
	m := make([]byte, cnt*md5Len)
	copy(m, md5Sum([]byte(password)))

	// Repeatedly call md5 until bytes generated is enough.
	// Each call to md5 uses data: prev md5 sum + password.
	d := make([]byte, md5Len+len(password))
	start := 0
	for i := 1; i < cnt; i++ {
		start += md5Len
		copy(d, m[start-md5Len:start])
		copy(d[md5Len:], password)
		copy(m[start:], md5Sum(d))
	}
	return m[:keyLen]
}

// Initializes the block cipher with CFB mode, returns IV.
func (c *Cipher) InitEncrypt() (iv []byte, err error) {
	if c.iv == nil || len(c.iv) == 0 {
		c.iv = make([]byte, IvLen)
		if _, err := io.ReadFull(rand.Reader, c.iv); err != nil {
			return nil, err
		}
		iv = c.iv
	} else {
		iv = c.iv
	}

	block, err := aes.NewCipher(c.key)
	c.enc, err = cipher.NewCFBEncrypter(block, iv), nil

	return iv, nil
}

func (c *Cipher) SetIv(iv []byte) {
	c.iv = iv
}

func (c *Cipher) InitDecrypt(iv []byte) (err error) {
	c.iv = iv

	block, err := aes.NewCipher(c.key)
	if err != nil {
		fmt.Println("create block error ...")
		return err
	}
	c.dec, err = cipher.NewCFBDecrypter(block, iv), nil
	return nil
}

func (c *Cipher) Encrypt(dst, src []byte) {
	c.enc.XORKeyStream(dst, src)
}

func (c *Cipher) Decrypt(dst, src []byte) {
	c.dec.XORKeyStream(dst, src)
}
