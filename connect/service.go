package connect

import (
	"time"
)

type Service struct {
	Buckets     []*Bucket
	bucketIndex uint32
	Option      ServiceOption
	operator    Operator
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

func NewService(b []*Bucket, o Operator, option ServiceOption) *Service {
	return &Service{
		Buckets:     b,
		Option:      option,
		bucketIndex: uint32(len(b)),
		operator:    o,
	}
}
