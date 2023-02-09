package db

import (
	"strconv"

	"github.com/zzjbattlefield/IM_GO/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	InitDB("im_go")
}

func InitDB(DBName string) {
	var err error
	mysqlConfig := config.Conf.Common.CommonMysql
	dsn := mysqlConfig.User + ":" + mysqlConfig.Password + "@tcp(" + mysqlConfig.Address + ":" + strconv.Itoa(mysqlConfig.Port) + ")/" + DBName + "?parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}
