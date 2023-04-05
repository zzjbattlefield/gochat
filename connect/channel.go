package connect

import (
	"net"

	"github.com/gorilla/websocket"
	"github.com/zzjbattlefield/IM_GO/proto"
)

type Channel struct {
	UserID    int
	Room      *Room
	broadcast chan *proto.Message
	conn      *websocket.Conn
	connTcp   *net.TCPConn
	next      *Channel
	prev      *Channel
}

func NewChannel(size int) *Channel {
	return &Channel{
		broadcast: make(chan *proto.Message, size),
		next:      nil,
		prev:      nil,
	}
}

func (c *Channel) Push(msg *proto.Message) (err error) {
	select {
	case c.broadcast <- msg:
	default:
	}
	return
}
