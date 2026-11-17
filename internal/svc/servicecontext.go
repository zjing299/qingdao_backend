package svc

import (
	"qingdao_backend/internal/client"
	"qingdao_backend/internal/config"
	"qingdao_backend/internal/logic/timestamp"
)

type ServiceContext struct {
	Config    config.Config
	Timestamp *timestamp.Service
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	return &ServiceContext{Config: c, Timestamp: timestamp.New(client.NewTimestampClient(c.Upstream))}, nil
}
func (s *ServiceContext) Close() error { return nil }
