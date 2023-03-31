package stickpackage

import (
	"encoding/binary"
	"io"
)

var VersionLength = 2
var LengthLength = 2
var TcpHeaderLength = VersionLength + LengthLength
var LengthStartLength = 2 //数据长度的起始位置
var LengthEndLength = 4   //数据长度的终止位置
var VersionContent = [2]byte{'v', '1'}

type Stickpackage struct {
	Version [2]byte
	Msg     []byte
	Length  int16
}

func (p *Stickpackage) Pack(writer io.Writer) (err error) {
	err = binary.Write(writer, binary.BigEndian, p.Version)
	err = binary.Write(writer, binary.BigEndian, p.Length)
	err = binary.Write(writer, binary.BigEndian, p.Msg)
	return
}

func (p *Stickpackage) Unpack(reader io.Reader) (err error) {
	err = binary.Read(reader, binary.BigEndian, &p.Version)
	err = binary.Read(reader, binary.BigEndian, &p.Length)
	p.Msg = make([]byte, p.Length-4)
	err = binary.Read(reader, binary.BigEndian, &p.Msg)
	return
}

func (p *Stickpackage) GetPackageLength() int16 {
	p.Length = int16(TcpHeaderLength) + int16(len(p.Msg))
	return p.Length
}
