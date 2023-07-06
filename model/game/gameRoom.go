package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"websocket/config"
	"websocket/impl"
	"websocket/lib/db"
	"websocket/lib/mylog"
	"websocket/lib/redis"
	"websocket/model/comm"
	"websocket/utils"
)

// 房间结构
type GameRoom struct {
	Id              string `gorm:"id"`
	UserId          uint32 `gorm:"user_id"`
	MasterId        uint32 `gorm:"master_id"`
	HidingSecond    int    `gorm:"hiding_second"`
	ArrestSecond    int    `gorm:"arrest_second"`
	Status          int    `gorm:"status"`
	StartHidingTime int    `gorm:"start_hiding_time"`
	StartArrestTime int    `gorm:"start_arrest_time"`
	EndTime         int    `gorm:"end_time"`
	Coordinate      string `gorm:"coordinate"`
	ChatRoom        string `gorm:"chat_room"`
	CreateTime      int64  `gorm:"create_time"`
	CloseTime       int64  `gorm:"close_time"`
}
type CreateRoomReq struct {
	HidingSecond int                      `json:"hiding_second"`
	ArrestSecond int                      `json:"arrest_second"`
	Coordinate   []map[string]interface{} `json:"coordinate"`
}

func CreateRoom(request impl.IRequest, userId uint32) (GameRoom, error) {
	var room GameRoom
	roomRes := CreateRoomReq{}
	var resp comm.ResponseMsg
	resp.MsgId = 205
	err := json.Unmarshal(request.GetData(), &roomRes)
	conn := request.GetConnection()
	if err != nil {
		resp.Code = 0
		resp.Msg = "请求参数不正确"
		comm.SendMsg(conn, 205, resp)
		return room, err
	}
	if roomRes.HidingSecond == 0 {
		resp.Code = 0
		resp.Msg = "请填写正确的躲藏时间"
		comm.SendMsg(conn, 205, resp)
		return room, err
	}
	if roomRes.ArrestSecond == 0 {
		resp.Code = 0
		resp.Msg = "请填写正确的抓捕时间"
		comm.SendMsg(conn, 205, resp)
		return room, err
	}
	if len(roomRes.Coordinate) < 3 {
		resp.Code = 0
		resp.Msg = "至少设置3个范围坐标"
		comm.SendMsg(conn, 205, resp)
		return room, err
	}
	// 获取随机数字

	var roomId string
	for {
		roomId = utils.RandNumString(4)
		db.Db.Table("fa_game_room").
			Where("id = ?", roomId).
			First(&room)
		if room.Id == "" {
			break
		}
	}
	coordinate, err := json.Marshal(roomRes.Coordinate)
	if err != nil {
		resp.Code = 0
		resp.Msg = "范围坐标参数异常"
		comm.SendMsg(conn, 205, resp)
		return room, err
	}
	room.Id = roomId
	room.UserId = userId
	room.MasterId = userId
	room.HidingSecond = roomRes.HidingSecond
	room.ArrestSecond = roomRes.ArrestSecond
	room.Coordinate = string(coordinate)
	room.Status = 0
	room.CreateTime = time.Now().Unix()
	// 创建聊天室
	var conf = &config.Conf{}
	err = config.ConfFile.Section("conf").MapTo(conf)
	if err != nil {
		mylog.Error("配置参数错误:" + err.Error())
		return room, err
	}
	//chatRoomId, err := CreateChatRoom(conf.WebAdmin+"/app/other/createGroup", userId, "定位捉迷藏", "定位捉迷藏游戏")
	//if err != nil {
	//	resp.Code = 0
	//	resp.Msg = err.Error()
	//	comm.SendMsg(request, 205, resp)
	//	return room, err
	//}
	//room.ChatRoom = chatRoomId
	var player = GamePlayer{
		UserId:     userId,
		RoomId:     roomId,
		Role:       0,
		Status:     0,
		CreateTime: time.Now().Unix(),
	}
	tx := db.Db.Begin()
	if err = tx.Table("fa_game_room").Create(room).Error; err != nil {
		tx.Rollback()
		return room, err
	}
	log.Printf("fa_game_player:%+v", player)
	if err = tx.Table("fa_game_player").Create(&player).Error; err != nil {
		tx.Rollback()
		return room, err
	}
	tx.Commit()
	resp.Code = 1
	resp.Msg = "success"
	resp.Data = room
	comm.SendMsg(conn, 205, resp)
	return room, nil
}

type CreateChatRoomResp struct {
	Code int
	Msg  string
	Data string
}

func GetRoomCache(roomId string) (GameRoom, error) {
	key := "gameRoom:roomId_" + roomId
	var room GameRoom
	result, err := redis.Redis.Get(key).Result()
	if err == nil {
		err = json.Unmarshal([]byte(result), &room)
		if err != nil {
			return room, err
		}
		return room, nil
	}
	db.Db.Table("fa_game_room").
		Where("id = ?", roomId).
		First(&room)
	if room.Id == "" {
		return room, errors.New("empty data")
	}
	marshal, err := json.Marshal(room)
	if err != nil {
		return room, err
	}
	redis.Redis.Set(key, marshal, 30*time.Second)
	return room, nil
}

// ClearRoomCache 清除房间缓存
func ClearRoomCache(roomId string) {
	key := "gameRoom:roomId_" + roomId
	redis.Redis.Del(key)
}

// CreateChatRoom 创建聊天室
func CreateChatRoom(url string, userId uint32, groupname, desc string) (string, error) {
	params := make(map[string]interface{})
	params["user_id"] = userId
	params["groupname"] = groupname
	params["desc"] = desc
	bytesData, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	res, err := http.Post(url,
		"application/json;charset=utf-8", bytes.NewBuffer(bytesData))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var resp CreateChatRoomResp
	err = json.Unmarshal(content, &resp)
	if err != nil {
		return "", err
	}
	return resp.Data, nil
}

// CloseRoom 关闭游戏
func CloseRoom(roomId string) {
	room, err := GetRoomCache(roomId)
	if err != nil {
		mylog.Error("获取游戏房间信息不正确,roomId:" + roomId)
		return
	}
	// 修改游戏状态
	room.Status = 4
	room.CloseTime = time.Now().Unix()
	tx := db.Db.Begin()
	if err = tx.Table("fa_game_room").Save(&room).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Table("fa_game_player").Where("room_id = ? AND status=?", room.Id, 0).
		Updates(GamePlayer{Status: 4}).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	ClearRoomCache(roomId)
	ClearPlayersCache(roomId)
}

// EndRoom 结束游戏
func EndRoom(roomId string) {
	room, err := GetRoomCache(roomId)
	if err != nil {
		mylog.Error("获取游戏房间信息不正确,roomId:" + roomId)
		return
	}
	// 修改游戏状态
	room.Status = 4
	room.CloseTime = time.Now().Unix()
	tx := db.Db.Begin()
	if err = tx.Table("fa_game_room").Save(&room).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Table("fa_game_player").Where("room_id = ? AND status=?", room.Id, 0).
		Updates(GamePlayer{Status: 4}).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	ClearRoomCache(roomId)
	ClearPlayersCache(roomId)
}
