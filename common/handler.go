package common

import (
	"fmt"
)

type PackageHandler interface {
	Handle(pkg *Package) (newPkg Package)
}

//encrypt handler
type EncryptHandler struct {
	encrypt      Cipher
	iv           []byte
	init         bool
	initPostHook func()
}

func (encryptHandler *EncryptHandler) SetInitPostHook(f func()) {
	encryptHandler.initPostHook = f
}

func (encryptHandler *EncryptHandler) GetIv() []byte {
	return encryptHandler.iv
}

func (encryptHandler *EncryptHandler) SetIv(iv []byte) {
	encryptHandler.iv = iv
}

func (encryptHandler *EncryptHandler) Init() {
	encryptHandler.init = true
	encryptHandler.encrypt.SetIv(encryptHandler.iv)
	_, _ = encryptHandler.encrypt.InitEncrypt()
}

func (encryptHandler *EncryptHandler) Handle(pkg *Package) (newPkg Package) {
	header := pkg.GetHeader()
	body := pkg.GetBody()

	var nPkg Package
	var encryptHeader []byte
	encryptBody := make([]byte, len(body))

	if !encryptHandler.init {
		//init
		iv, err := encryptHandler.encrypt.InitEncrypt()
		if err != nil {
			fmt.Println("init encrypt handler error ...")
			iv = make([]byte, 16)
		} else {

		}

		encryptHandler.iv = iv
		encryptHeader = make([]byte, len(header)+4+len(iv))

		tmp := make([]byte, len(header))
		encryptHandler.encrypt.Encrypt(tmp, header)
		encryptHandler.encrypt.Encrypt(encryptBody, body)

		copy(encryptHeader[:4], IntToBytes(len(iv))[:])
		copy(encryptHeader[4:4+len(iv)], iv[:])
		copy(encryptHeader[4+len(iv):], tmp[:])
		encryptHandler.init = true
		encryptHandler.initPostHook()
	} else {
		if len(header) != 0 {
			encryptHeader = make([]byte, len(header))
			encryptHandler.encrypt.Encrypt(encryptHeader, header)
		}

		if len(body) != 0 {
			encryptHandler.encrypt.Encrypt(encryptBody, body)
		}
	}

	nPkg.ValueOf(encryptHeader, encryptBody)
	return nPkg
}

func NewEncryptHandler(cipher *Cipher) *EncryptHandler {
	return &EncryptHandler{encrypt: *cipher}
}

//decrypt handler
type DecryptHandler struct {
	decrypt      Cipher
	iv           []byte
	init         bool
	initPostHook func()
}

func NewDecryptHandler(cipher *Cipher) *DecryptHandler {
	return &DecryptHandler{decrypt: *cipher}
}

func (decryptHandler *DecryptHandler) SetInitPostHook(f func()) {
	decryptHandler.initPostHook = f
}

func (decryptHandler *DecryptHandler) GetIv() []byte {
	return decryptHandler.iv
}

func (decryptHandler *DecryptHandler) SetIv(iv []byte) {
	decryptHandler.iv = iv
}

func (decryptHandler *DecryptHandler) Init() {
	decryptHandler.init = true
	_ = decryptHandler.decrypt.InitDecrypt(decryptHandler.iv)
}

func (decryptHandler *DecryptHandler) Handle(pkg *Package) (newPkg Package) {
	header := pkg.GetHeader()
	body := pkg.GetBody()

	var nPkg Package
	var decryptHeader []byte
	decryptBody := make([]byte, len(body))

	if !decryptHandler.init {
		lvLen := BytesToInt(header[:4])
		iv := header[4 : 4+lvLen]

		tmp := make([]byte, len(iv))
		copy(tmp[:], iv[:])

		decryptHandler.iv = tmp
		_ = decryptHandler.decrypt.InitDecrypt(tmp)

		decryptHeader = make([]byte, len(header)-4-len(iv))
		decryptHandler.decrypt.Decrypt(decryptHeader, header[4+len(iv):])
		decryptHandler.decrypt.Decrypt(decryptBody, body)

		nPkg = *NewPackage()
		nPkg.ValueOf(decryptHeader, decryptBody)

		decryptHandler.init = true
		decryptHandler.initPostHook()
	} else {
		decryptHeader = make([]byte, len(header))
		if len(header) != 0 {
			decryptHandler.decrypt.Decrypt(decryptHeader, header[:])
		}

		if len(body) != 0 {
			decryptHandler.decrypt.Decrypt(decryptBody, body)
		}
	}

	nPkg.ValueOf(decryptHeader, decryptBody)
	return nPkg
}

type CompressHandler struct {
}

type DecompressHandler struct {
}
