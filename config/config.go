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
	FailReplyCode    = 1
	SuccessReplyCode = 0
)

type Config struct {
	Common    Common
	ApiConfig ApiConfig
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
		Conf = new(Config)
		viper.Unmarshal(&Conf.Common)
		viper.Unmarshal(&Conf.ApiConfig)
	})
}
