package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/zzjbattlefield/IM_GO/db"
)

type UserModel struct {
	ID        int `gorm:"primary_key"`
	UserName  string
	Password  string
	CreatedAt time.Time
}

func (user *UserModel) CheckHaveUserName(userName string) *UserModel {
	data := &UserModel{}
	db.DB.Debug().Table("user").Where("user_name = ?", userName).First(data)
	fmt.Println(data)
	return data
}

func (user *UserModel) Add() (userID int, err error) {
	if user.UserName == "" || user.Password == "" {
		return 0, errors.New("用户名和密码不能为空")
	}
	user.CreatedAt = time.Now()
	if err := db.DB.Table("user").Create(user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (user *UserModel) GetUserInfoByUserId(userID int) (err error) {
	result := db.DB.Table("user").Where("id = ?", userID).First(user)
	return result.Error
}
