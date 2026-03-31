package service

import (
	"context"
)

type UserService struct{}

type GetUserRequest struct {
	UID int64 `json:"uid"`
}

type GetUserResponse struct {
	Code int    `json:"code"`
	UID  int64  `json:"uid"`
	Name string `json:"name"`
}

func (s *UserService) GetUser(ctx context.Context, req *GetUserRequest, res *GetUserResponse) error {
	res.Code = 0
	res.UID = req.UID
	res.Name = "Alice"
	return nil
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Code  int    `json:"code"`
	Token string `json:"token"`
	UID   int64  `json:"uid"`
}

func (s *UserService) Login(ctx context.Context, req *LoginRequest, res *LoginResponse) error {
	if req.Username == "admin" && req.Password == "123456" {
		res.Code = 0
		res.Token = "generated_token_123456"
		res.UID = 1
	} else {
		res.Code = 1
		res.Token = ""
		res.UID = 0
	}
	return nil
}
