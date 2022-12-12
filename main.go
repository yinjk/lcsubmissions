// @Desc
// @Author  inori
// @Update
package main

import (
	"github.com/BurntSushi/toml"
	"lcsubmissions/dao"
	"lcsubmissions/handler"
	"time"
)

type Config struct {
	Port     int
	Users    [][]string
	Dsn      string
	Interval time.Duration
}

var config Config

func init() {
	if _, err := toml.DecodeFile("conf/app.toml", &config); err != nil {
		panic(err)
	}
	if config.Port == 0 {
		config.Port = 8080
	}
}

func main() {
	if err := dao.InitMysql(config.Dsn); err != nil {
		panic(err)
	}
	handler.ListenAndStart(config.Port, config.Users, config.Interval)
}
