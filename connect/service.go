package connect

import (
	"time"

	"github.com/zzjbattlefield/IM_GO/proto"
)

type Service struct {
	ID          string
	Buckets     []*Bucket
	bucketIndex int
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
		bucketIndex: len(b),
	}
}

func (s *Service) Connect(request proto.ConnectRequest) {

}

func (s *Service) DisConnect() {

}
