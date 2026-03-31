package handler

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/codes"
	"github.com/dobyte/due/v2/log"
)

const RouteChat = 2

type ChatRequest struct {
	Content string `json:"content"`
}

type ChatResponse struct {
	Code int `json:"code"`
}

func ChatHandler(ctx node.Context) {
	req := &ChatRequest{}
	res := &ChatResponse{}

	defer func() {
		ctx.Response(res)
	}()

	if err := ctx.Parse(req); err != nil {
		res.Code = codes.InternalError.Code()
		return
	}

	log.Infof("玩家聊天：uid=%d, content=%s", ctx.Uid(), req.Content)
	res.Code = codes.OK.Code()
}
