package config

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	bootConfig "gogs.chinen.cc/chinen/common-boot-starter/config"
)

type Config struct {
	rest.RestConf
	LogConf   logx.LogConf           `json:",optional"`
	Service   bootConfig.ServiceConf `json:"Service,optional"`
	Timestamp TimestampConf          `json:",optional"`
}

// TimestampConf reserves the upstream endpoints; credentials stay on the server.
type TimestampConf struct {
	BaseURL    string `json:",optional"`
	TokenPath  string `json:",default=/token"`
	RSAPath    string `json:",default=/rsa"`
	SMPath     string `json:",default=/sm"`
	VerifyPath string `json:",default=/verify"`
	AppKey     string `json:",optional"`
	AppSecret  string `json:",optional"`
}

func (c *Config) GetLogConf() logx.LogConf                { return c.LogConf }
func (c *Config) GetPort() int                            { return c.Port }
func (c *Config) GetRestConf() rest.RestConf              { return c.RestConf }
func (c *Config) GetServiceConf() *bootConfig.ServiceConf { return &c.Service }
