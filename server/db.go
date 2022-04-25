package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	. "lms/services"
	"log"
)

var dbAgent DBAgent

func connectDB(cfg *ini.File) {
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
	dbAgent.DB = db
}
