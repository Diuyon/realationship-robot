package main

import (
	"context"
	"relationship-robot/business"
	"relationship-robot/client"
	"relationship-robot/util"
	"time"
)

// Token Type 为 Bot
const (
	APPID    = 101999642
	APPTOKEN = "Vgpy8Qlkq8G54ADJESgb2YvLFTAZspVM"
)

var token string

func init() {
	token = util.GetToken(APPID, APPTOKEN)
}

func main() {
	cfg := client.Cfg{
		ApiCfg: client.ApiClientCfg{
			Sandbox: false,
			Timeout: 3 * time.Second,
		},
	}
	c := client.New(context.Background(), token, cfg)
	err := c.Start(context.Background(), business.ReturnAtMessage)
	if err != nil {
		panic(err)
	}

	select {} // TODO 完善 start
}
