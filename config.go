package main

import (
	"github.com/astaxie/beego/config"

	"util/logs"
)

var _ = logs.Debug

//
type Config struct {
	PlayerLsnAddr string
	PlayerMaxConn int
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

	// echo
	logs.Info("gate config:%+v", g_config)

	return true
}

// to do add check func
