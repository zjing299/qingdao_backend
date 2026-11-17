package bootstrap

import (
	"fmt"
	"github.com/zeromicro/go-zero/rest"
	"gogs.chinen.cc/chinen/common-boot-starter/boot"
	"qingdao_backend/internal/config"
	"qingdao_backend/internal/handler"
	"qingdao_backend/internal/svc"
)

type App struct {
	cfg *config.Config
	ctx *svc.ServiceContext
}

func NewApp(cfg *config.Config) boot.App { return &App{cfg: cfg} }
func (a *App) Init() error {
	ctx, err := svc.NewServiceContext(*a.cfg)
	if err != nil {
		return fmt.Errorf("failed to init service context: %w", err)
	}
	a.ctx = ctx
	return nil
}
func (a *App) RegisterHandlers(server *rest.Server) { handler.RegisterHandlers(server, a.ctx) }
func (a *App) Close() error {
	if a.ctx != nil {
		return a.ctx.Close()
	}
	return nil
}
