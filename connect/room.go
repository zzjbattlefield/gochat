package connect

import "sync"

type Room struct {
	ID          int
	rlock       sync.RWMutex
	OnlineCount int
	next        *Channel
	drop        bool //如果房间已经没有连接了为true bucket会把房间删掉
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
