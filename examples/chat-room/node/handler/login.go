package handler

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/codes"
	"github.com/dobyte/due/v2/log"
)

const RouteLogin = 1

type LoginRequest struct {
	UID  int64  `json:"uid"`
	Name string `json:"name"`
}

type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	RoomID  string `json:"room_id"`
}

func LoginHandler(ctx node.Context) {
	req := &LoginRequest{}
	res := &LoginResponse{}

	defer func() {
		ctx.Response(res)
	}()

	if err := ctx.Parse(req); err != nil {
		res.Code = codes.InternalError.Code()
		return
	}

	res.Code = codes.OK.Code()
	res.Message = "登录成功"
	res.RoomID = "room_001"

	log.Infof("玩家登录：uid=%d, name=%s", req.UID, req.Name)
}
