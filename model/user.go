package model

import (
	"encoding/json"
	"fmt"
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
}

func SetUserLocation(request User) {
	userKey := "quparty_user:uid_" + fmt.Sprintf("%v", request.Id)
	result, err := redis.Redis.Get(userKey).Result()
	var user User
	if err != nil {
		db.Db.Table("fa_user").Select("id,nickname,mobile,avatar,gender").Where("id = ?", request.Id).First(&user)
		fmt.Println(user)
		if user.Id == 0 {
			return
		}
		user.Avatar = utils.CdnUrl(user.Avatar)
		marshal, err := json.Marshal(user)
		if err != nil {
			return
		}
		redis.Redis.Set(userKey, marshal, 1200*time.Second)
	} else {
		var user User
		err := json.Unmarshal([]byte(result), &user)
		if err != nil {
			return
		}
	}
	user.Longitude = request.Longitude
	user.Latitude = request.Latitude
	user.Electricity = request.Electricity
	key := "mapLocation:uid_" + fmt.Sprintf("%v", user.Id)
	marshal, err := json.Marshal(user)
	if err != nil {
		return
	}
	redis.Redis.Set(key, marshal, 300*time.Second)
}

func GetUserLocation(userId uint32) User {
	var user User
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
