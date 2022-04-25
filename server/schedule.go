package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	. "lms/services"
	"lms/util"
	"log"
	"net/url"
	"strconv"
	"time"
)

var scheduleDB *gorm.DB
var interval int
var reserveHours int

func updatePay() {
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
}

func updateReserve() {
	now := time.Now()
	_ = scheduleDB.Transaction(func(tx *gorm.DB) error {
		reserves := make([]ReserveBook, 0)
		tx.Where("end_time is null").Find(&reserves)
		for _, reserve := range reserves {
			startTime := util.StringToTime(reserve.StartTime)
			if int(now.Sub(startTime).Seconds()) > reserveHours * 3600 {
				reserve.EndTime = util.TimeToString(now)
				tx.Model(&reserve).Select("end_time").Updates(&reserve)
			}
		}
		return nil
	})
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
	reserveHours = scheduleCfg.Key("reserveHours").MustInt(4)
}

func startSchedule() {
	go func() {
		for {
			updatePay()
			updateReserve()
			time.Sleep(time.Second * time.Duration(interval))
		}
	}()
}