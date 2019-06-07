package server

import (
	"context"

	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/proto/cynosure"
)

func newHandler(c *common.Config) cynosure.APIServer {
	return &cynoHandler{
		c: c,
	}
}

type cynoHandler struct {
	c *common.Config
}

func (c *cynoHandler) Environment(context.Context, *cynosure.EnvironmentRequest) (*cynosure.EnvironmentResponse, error) {
	panic("implement me")
}

func (c *cynoHandler) Image(context.Context, *cynosure.ImageRequest) (*cynosure.ImageResponse, error) {
	panic("implement me")
}

func (c *cynoHandler) Info(context.Context, *cynosure.InfoRequest) (*cynosure.InfoResponse, error) {
	panic("implement me")
}

func (c *cynoHandler) Logs(context.Context, *cynosure.LogsRequest) (*cynosure.LogsResponse, error) {
	panic("implement me")
}

func (c *cynoHandler) Running(context.Context, *cynosure.RunningRequest) (*cynosure.RunningResponse, error) {
	panic("implement me")
}

func (c *cynoHandler) Start(context.Context, *cynosure.StartRequest) (*cynosure.StartResponse, error) {
	panic("implement me")
}

func (c *cynoHandler) Stop(context.Context, *cynosure.StopRequest) (*cynosure.StopResponse, error) {
	panic("implement me")
}
