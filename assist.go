package main

import (
	"reflect"
	"strings"

	"util/logs"
)

//
var _ = logs.Debug

// protobuf message example
//message CSExample
//{
//    optional string gateway = 99[default="to=data|url=world|accId="];
//    required string account = 1;    // 账号
//}

//
type IProtoGateway interface {
	GetGateway() string
}

func parseGatewayByType(typ reflect.Type) []string {
	//
	v := reflect.Zero(typ)
	f, ok := v.Interface().(IProtoGateway)
	if !ok {
		return nil
	}

	gw := f.GetGateway()
	return parseGateway(gw)
}

// "set=userId|to=client|cache=data|del_cache=matchurl"
func parseGateway(tag string) []string {
	if "" == tag {
		return nil
	}

	var ret []string
	var findTo bool

	s := strings.Split(tag, "|")
	for _, v := range s {
		ss := strings.Split(v, "=")
		if len(ss) != 2 || ss[0] == "" {
			logs.Panicln("invalid tag:" + tag)
		}

		ret = append(ret, ss[0], ss[1])
		if GK_Proc_To == ss[0] {
			findTo = true
		}
	}

	if !findTo {
		return nil
	}

	return ret
}
