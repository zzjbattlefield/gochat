package connect

import (
	"sync"

	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/proto"
)

const noRoom = -1

type Room struct {
	ID          int
	rlock       sync.RWMutex
	OnlineCount int
	next        *Channel
	drop        bool //如果房间已经没有连接了为true bucket会把房间删掉
}

func NewRoom(roomID int) *Room {
	room := new(Room)
	room.ID = roomID
	room.drop = false
	return room
}

// 把channel从双向链表中删除
func (r *Room) DeleteChannel(ch *Channel) bool {
	r.rlock.RLock()
	defer r.rlock.RUnlock()
	if ch.next != nil {
		ch.next.prev = ch.prev
	}
	if ch.prev != nil {
		ch.prev.next = ch.next
	} else {
		r.next = ch.next
	}
	r.drop = false
	r.OnlineCount--
	if r.OnlineCount <= 0 {
		r.drop = true
	}

	return r.drop
}

// 遍历room下的所有用户把msg传给他们
func (r *Room) Push(msg *proto.Message) {
	r.rlock.RLock()
	defer r.rlock.RUnlock()
	for ch := r.next; ch != nil; ch = ch.next {
		if err := ch.Push(msg); err != nil {
			config.Zap.Errorf("push msg err:%v", err.Error())
		}
	}
}

// 将ch加到双链表中
func (r *Room) Put(ch *Channel) (err error) {
	r.rlock.Lock()
	defer r.rlock.Unlock()
	if r.next != nil {
		r.next.prev = ch
	}
	ch.next = r.next
	ch.prev = nil
	r.next = ch
	r.OnlineCount++
	return
}
