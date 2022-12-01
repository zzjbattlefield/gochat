package connect

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zzjbattlefield/IM_GO/config"
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

func (s *Service) writePump(ch *Channel, c *Connect) {
	ticker := time.NewTicker(s.Option.PingPeriod)
	defer func() {
		ticker.Stop()
		ch.conn.Close()
	}()
}

// 接收前端发送的ws连接 主要实现注册Connect的功能
func (s *Service) readPump(ch *Channel, c *Connect) {
	defer func() {
		//断线的时候要清空一下redis里的连接信息
		config.Zap.Infoln("start disConnect")

	}()
}
