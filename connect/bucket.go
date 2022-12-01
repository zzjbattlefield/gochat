package connect

import (
	"sync"

	"github.com/zzjbattlefield/IM_GO/proto"
)

type Bucket struct {
	lock                 sync.RWMutex
	rooms                map[int]*Room
	chs                  map[int]*Channel
	option               BucketOption
	pushRoomMesssageChan []chan *proto.PushRoomMessageReqeust //发送群组消息的channel 接收到消息后直接丢进redis里
}

type BucketOption struct {
	ChannelSize              int //chs的size
	pushRoomMesssageChanSize int //发送群组消息的channel的size
	routineNum               int //发送message的goroutine个数
}

func NewBucket(option *BucketOption) (bucket *Bucket) {
	bucket = &Bucket{
		rooms:                make(map[int]*Room),
		chs:                  make(map[int]*Channel, option.ChannelSize),
		option:               *option,
		pushRoomMesssageChan: make([]chan *proto.PushRoomMessageReqeust, option.routineNum),
	}
	for i := 0; i < option.pushRoomMesssageChanSize; i++ {
		messageChan := make(chan *proto.PushRoomMessageReqeust, option.pushRoomMesssageChanSize)
		bucket.pushRoomMesssageChan[i] = messageChan
		go pushRoomMessageToQueue(messageChan)
	}
	return
}

func pushRoomMessageToQueue(ch chan *proto.PushRoomMessageReqeust) {
	//TODO:将message推送到队列中
}

func (bucket *Bucket) DeleteChannel(ch *Channel) {
	var (
		ok   bool
		room *Room
	)
	bucket.lock.RLock()
	defer bucket.lock.RUnlock()
	if ch, ok = bucket.chs[ch.UserID]; ok {
		room = ch.Room
		delete(bucket.chs, ch.UserID)
	}
	if room != nil && room.DeleteChannel(ch) {
		//也要把room里的这个channel给删掉
		if room.drop {
			delete(bucket.rooms, room.ID)
		}
	}
}
