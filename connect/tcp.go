package connect

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/smallnest/rpcx/log"
	apirpc "github.com/zzjbattlefield/IM_GO/api/rpc"
	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/pkg/stickpackage"
	"github.com/zzjbattlefield/IM_GO/proto"
)

func init() {
	apirpc.InitLogicRpcClient()
}

func (c *Connect) initTcpServer() (err error) {
	var (
		addr     *net.TCPAddr
		listener *net.TCPListener
	)
	cpuNum := config.Conf.Connect.ConnectBucket.CpuNum
	tcpAddr := strings.Split(config.Conf.Connect.ConnectTcp.Bind, ",")
	for _, ipPort := range tcpAddr {
		if addr, err = net.ResolveTCPAddr("tcp", ipPort); err != nil {
			config.Zap.Fatalf("initTcpServer ResolveTCPAddr err:%s", err.Error())
		}
		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			config.Zap.Fatalf("initTcpServer ListenTCP err:%s", err.Error())
		}
		config.Zap.Infof("start listen tcp at : %s", ipPort)
		for i := 0; i < cpuNum; i++ {
			go c.accectTcp(listener)
		}
	}
	return
}

func (c *Connect) accectTcp(lintener *net.TCPListener) {
	connectTcpConfig := config.Conf.Connect.ConnectTcp
	var (
		conn *net.TCPConn
		err  error
	)
	for {
		if conn, err = lintener.AcceptTCP(); err != nil {
			config.Zap.Errorf("AcceptTCP error:%s ", err.Error())
		}
		if err = conn.SetKeepAlive(connectTcpConfig.KeepAlive); err != nil {
			config.Zap.Errorf("SetKeepAlive error:%s ", err.Error())
		}
		if err = conn.SetReadBuffer(connectTcpConfig.ReceiveBuf); err != nil {
			config.Zap.Errorf("SetReadBuffer error:%s ", err.Error())
		}
		if err = conn.SetWriteBuffer(connectTcpConfig.SendBuf); err != nil {
			config.Zap.Errorf("SetWriteBuffer error:%s ", err.Error())
		}
		go c.ServeTcp(DefaultService, conn)
	}
}

func (c *Connect) ServeTcp(server *Service, conn *net.TCPConn) {
	var ch *Channel
	ch = NewChannel(server.Option.BroadcastSize)
	ch.connTcp = conn
	go c.readDataFromTcp(server, ch)
	go c.writeDataToTcp(server, ch)
}

func (c *Connect) readDataFromTcp(s *Service, ch *Channel) {
	defer func() {
		config.Zap.Info("start exec disTcpConn")
		if ch.Room == nil || ch.UserID == 0 {
			ch.connTcp.Close()
			return
		}
		config.Zap.Info("exec disconnect...")
		disconnReq := &proto.DisConnectRequest{
			RoomID: ch.Room.ID,
			UserID: ch.UserID,
		}
		s.Bucket(ch.UserID).DeleteChannel(ch)
		if err := s.operator.DisConnect(disconnReq); err != nil {
			config.Zap.Errorf("disConnect error :%s", err.Error())
		}
		ch.connTcp.Close()
	}()
	scanner := bufio.NewScanner(ch.connTcp)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if !atEOF && data[0] == 'v' && len(data) >= stickpackage.TcpHeaderLength {
			packMsgLen := int16(0)
			binary.Read(bytes.NewReader(data[stickpackage.LengthStartLength:stickpackage.LengthEndLength]), binary.BigEndian, &packMsgLen)
			if int(packMsgLen) <= len(data) {
				return int(packMsgLen), data[:packMsgLen], nil
			}
		}
		return
	})
	for scanner.Scan() {
		scannerPacker := new(stickpackage.Stickpackage)
		if err := scannerPacker.Unpack(bytes.NewReader(scanner.Bytes())); err != nil {
			config.Zap.Errorf("scan unpack tcp error:%s", err.Error())
			return
		}
		config.Zap.Info("get a msg", string(scannerPacker.Msg))
		var rawTcpMsg proto.SendTcp
		if err := json.Unmarshal(scannerPacker.Msg, &rawTcpMsg); err != nil {
			config.Zap.Errorf("Unmarshal tcp message error:%s", err.Error())
			return
		}
		config.Zap.Infof("rawTcpMsg is %+v", rawTcpMsg)
		if rawTcpMsg.AuthToken == "" || rawTcpMsg.RoomId <= 0 {
			config.Zap.Errorf("wrong params")
			return
		}
		switch rawTcpMsg.Op {
		case config.OpBulidTcpConn:
			connReq := &proto.ConnectRequest{
				AuthToken: rawTcpMsg.AuthToken,
				RoomID:    rawTcpMsg.RoomId,
				ServiceID: c.ServiceID,
			}
			userID, err := s.operator.Connect(connReq)
			if err != nil {
				config.Zap.Errorf("bulid tcp conn err:%s", err.Error())
				return
			} else if userID == 0 {
				config.Zap.Errorln("empty userID")
				return
			}
			config.Zap.Infof("conn success userID=%d", userID)
			pack := &stickpackage.Stickpackage{
				Version: stickpackage.VersionContent,
			}
			pack.Msg = []byte("hello client")
			pack.Length = pack.GetPackageLength()
			config.Zap.Info("write msg to tcp conn", string(pack.Msg))
			if err := pack.Pack(ch.connTcp); err != nil {
				config.Zap.Errorf("write test msg to tcp conn err:%s", err.Error())
				return
			}
			bucket := s.Bucket(userID)
			if err = bucket.Put(userID, rawTcpMsg.RoomId, ch); err != nil {
				config.Zap.Errorf("bucket put user error:%s", err.Error())
				return
			}
		case config.OpRoomSend:
			req := &proto.Send{
				Msg:          rawTcpMsg.Msg,
				RoomId:       rawTcpMsg.RoomId,
				FromUserId:   ch.UserID,
				FromUserName: rawTcpMsg.FromUserName,
			}
			code, msg := apirpc.RpcLoginObj.PushRoom(req)
			config.Zap.Info("tcp conn push msg to room", code, msg)
		}
		if scanner.Err() != nil {
			config.Zap.Errorf("scanner err:%s", scanner.Err().Error())
			return
		}
	}
}

func (c *Connect) writeDataToTcp(s *Service, ch *Channel) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		ch.connTcp.Close()
	}()
	pack := &stickpackage.Stickpackage{
		Version: stickpackage.VersionContent,
	}
	for {
		select {
		case msg, ok := <-ch.broadcast:
			if !ok {
				ch.connTcp.Close()
			}
			pack.Msg = msg.Body
			pack.Length = pack.GetPackageLength()
			log.Info("write msg to tcp conn", string(msg.Body))
			if err := pack.Pack(ch.connTcp); err != nil {
				config.Zap.Errorf("write msg to tcp conn err:%s", err.Error())
				return
			}
		case <-ticker.C:
			config.Zap.Info("send tcp ping message")
			pack.Msg = []byte("ping msg")
			pack.Length = pack.GetPackageLength()
			if err := pack.Pack(ch.connTcp); err != nil {
				return
			}
		}
	}
}
