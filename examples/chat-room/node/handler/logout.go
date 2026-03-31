package handler

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/codes"
	"github.com/dobyte/due/v2/log"
)

const RouteLogout = 3

type LogoutRequest struct{}

type LogoutResponse struct {
	Code int `json:"code"`
}

func LogoutHandler(ctx node.Context) {
	res := &LogoutResponse{}
	defer func() {
		ctx.Response(res)
	}()

	log.Infof("玩家登出：uid=%d", ctx.Uid())
	res.Code = codes.OK.Code()
}
