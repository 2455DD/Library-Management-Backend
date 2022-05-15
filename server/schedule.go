package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	. "lms/services"
	"log"
	"time"
)

func updatePay() {
	agent.UpdatePay()
}

func updateReserve() {
	agent.UpdateReserve()
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
	ScheduleDB = db

	scheduleCfg, err := cfg.GetSection("schedule")
	if err != nil {
		log.Fatal("Fail to load section 'schedule': ", err)
	}
	Interval = scheduleCfg.Key("interval").MustInt(5)
	ReserveHours = scheduleCfg.Key("reserveHours").MustInt(4)
}

func startSchedule() {
	go func() {
		for {
			updatePay()
			updateReserve()
			time.Sleep(time.Second * time.Duration(Interval))
		}
	}()
}
