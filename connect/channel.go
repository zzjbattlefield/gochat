package connect

import (
	"net"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zzjbattlefield/IM_GO/proto"
)

type ChannelClients map[string]*Channel

type Channel struct {
	UUID      string
	UserID    int
	Room      *Room
	broadcast chan *proto.Message
	conn      *websocket.Conn
	connTcp   *net.TCPConn
	next      *Channel
	prev      *Channel
}

func NewChannelClients() ChannelClients {
	return make(map[string]*Channel, 0)
}

func NewChannel(size int) *Channel {
	return &Channel{
		UUID:      uuid.New().String(),
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
