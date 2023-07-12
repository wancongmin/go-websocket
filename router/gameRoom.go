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

// Handle 创建房间 MsgId=105
func (this *CreateRoom) Handle(request impl.IRequest) {
	userId := request.GetConnection().GetConnID()
	room, err := game.CreateRoom(request, userId)
	if err != nil {
		log.Printf("房间创建失败,err:" + err.Error())
		return
	}
	// 启动服务
	roomChecker := game.NewGameRoomChecker(3*time.Second, room)
	roomChecker.CloseRoomHandle = game.CloseRoom
	roomChecker.EndRoomHandle = game.EndRoom
	roomChecker.Start()
	log.Printf("房间创建成功,roomId:" + room.Id)
}
