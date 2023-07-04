package model

import (
	"encoding/json"
	"fmt"
	"time"
	"websocket/config"
	"websocket/impl"
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
type UserType struct {
	Type   string
	RoomId string
}

func SetUserLocation(request User) {
	userKey := "quparty_user:uid_" + fmt.Sprintf("%v", request.Id)
	mpaKey := "mapLocation:uid_" + fmt.Sprintf("%v", request.Id)
	tempMpaKey := "tempMapLocation:uid_" + fmt.Sprintf("%v", request.Id)
	result, err := redis.Redis.Get(userKey).Result()
	var user User
	isGhost := false
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
		// 处理幽灵模式
		if user.GhostType == 1 {
			redis.Redis.Del(mpaKey)
			isGhost = true
		}
		if user.GhostType == 2 || user.GhostType == 3 {
			if user.GhostTime > int(time.Now().Unix()) {
				redis.Redis.Del(mpaKey)
				isGhost = true
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
	if !isGhost {
		redis.Redis.Set(mpaKey, marshal, base.MapLocationExpire*time.Second)
	}
	redis.Redis.Set(tempMpaKey, marshal, 60*time.Second)
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

func GetUserTempLocation(userId uint32) User {
	user := User{}
	key := "tempMapLocation:uid_" + fmt.Sprintf("%v", userId)
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

// userType 缓存
func SetUserType(request impl.IRequest, userType UserType) {
	uid := request.GetConnection().GetConnID()
	request.GetConnection().SetProperty("type", userType.Type)
	if userType.RoomId != "" {
		request.GetConnection().SetProperty("roomId", userType.RoomId)
	}
	key := "UserType:uid_" + fmt.Sprintf("%v", uid)
	marshal, err := json.Marshal(userType)
	if err != nil {
		return
	}
	redis.Redis.Set(key, marshal, 600*time.Second)
}

func GetUserType(uid uint32) UserType {
	key := "UserType:uid_" + fmt.Sprintf("%v", uid)
	userType := UserType{}
	result, err := redis.Redis.Get(key).Result()
	if err != nil {
		return userType
	} else {
		err := json.Unmarshal([]byte(result), &userType)
		if err != nil {
			return userType
		}
	}
	return userType
}
