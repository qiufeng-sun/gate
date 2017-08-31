package main

import (
	"path/filepath"

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
	LanCfg  *lan.LanCfg
	EtcdCfg etcd.SrvCfg
}

func (this *Config) init(fileName string) bool {
	confd, e := config.NewConfig("ini", fileName)
	if e != nil {
		logs.Panicln("load config file failed! file:", fileName, "error:", e)
	}

	//[scribe]
	//open=false
	//addr=localhost:7915

	// player
	this.PlayerLsnAddr = confd.String("player::lsn_addr")
	this.PlayerMaxConn = confd.DefaultInt("player::max_conn", 100)
	this.PlayerMaxLoad = confd.DefaultInt("player::max_load", 100)

	//[server]
	srvName := confd.String("server::name")
	srvAddr := confd.String("server::addr")
	this.LanCfg = lan.NewLanCfg(srvName, srvAddr)

	//[etcd]
	this.EtcdCfg.EtcdAddrs = confd.Strings("etcd::addrs")
	this.EtcdCfg.SrvAddr = srvAddr
	this.EtcdCfg.SrvRegPath = confd.String("etcd::reg_path")
	this.EtcdCfg.SrvRegUpTick = confd.DefaultInt64("etcd::reg_uptick", 2000)

	this.EtcdCfg.WatchPaths = confd.Strings("etcd::watch_path")

	//#close client notify
	//close_notify_must=match;data
	//close_notify_cached=battle

	// echo
	logs.Info("gate config:%+v", *this)

	return true
}

//
var g_config = &Config{}

func Cfg() *Config {
	return g_config
}

//
func LoadConfig(confPath string) bool {
	// config
	confFile := filepath.Clean(confPath + "/gate.ini")

	return g_config.init(confFile)
}

// to do add check func
