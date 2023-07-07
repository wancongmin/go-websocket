package router

import (
	"log"
	"time"
	"websocket/impl"
	"websocket/model/game"
	"websocket/service"
)

type CreateRoom struct {
	service.BaseRouter
}

// 创建房间 MsgId=105
func (this *CreateRoom) Handle(request impl.IRequest) {
	userId := request.GetConnection().GetConnID()
	room, err := game.CreateRoom(request, userId)
	if err != nil {
		log.Printf("房间创建失败,err:" + err.Error())
		return
	}
	log.Printf("房间创建成功,roomId:" + room.Id)
	// 启动服务
	roomChecker := game.NewGameRoomChecker(3*time.Second, room)
	roomChecker.CloseRoomHandle = game.CloseRoom
	roomChecker.EndRoomHandle = game.EndRoom
	roomChecker.Start()
}

type EnterRoom struct {
	service.BaseRouter
}

// Handle MsgId=105  进入房间
//func (this *EnterRoom) Handle(request impl.IRequest) {
//	uid := request.GetConnection().GetConnID()
//	if uid == 0 {
//		return
//	}
//	res := comm.ReceiveMsg{}
//	err := json.Unmarshal(request.GetData(), &res)
//	if err != nil {
//		mylog.Error("Unmarshal msg err:" + err.Error())
//		return
//	}
//	var resp comm.ResponseMsg
//	resp.MsgId = 205
//	resp.Code = 1
//	resp.Msg = "success"
//	var playerId string
//	var ok bool
//	conn := request.GetConnection()
//	if playerId, ok = res.Data["player_id"]; !ok {
//		resp.Code = 0
//		resp.Msg = "请求参数不正确"
//		comm.SendMsg(conn, 205, resp)
//		return
//	}
//	_, err = game.GetPlayerById(uid, playerId)
//	if err != nil {
//		resp.Code = 0
//		resp.Msg = "不可加入当前房间"
//		comm.SendMsg(conn, 205, resp)
//		return
//	}
//	// TODO 给队房间用户发消息
//
//	comm.SendMsg(conn, 205, resp)
//}
