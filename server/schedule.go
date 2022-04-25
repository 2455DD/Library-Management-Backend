package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	. "lms/services"
	"log"
	"net/url"
	"strconv"
	"time"
)

var scheduleDB *gorm.DB
var interval int

func updatePay() {
	for {
		_ = scheduleDB.Transaction(func(tx *gorm.DB) error {
			pays := make([]Pay, 0)
			tx.Where("done = ?", 0).Find(&pays)
			for _, pay := range pays {
				u := url.Values{"outtradeno": []string{strconv.Itoa(pay.Id)}}
				if ok, _ := agent.PayClient.VerifySign(u); ok {
					tx.Model(&pay).Select("done").Updates(&pay)
				}
			}
			return nil
		})
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func initSchedule(cfg *ini.File) {
	mysqlCfg, err := cfg.GetSection("mysql")
	if err != nil {
		log.Fatal("Fail to load section 'mysql': ", err)
	}
	username := mysqlCfg.Key("username").MustString("")
	password := mysqlCfg.Key("password").MustString("")
	address := mysqlCfg.Key("address").MustString("")
	tableName := mysqlCfg.Key("table").MustString("")
	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v?parseTime=true", username, password, address, tableName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("connect to DB failed: " + err.Error())
	}
	scheduleDB = db

	scheduleCfg, err := cfg.GetSection("schedule")
	if err != nil {
		log.Fatal("Fail to load section 'schedule': ", err)
	}
	interval = scheduleCfg.Key("interval").MustInt(5)
}

func startSchedule() {
	go updatePay()
}