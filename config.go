package main

import (
	"github.com/astaxie/beego/config"

	"util/etcd"
	"util/logs"

	"core/net/lan"
)

var _ = logs.Debug

//
type Config struct {
	// player
	PlayerLsnAddr string // 监听客户端连接的地址
	PlayerMaxConn int    // 允许连接的最大客户端数
	PlayerMaxLoad int    // 每秒处理的消息上限，超过则踢出

	// server // to do
	LanCfg  lan.LanCfg
	EtcdCfg etcd.SrvCfg
}

//
var g_config = &Config{}

func Cfg() *Config {
	return g_config
}

//
func LoadConfig(confPath string) bool {
	// config
	confFile := confPath + "gate.ini"
	confd, e := config.NewConfig("ini", confFile)
	if e != nil {
		logs.Panicln("load config file failed! file:", confFile, "error:", e)
	}

	// player
	g_config.PlayerLsnAddr = confd.String("player::lsn_addr")
	g_config.PlayerMaxConn = confd.DefaultInt("player::max_conn", 100)
	g_config.PlayerMaxLoad = confd.DefaultInt("player::max_load", 100)

	// echo
	logs.Info("gate config:%+v", g_config)

	return true
}

// to do add check func
