package model

import (
	"encoding/json"
	"fmt"
	"time"
	"websocket/config"
	"websocket/lib/db"
	"websocket/lib/mylog"
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
	ChooseType  int `gorm:"-"`
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
		if user.Id == 0 {
			return
		}
		if user.GhostType == 1 {
			redis.Redis.Del(mpaKey)
			redis.Redis.Del(userKey)
			return
		}
		if user.GhostType == 2 || user.GhostType == 3 {
			if user.GhostTime > int(time.Now().Unix()) {
				redis.Redis.Del(mpaKey)
				redis.Redis.Del(userKey)
				return
			}
		}
		//user.Avatar = utils.CdnUrl(user.Avatar) + "?x-oss-process=image/resize,w_100,m_lfit"
		user.Avatar = utils.RoundThumb(user.Avatar, "100", "0")
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
	var base = &config.Base{}
	err = config.ConfFile.Section("base").MapTo(base)
	if err != nil {
		mylog.Error("获取配置参数不正确:" + err.Error())
		return
	}
	redis.Redis.Set(mpaKey, marshal, base.MapLocationExpire*time.Second)
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

// user 缓存
func SetUserInfo(user User) {
	key := "mapUserInfo:uid_" + fmt.Sprintf("%v", user.Id)
	result, err := redis.Redis.Get(key).Result()
	if err != nil {
		return
	}
	var resUser User
	err = json.Unmarshal([]byte(result), &resUser)
	if err != nil {
		return
	}

}
