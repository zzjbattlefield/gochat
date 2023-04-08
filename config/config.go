package config

import (
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var once sync.Once
var realPath string
var Conf *Config
var Zap *zap.SugaredLogger

const (
	FailReplyCode         = 1
	SuccessReplyCode      = 0
	RedisBaseValidTime    = 86400
	QueueName             = "chat_queue"
	RedisPrefix           = "chat_"
	RedisUserfix          = "chat_user_"
	RedisRoomPrefix       = "chat_room_"
	RedisRoomOnlinePrefix = "chat_room_online_count_"
	OpSingleSend          = 2 // single user
	OpRoomSend            = 3 // send to room
	OpRoomCountSend       = 4 // get online user count
	OpRoomInfoSend        = 5 // send info to room
	OpBulidTcpConn        = 6
)

type Config struct {
	LogicConfig LogicConfig
	Common      Common
	Connect     ConnectConfig
	ApiConfig   ApiConfig
	TaskConfig  TaskConfig
}

type LogicConfig struct {
	LogicBase LogicBase `mapstructure:"logic-base"`
}

type LogicBase struct {
	RpcAddress string `mapstructure:"rpcAddress"`
}

type ConnectConfig struct {
	ConnectBase                ConnectBase                `mapstructure:"connect-base"`
	ConnectBucket              ConnectBucket              `mapstructure:"connect-bucket"`
	ConnectWebSocket           ConnectWebSocket           `mapstructure:"connect-websocket"`
	ConnectRpcAddressWebSocket ConnectRpcAddressWebSocket `mapstructure:"connect-rpcAddress-websocket"`
	ConnectTcp                 ConnectTcp                 `mapstructure:"connect-tcp"`
	ConnectRpcAddressTcp       ConnectRpcAddressTcp       `mapstructure:"connect-rpcAddress-tcp"`
}

type ConnectTcp struct {
	ServerId      string `mapstructure:"serverId"`
	Bind          string `mapstructure:"bind"`
	SendBuf       int    `mapstructure:"sendbuf"`
	ReceiveBuf    int    `mapstructure:"receivebuf"`
	KeepAlive     bool   `mapstructure:"keepalive"`
	Reader        int    `mapstructure:"reader"`
	ReadBuf       int    `mapstructure:"readBuf"`
	ReadBufSize   int    `mapstructure:"readBufSize"`
	Writer        int    `mapstructure:"writer"`
	WriterBuf     int    `mapstructure:"writerBuf"`
	WriterBufSize int    `mapstructure:"writeBufSize"`
}
type ConnectRpcAddressWebSocket struct {
	Address string `mapstructure:"address"`
}

type ConnectRpcAddressTcp struct {
	Address string `mapstructure:"address"`
}

type ConnectBase struct {
}

type ConnectBucket struct {
	CpuNum        int    `mapstructure:"cpuNum"`
	Channel       int    `mapstructure:"channel"`
	Room          int    `mapstructure:"room"`
	SrvProto      int    `mapstructure:"svrProto"`
	RoutineAmount uint64 `mapstructure:"routineAmount"`
	RoutineSize   int    `mapstructure:"routineSize"`
}

type ConnectWebSocket struct {
	Bind string `mapstructure:"bind"`
}

type ApiConfig struct {
	ApiBase ApiBase `mapstructure:"api-base"`
}

type ApiBase struct {
	ListenPort int `mapstructure:"listenPort"`
}

type Common struct {
	CommonMysql CommonMysql `mapstructure:"common-mysql"`
	CommonRedis CommonRedis `mapstructure:"common-redis"`
	CommonEtcd  CommonEtcd  `mapstructure:"common-etcd"`
}

type CommonEtcd struct {
	Host              string `mapstructure:"host"`
	BasePath          string `mapstructure:"basePath"`
	ServerPathLogic   string `mapstructure:"serverPathLogic"`
	ServerPathConnect string `mapstructure:"serverPathConnect"`
	UserName          string `mapstructure:"userName"`
	Password          string `mapstructure:"password"`
	ConnectionTimeout int    `mapstructure:"connectionTimeout"`
}

type CommonMysql struct {
	Port     int    `mapstructure:"mysqlPort"`
	Address  string `mapstructure:"mysqlAddress"`
	Password string `mapstructure:"mysqlPassword"`
	User     string `mapstructure:"mysqlUser"`
}

type CommonRedis struct {
	RedisAddress  string `mapstructure:"redisAddress"`
	RedisPassword string `mapstructure:"redisPassword"`
	Db            int    `mapstructure:"db"`
}

type TaskConfig struct {
	TaskBase TaskBase `mapstructure:"task-base"`
}

type TaskBase struct {
	RedisAddress  string `mapstructure:"redisAddress"`
	RedisPassword string `mapstructure:"redisPassword"`
	RpcAddress    string `mapstructure:"rpcAddress"`
	PushChan      int    `mapstructure:"pushChan"`
	PushChanSize  int    `mapstructure:"pushChanSize"`
}

// 获取config的文件夹路径
func getCurrentDir() string {
	_, file, _, _ := runtime.Caller(1)
	path := strings.Split(file, "/")
	dir := strings.Join(path[0:len(path)-1], "/")
	return dir
}

func GetMode() string {
	env := os.Getenv("IM_GO_MODE")
	if env == "" {
		env = "dev"
	}
	return env
}

func init() {
	Init()
}

func Init() {
	once.Do(func() {
		Zap = zap.NewExample().Sugar()
		env := GetMode()
		realPath = getCurrentDir()
		configPath := realPath + "/" + env + "/"
		viper.SetConfigType("toml")
		viper.AddConfigPath(configPath)
		viper.SetConfigName("/common")
		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}
		viper.SetConfigName("/api")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}
		viper.SetConfigName("/connect")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}
		viper.SetConfigName("/logic")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}
		viper.SetConfigName("/task")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}
		Conf = new(Config)
		viper.Unmarshal(&Conf.Common)
		viper.Unmarshal(&Conf.ApiConfig)
		viper.Unmarshal(&Conf.Connect)
		viper.Unmarshal(&Conf.TaskConfig)
		viper.Unmarshal(&Conf.LogicConfig)
	})
}
