package stickpackage

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
	"testing"
	"time"

	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/proto"
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

func Test_TcpClient(t *testing.T) {
	var (
		tcpClient net.Conn
		err       error
	)
	roomId := 1
	fromUserID := 5
	authToken := "UN_kcQQBv7MRZI9GYfD5OA4FxCXZ058xw6BvU3TqTZY="
	tcpAddrRemote, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:7001")
	tcpClient, err = net.DialTCP("tcp", nil, tcpAddrRemote)
	defer func() {
		_ = tcpClient.Close()
	}()
	if err != nil {
		panic("conn err:" + err.Error())
	}
	//读取服务端广播的信息
	go func() {
		scanner := bufio.NewScanner(tcpClient)
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
		failTime := 0
		for {
			log.Println("start read tcp msg from conn...")
			failTime++
			if failTime >= 3 {
				log.Println("fail to many time")
				break
			}
			for scanner.Scan() {
				log.Println("read tcp msg from conn ok")
				scannerPackage := &Stickpackage{}
				if err := scannerPackage.Unpack(bytes.NewReader(scanner.Bytes())); err != nil {
					log.Printf("packer unpack err:%s", err.Error())
					break
				}
				log.Printf("read msg from tcp ok version:%s length:%d msg:%s", scannerPackage.Version, scannerPackage.Length, scannerPackage.Msg)
			}
			if scanner.Err() != nil {
				log.Printf("scanner err:%s", scanner.Err().Error())
				break
			}
		}
	}()
	tcpReq := &proto.SendTcp{
		Op:         config.OpBulidTcpConn,
		RoomId:     roomId,
		FromUserId: fromUserID,
		AuthToken:  authToken,
	}
	tcpReqBytes, _ := json.Marshal(tcpReq)
	tcpPack := &Stickpackage{
		Version: VersionContent,
		Msg:     tcpReqBytes,
	}
	tcpPack.Length = tcpPack.GetPackageLength()
	tcpPack.Pack(tcpClient)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			messageBody := "now time:" + time.Now().Format("2006-01-02 15:04:05")
			tcpRep := &proto.SendTcp{
				Op:         config.OpRoomSend,
				Msg:        messageBody,
				RoomId:     roomId,
				FromUserId: fromUserID,
				AuthToken:  authToken,
			}
			reqJson, _ := json.Marshal(tcpRep)
			pack := &Stickpackage{
				Version: VersionContent,
				Msg:     reqJson,
			}
			pack.Length = pack.GetPackageLength()
			pack.Pack(tcpClient)
			log.Printf("send msg to tcp ok")
		}
	}()

	time.Sleep(time.Minute * 30)
}
