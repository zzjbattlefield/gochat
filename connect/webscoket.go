package connect

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/proto"
	"github.com/zzjbattlefield/IM_GO/tools"
)

func (c *Connect) initWebsocket() error {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c.serveWs(DefaultService, w, r)
	})
	err := http.ListenAndServe(config.Conf.Connect.ConnectWebSocket.Bind, nil)
	return err
}

func (c *Connect) serveWs(s *Service, w http.ResponseWriter, r *http.Request) {
	upGrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upGrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		config.Zap.Errorln("serveWs error:", err.Error())
	}
	ch := NewChannel(1024)
	ch.conn = conn
	go s.writePump(ch, c)
	go s.readPump(ch, c)
}

// 通过hash来确定这个用户属于哪个bucket
func (s *Service) Bucket(userID int) (bucket *Bucket) {
	userIDStr := strconv.Itoa(userID)
	index := tools.CityHash32([]byte(userIDStr), uint32(len(userIDStr))) % s.bucketIndex
	return s.Buckets[index]
}

func (s *Service) writePump(ch *Channel, c *Connect) {
	ticker := time.NewTicker(s.Option.PingPeriod)
	defer func() {
		ticker.Stop()
		ch.conn.Close()
	}()
	for {
		select {
		case message := <-ch.broadcast:
			err := ch.conn.SetWriteDeadline(time.Now().Add(s.Option.WriteWait))
			if err != nil {
				config.Zap.Errorln("设置websocket writeDeadline失败:", err.Error())
				ch.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := ch.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				config.Zap.Errorln("创建nextWrite失败")
			}
			_, err = w.Write(message.Body)
			if err != nil {
				w.Close()
			}
		case <-ticker.C:
			//ping一下客户端
			err := ch.conn.SetWriteDeadline(time.Now().Add(s.Option.WriteWait))
			if err != nil {
				config.Zap.Errorln("设置websocket writeDeadline失败:", err.Error())
				ch.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
		}
	}
}

// 接收前端发送的ws连接 主要实现注册Connect的功能
func (s *Service) readPump(ch *Channel, c *Connect) {
	defer func() {
		//断线的时候要清空一下redis里的连接信息
		config.Zap.Infoln("start disConnect")
		if ch.Room == nil || ch.UserID == 0 {
			config.Zap.Infoln("room and userid eq 0")
			ch.conn.Close()
			return
		}
		config.Zap.Infoln("exec disConnect...")
		disConnectReq := new(proto.DisConnectRequest)
		disConnectReq.RoomID = ch.Room.ID
		disConnectReq.UserID = ch.UserID
		s.Bucket(ch.UserID).DeleteChannel(ch)
		if err := s.DisConnect(disConnectReq); err != nil {
			config.Zap.Errorf("disConnect error :%v", err)
		}
		ch.conn.Close()
	}()
	for {
		_, message, err := ch.conn.ReadMessage()
		if err != nil {
			websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)
			config.Zap.Errorln("ws read message error:", err)
			return
		}
		if message == nil {
			return
		}
		var connRequest = new(proto.ConnectRequest)
		config.Zap.Infoln("get message:", string(message))
		if err = json.Unmarshal(message, &connRequest); err != nil {
			config.Zap.Errorf("message struct:%v, error is:%v", &connRequest, err.Error())
		}
		if connRequest.AuthToken == "" || connRequest == nil {
			config.Zap.Errorln("readPump message no authToken")
		}
		connRequest.ServiceID = c.ServiceID
		userID, err := s.Connect(connRequest)
		if err != nil {
			config.Zap.Errorf("ws read connect err:%v", err)
		}
		//把用户id放到bucket里
		bucket := s.Bucket(userID)
		err = bucket.Put(userID, connRequest.RoomID, ch)
		if err != nil {
			ch.conn.Close()
			config.Zap.Errorln("conn close err:", err.Error())
		}
	}
}
