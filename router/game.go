package router

import (
	"encoding/json"
	"websocket/impl"
	"websocket/lib/mylog"
	"websocket/model"
	"websocket/service"
)

type EnterRoom struct {
	service.BaseRouter
}

// Handle MsgId=105  进入房间
func (this *EnterRoom) Handle(request impl.IRequest) {
	uid := request.GetConnection().GetConnID()
	if uid == 0 {
		return
	}
	res := model.ReceiveMsg{}
	err := json.Unmarshal(request.GetData(), &res)
	if err != nil {
		mylog.Error("Unmarshal msg err:" + err.Error())
		return
	}
	var resp model.ResponseMsg
	resp.MsgId = 205
	resp.Code = 1
	resp.Msg = "success"
	var playerId string
	var ok bool
	if playerId, ok = res.Data["player_id"]; !ok {
		resp.Code = 0
		resp.Msg = "请求参数不正确"
		model.SendMsg(request, 205, resp)
		return
	}

	_, err = model.GetPlayerById(uid, playerId)
	if err != nil {
		resp.Code = 0
		resp.Msg = "不可加入当前房间"
		model.SendMsg(request, 205, resp)
		return
	}
	// TODO 给队房间用户发消息

	model.SendMsg(request, 205, resp)
}
