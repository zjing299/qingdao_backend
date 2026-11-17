package main

import (
	_ "embed"
	"os"

	"github.com/zeromicro/go-zero/core/logx"
	"gogs.chinen.cc/chinen/common-boot-starter/boot"
	"qingdao_backend/internal/bootstrap"
	"qingdao_backend/internal/config"
)

//go:embed etc/config.yaml
var configBytes []byte

func main() {
	var c config.Config
	if err := boot.Run(configBytes, bootstrap.NewApp(&c), &c); err != nil {
		logx.Errorf("bootstrap failed: %v", err)
		os.Exit(1)
	}
}
