package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"github.com/smartwalle/alipay/v3"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	. "lms/services"
	"lms/util"
	"log"
	"strconv"
	"time"
)

var scheduleDB *gorm.DB
var interval int
var reserveHours int

func updatePay() {
	userIdArr := make([]int, 0)
	_ = scheduleDB.Transaction(func(tx *gorm.DB) error {
		pays := make([]Pay, 0)
		tx.Where("done = ?", 0).Find(&pays)
		for _, pay := range pays {
			query := alipay.TradeQuery{
				OutTradeNo:   strconv.Itoa(pay.Id),
				QueryOptions: nil,
			}
			result, err := agent.PayClient.TradeQuery(query)
			if err != nil {
				continue
			}
			if result.Content.TradeStatus == alipay.TradeStatusSuccess {
				tx.Model(&pay).Select("done").Update("done", 1)
				userIdArr = append(userIdArr, pay.UserId)
			}
		}
		return nil
	})
	for _, userId := range userIdArr {
		agent.GetMemberFine(userId)
	}
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

				user := User{}
				tx.First(&user, reserve.UserId)
				book := Book{}
				tx.First(&book, reserve.BookId)
				content := fmt.Sprintf("The book 《%s》 you reserved at %s has been cancelled at %s", book.Name, util.TimeToString(util.StringToTime(reserve.StartTime)), reserve.EndTime)
				go sendEmail(user.Email, content)
			}
		}
		return nil
	})
}

func sendEmail(toEmail string, content string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "386401059@qq.com")
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Automatic reservation cancellation notice")
	m.SetBody("text/html", content)

	d := gomail.NewDialer("smtp.qq.com", 465, "386401059@qq.com", "fqiqwnwnjbvhbgbg")

	if err := d.DialAndSend(m); err != nil {
		log.Println("Send Email Failed, err ", err)
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