package sysconfig

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type sysStruct struct {
	Port     string
	Mysql    mysql
	Redis    redis
	RabbitMQ rabbitMQ
}
type mysql struct {
	UserName string
	Password string
	Ip       string
	Port     string
	DbName   string
}
type redis struct {
	Addr     string
	Password string
}
type rabbitMQ struct {
	UserName string
	Password string
	Addr     string
}

var SysConfig = &sysStruct{}

func init() {
	c, e := ioutil.ReadFile("./config.json")
	if e != nil {
		log.Fatal("read config err")
	}
	e = json.Unmarshal(c, SysConfig)
	if e != nil {
		log.Fatal("config json err")
	}
}
