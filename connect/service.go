package connect

import (
	"time"

	"github.com/zzjbattlefield/IM_GO/proto"
)

type Service struct {
	Buckets     []*Bucket
	bucketIndex uint32
	Option      ServiceOption
}

type ServiceOption struct {
	WriteWait       time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	MaxMessageSize  int64
	ReadBufferSize  int
	WriteBufferSize int
	BroadcastSize   int
}

func NewService(b []*Bucket, option ServiceOption) *Service {
	return &Service{
		Buckets:     b,
		Option:      option,
		bucketIndex: uint32(len(b)),
	}
}

func (s *Service) Connect(request *proto.ConnectRequest) (userID int, err error) {
	connRpc := new(ConnectRpc)
	userID, err = connRpc.Connect(request)
	return
}

func (s *Service) DisConnect(request *proto.DisConnectRequest) error {
	connRpc := new(ConnectRpc)
	err := connRpc.DisConnect(request)
	return err
}
