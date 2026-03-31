package handler

import (
	"fmt"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/codes"
	"github.com/dobyte/due/v2/log"
	"github.com/dobyte/due/v2/utils/xtime"
)

const GreetRoute = 1

type GreetRequest struct {
	Message string `json:"message"`
}

type GreetResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func GreetHandler(ctx node.Context) {
	req := &GreetRequest{}
	res := &GreetResponse{}

	defer func() {
		if err := ctx.Response(res); err != nil {
			log.Errorf("response message failed: %v", err)
		}
	}()

	if err := ctx.Parse(req); err != nil {
		log.Errorf("parse request message failed: %v", err)
		res.Code = codes.InternalError.Code()
		return
	}

	log.Info(req.Message)
	res.Code = codes.OK.Code()
	res.Message = fmt.Sprintf("I'm server, and the current time is: %s",
		xtime.Now().Format(xtime.DateTime))
}
