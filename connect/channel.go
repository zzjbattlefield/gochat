package connect

import (
	"github.com/gorilla/websocket"
	"github.com/zzjbattlefield/IM_GO/proto"
)

type Channel struct {
	UserID    int
	Room      *Room
	broadcast chan *proto.Message
	conn      *websocket.Conn
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
