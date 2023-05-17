package model

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"websocket/lib/db"
	"websocket/lib/redis"
	"websocket/utils"
)

type User struct {
	Id          uint32
	Nickname    string
	Mobile      string
	Avatar      string
	Gender      int
	Longitude   string
	Latitude    string
	Electricity string
	GhostType   int `gorm:"ghost_type"`
	GhostTime   int `gorm:"ghost_time"`
}

func SetUserLocation(request User) {
	userKey := "quparty_user:uid_" + fmt.Sprintf("%v", request.Id)
	mpaKey := "mapLocation:uid_" + fmt.Sprintf("%v", request.Id)
	result, err := redis.Redis.Get(userKey).Result()
	var user User
	if err == nil {
		err := json.Unmarshal([]byte(result), &user)
		if err != nil {
			return
		}
	} else {
		db.Db.Table("fa_user").
			Select("id,nickname,mobile,avatar,gender,ghost_type,ghost_time").
			Where("id = ?", request.Id).
			First(&user)
		fmt.Printf("%#v\n", user)
		if user.Id == 0 {
			return
		}
		if user.GhostType == 1 {
			redis.Redis.Del(mpaKey)
			return
		}
		if user.GhostType == 2 || user.GhostType == 3 {
			if user.GhostTime > int(time.Now().Unix()) {
				redis.Redis.Del(mpaKey)
				return
			}
		}
		user.Avatar = utils.CdnUrl(user.Avatar)
		marshal, err := json.Marshal(user)
		if err != nil {
			return
		}
		redis.Redis.Set(userKey, marshal, 600*time.Second)
	}
	user.Longitude = request.Longitude
	user.Latitude = request.Latitude
	user.Electricity = request.Electricity
	marshal, err := json.Marshal(user)
	if err != nil {
		return
	}
	redis.Redis.Set(mpaKey, marshal, 300*time.Second)
	log.Println("-----------上传定位成功------------")
	log.Printf("%+v", user)
}

func GetUserLocation(userId uint32) User {
	user := User{}
	key := "mapLocation:uid_" + fmt.Sprintf("%v", userId)
	result, err := redis.Redis.Get(key).Result()
	if err != nil {
		return user
	} else {
		err := json.Unmarshal([]byte(result), &user)
		if err != nil {
			return user
		}
	}
	return user
}
