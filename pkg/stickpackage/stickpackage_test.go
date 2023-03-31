package stickpackage

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"log"
	"testing"
	"time"
)

func Test_TestStick(t *testing.T) {
	pack := &Stickpackage{
		Version: VersionContent,
		Msg:     []byte("now time:" + time.Now().Format("2006-01-02 15:04:05")),
	}
	pack.Length = pack.GetPackageLength()
	buff := bytes.NewBuffer(make([]byte, 0))
	pack.Pack(buff)
	pack.Pack(buff)
	pack.Pack(buff)
	pack.Pack(buff)

	scanner := bufio.NewScanner(buff)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if !atEOF && data[0] == 'v' && len(data) > 4 {
			packSumLength := int16(0)
			binary.Read(bytes.NewReader(data[2:4]), binary.BigEndian, &packSumLength)
			if int(packSumLength) <= len(data) {
				return int(packSumLength), data[:packSumLength], nil
			}
		}
		return
	})

	pack = &Stickpackage{}
	for scanner.Scan() {
		if err := pack.Unpack(bytes.NewReader(scanner.Bytes())); err != nil {
			log.Println(err.Error())
		}
		log.Println(string(pack.Msg))
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("invalid data")
		t.Fail()
	}
}
