package main

import (
	"github.com/go-ini/ini"
	"github.com/smartwalle/alipay/v3"
	. "lms/services"
	"log"
)

var payAgent PayAgent

func initPayClient(cfg *ini.File) {
	ali, _ := cfg.GetSection("alipay")
	appId, _ := ali.GetKey("appId")
	private, _ := ali.GetKey("private")
	aliPublic, _ := ali.GetKey("aliPublic")

	client, err := alipay.New(appId.String(), private.String(), false)
	if err != nil {
		log.Fatal("alipay init err, ", err)
	}

	_ = client.LoadAliPayPublicKey(aliPublic.String())

	payAgent.PayClient = client
}
