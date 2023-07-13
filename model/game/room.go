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
type Room struct {
	Id              string `gorm:"id"`
	UserId          uint32 `gorm:"user_id"`
	MasterId        uint32 `gorm:"master_id"`
	HidingSecond    int    `gorm:"hiding_second"`
	ArrestSecond    int    `gorm:"arrest_second"`
	Status          int    `gorm:"status"`
	StartHidingTime int64  `gorm:"start_hiding_time"`
	StartArrestTime int64  `gorm:"start_arrest_time"`
	EndTime         int64  `gorm:"end_time"`
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

func CreateRoom(request impl.IRequest, userId uint32) (Room, error) {
	var room Room
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
		return room, errors.New("请填写正确的躲藏时间")
	}
	if roomRes.ArrestSecond == 0 {
		resp.Code = 0
		resp.Msg = "请填写正确的抓捕时间"
		comm.SendMsg(conn, 205, resp)
		return room, errors.New("请填写正确的抓捕时间")
	}
	if len(roomRes.Coordinate) < 3 {
		resp.Code = 0
		resp.Msg = "至少设置3个范围坐标"
		comm.SendMsg(conn, 205, resp)
		return room, errors.New("至少设置3个范围坐标")
	}
	runPlayers := GetRunningPlayer(userId)
	if len(runPlayers) > 0 {
		//退出其他游戏房间
		for _, runPlayer := range runPlayers {
			_ = ExitRoom(runPlayer)
		}
	}
	// 获取随机数字
	var roomId string
	for {
		roomId = utils.RandNumString(4)
		db.Db.Table("fa_game_room").
			Where("id = ?", roomId).
			First(&room)
		log.Printf("创建room:%+v", room)
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
	chatRoomId, err := CreateChatRoom(conf.WebAdmin+"/app/other/createGroup", userId, "定位捉迷藏", "定位捉迷藏游戏")
	if err != nil {
		resp.Code = 0
		resp.Msg = err.Error()
		comm.SendMsg(conn, 205, resp)
		return room, err
	}
	room.ChatRoom = chatRoomId
	var player = Player{
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
	if err = tx.Table("fa_game_player").Create(&player).Error; err != nil {
		tx.Rollback()
		return room, err
	}
	tx.Commit()
	resp.Code = 1
	resp.Msg = "success"
	resp.Data = room
	comm.SendMsg(conn, 205, resp)
	log.Printf("【Game】房间创建成功:%+v", room)
	return room, nil
}

type CreateChatRoomResp struct {
	Code int
	Msg  string
	Data string
}

func GetRoomCache(roomId string) (Room, error) {
	key := "gameRoom:roomId_" + roomId
	var room Room
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
func CloseRoom(room Room, errorMsg string) {
	room.Status = 4
	room.CloseTime = time.Now().Unix()
	tx := db.Db.Begin()
	var err error
	if err = tx.Table("fa_game_room").Save(&room).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Table("fa_game_player").Where("room_id = ? AND status=?", room.Id, 0).
		Updates(Player{Status: 4, ErrorMsg: errorMsg}).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	ClearRoomCache(room.Id)
	ClearPlayersCache(room.Id)
	msg := comm.ResponseMsg{
		Code: 1,
		Msg:  "游戏关闭",
		Data: map[string]string{"roomId": room.Id},
	}
	SendMsgToPlayers(room.Id, 0, []Player{}, 209, msg)
}

// StartArrest 开始抓捕
func StartArrest(room Room) {
	db.Db.Table("fa_game_room").Where("id = ?", room.Id).Updates(Player{Status: 2})
	ClearRoomCache(room.Id)
}

// Referee 游戏结果裁决
func Referee(roomId string, room Room, winRole int, players []Player) {
	var err error
	if room.Id == "" {
		room, err = GetRoomCache(roomId)
		if err != nil {
			mylog.Error("获取游戏房间信息不正确,roomId:" + roomId)
			return
		}
	}
	var loseRole = 1
	if winRole == 1 {
		loseRole = 2
	}
	// 修改游戏状态
	room.Status = 3
	room.CloseTime = time.Now().Unix()
	tx := db.Db.Begin()
	if err = tx.Table("fa_game_room").Save(&room).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Table("fa_game_player").Where("room_id = ? AND status=? AND role = ?", room.Id, 0, winRole).
		Updates(Player{Status: 1}).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Table("fa_game_player").Where("room_id = ? AND status=? AND role =?", room.Id, 0, loseRole).
		Updates(Player{Status: 2}).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	ClearRoomCache(roomId)
	ClearPlayersCache(roomId)
	winnerData := make(map[string]interface{}) //获得胜利玩家信息
	winnerData["winRole"] = winRole
	//var winPlayers []Player
	var winPlayers []*Player
	if winRole == 1 {
		// 获胜狼的列表
		db.Db.Table("fa_game_player p").
			Select("p.user_id,p.room_id,p.role,p.status,count(*) cn").
			Joins("left join fa_game_vote v on v.room_id = p.room_id and v.user_id=p.user_id and v.status=1").
			Where("p.room_id = ? AND p.role = ? AND p.status = ?", roomId, 1, 1).
			Group("user_id").
			Order("cn desc").
			Limit(3).
			Find(winPlayers)
	} else {
		// 获胜羊的列表
		for _, player := range players {
			if player.Role == 2 {
				winPlayers = append(winPlayers, &player)
			}
		}
	}
	for _, player := range winPlayers {
		player.User = comm.GetUserById(player.UserId)
	}
	winnerData["winPlayers"] = winPlayers
	msg := comm.ResponseMsg{
		Code: 1,
		Msg:  "游戏结束",
		Data: winnerData,
	}
	SendMsgToPlayers(room.Id, 0, players, 208, msg)
	//msg = comm.ResponseMsg{
	//	Code: 1,
	//	Msg:  "游戏失败，请再接再厉！",
	//	Data: room,
	//}
	//SendMsgToPlayers(room.Id, loseRole, players, 208, msg)
	log.Printf("【Game】游戏结束,roomId:%s,winRole:%d", room.Id, winRole)
}

// EndRoom 游戏结束
// 游戏时间到，有猎物则猎物赢
// 全部为猎人-猎人赢/全部为猎物-猎物赢
func EndRoom(room Room) {
	players := GetRunningPlayersByRoomId(room.Id)
	isPreyWin := false //是否猎物赢
	for _, player := range players {
		if player.Role == 2 {
			isPreyWin = true
			break
		}
	}
	if isPreyWin {
		Referee(room.Id, room, 2, players)
	} else {
		Referee(room.Id, room, 1, players)
	}
}
