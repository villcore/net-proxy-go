package common

import (
	"errors"
	"io"
	"net"
)

type Package struct {
	len       [4]byte
	headerLen [4]byte
	bodyLen   [4]byte
	header    []byte
	body      []byte
}

func NewPackage() *Package {
	return &Package{}
}

func (pkg *Package) GetHeader() []byte {
	return pkg.header
}

func (pkg *Package) GetBody() []byte {
	return pkg.body
}

//readWithHeader
func (pkg *Package) ReadWithHeader(reader net.Conn) (err error) {
	sizeBuf := make([]byte, 4+4+4)
	n, err := io.ReadAtLeast(reader, sizeBuf, 12)

	if n < 0 || n > 1*1024*1024 || err != nil {
		return errors.New("read error")
	}

	headerLen := BytesToInt(sizeBuf[4:8])
	bodyLen := BytesToInt(sizeBuf[8:12])

	total := make([]byte, headerLen+bodyLen)
	n2, err := io.ReadAtLeast(reader, total, headerLen+bodyLen)

	header := total[0:headerLen]
	body := total[headerLen : headerLen+bodyLen]

	copy(pkg.len[:], sizeBuf[:4])
	copy(pkg.headerLen[:], sizeBuf[4:8])
	copy(pkg.bodyLen[:], sizeBuf[8:12])
	pkg.header = header
	pkg.body = body

	if n2 < 0 || err != nil {
		return errors.New("read error")
	}
	return nil
}

//readWithoutHeader
func (pkg *Package) ReadWithoutHeader(reader io.Reader) (err error) {
	buf := make([]byte, 0, 100*1024)
	n, err := reader.Read(buf)

	pkg.body = buf
	pkg.header = make([]byte, 0)

	pkgLen := 0 + n + 12
	headerLen := 0
	bodyLen := n

	copy(pkg.len[:], IntToBytes(pkgLen)[:])
	copy(pkg.headerLen[:], IntToBytes(headerLen)[:])
	copy(pkg.bodyLen[:], IntToBytes(bodyLen)[:])

	if err != nil {
		return err
	} else {
		return nil
	}
}

//value of
func (pkg *Package) ValueOf(header []byte, body []byte) {
	headerLen := len(header)
	bodyLen := len(body)
	pkgLen := headerLen + bodyLen + 12

	copy(pkg.len[:], IntToBytes(pkgLen)[:])
	copy(pkg.headerLen[:], IntToBytes(headerLen)[:])
	copy(pkg.bodyLen[:], IntToBytes(bodyLen)[:])

	pkg.header = header
	pkg.body = body
}

//to bytes
func (pkg *Package) ToBytes() []byte {
	totalBytes := make([]byte, 4+4+4+len(pkg.header)+len(pkg.body))

	copy(totalBytes[:4], pkg.len[:])
	copy(totalBytes[4:8], pkg.headerLen[:])
	copy(totalBytes[8:12], pkg.bodyLen[:])
	copy(totalBytes[12:12+len(pkg.header)], pkg.header[:])
	copy(totalBytes[12+len(pkg.header):12+len(pkg.header)+len(pkg.body)], pkg.body[:])

	return totalBytes
}
