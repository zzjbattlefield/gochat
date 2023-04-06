package connect

import (
	"sync"
	"sync/atomic"

	"github.com/zzjbattlefield/IM_GO/proto"
)

type Bucket struct {
	lock       sync.RWMutex
	rooms      map[int]*Room
	chs        map[int]ChannelClients
	option     BucketOption
	routines   []chan *proto.PushRoomMessageReqeust //发送群组消息的channel 接收到消息后直接丢进redis里
	routineNum uint64
	broadcast  chan []byte
}

type BucketOption struct {
	ChannelSize    int //chs的size
	RoomSize       int
	routinueAmount uint64 // 存放投递push消息的通道切片的数量
	routinueSize   int    // push消息通道的缓冲区大小
}

func NewBucket(option *BucketOption) (bucket *Bucket) {
	bucket = &Bucket{
		rooms:    make(map[int]*Room),
		chs:      make(map[int]ChannelClients),
		option:   *option,
		routines: make([]chan *proto.PushRoomMessageReqeust, option.routinueAmount),
	}
	for i := uint64(0); i < option.routinueAmount; i++ {
		messageChan := make(chan *proto.PushRoomMessageReqeust, option.routinueSize)
		bucket.routines[i] = messageChan
		go bucket.pushRoom(messageChan)
	}
	return
}

// 将消息发送到指定的房间
func (b *Bucket) pushRoom(ch chan *proto.PushRoomMessageReqeust) {
	for {
		var (
			arg  *proto.PushRoomMessageReqeust
			room *Room
		)
		arg = <-ch
		if room = b.Room(arg.RoomID); room != nil {
			room.Push(&arg.Msg)
		}
	}

}

func (b *Bucket) Room(rid int) (room *Room) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.rooms[rid]
}

func (bucket *Bucket) DeleteChannel(ch *Channel) {
	var (
		ok   bool
		room *Room
	)
	bucket.lock.RLock()
	defer bucket.lock.RUnlock()
	//先获取这个用户的所有客户端Channel标识
	userClient, ok := bucket.chs[ch.UserID]
	if ok {
		//再判断这个这个客户端是否存在于这个用户的所有客户端中
		if ch, ok = userClient[ch.UUID]; ok {
			room = ch.Room
			delete(userClient, ch.UUID)
			if len(userClient) == 0 {
				//如果这个用户的所有客户端都已经断开了 那么就把这个用户从bucket里删掉
				delete(bucket.chs, ch.UserID)
			}
		}
	}
	if room != nil && room.DeleteChannel(ch) {
		//也要把room里的这个channel给删掉
		if room.drop {
			delete(bucket.rooms, room.ID)
		}
	}
}

func (bucket *Bucket) Put(userID int, roomID int, ch *Channel) (err error) {
	var (
		room *Room
		ok   bool
	)
	bucket.lock.Lock()
	if roomID != noRoom {
		if room, ok = bucket.rooms[roomID]; !ok {
			//创建新房间
			room = NewRoom(roomID)
			bucket.rooms[roomID] = room
		}
		ch.Room = room
	}
	ch.UserID = userID
	if _, ok = bucket.chs[userID]; !ok {
		//创建新的客户端hash表
		bucket.chs[userID] = NewChannelClients()
	}
	bucket.chs[userID][ch.UUID] = ch
	bucket.lock.Unlock()

	if room != nil {
		room.Put(ch)
	}
	return
}

// 广播整个房间
func (bucket *Bucket) BroadcastRoom(req *proto.PushRoomMessageReqeust) {
	num := atomic.AddUint64(&bucket.routineNum, 1) % bucket.option.routinueAmount
	bucket.routines[num] <- req
}
