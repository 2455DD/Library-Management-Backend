package main

import (
	"flag"
	"github.com/go-ini/ini"
	. "lms/services"
	"log"
	"os"
	"path/filepath"
)

func loadConfig(configPath string) {
	cfg, err := ini.Load(configPath)
	if err != nil {
		log.Fatal("Fail to Load config: ", err)
	}

	initServer(cfg)
	initPayClient(cfg)
	connectDB(cfg)
	initSchedule(cfg)

	MediaPath = filepath.Join(path, "media")
	err = os.MkdirAll(MediaPath, os.ModePerm)
	if err != nil {
		log.Fatal("file system failed to create path: " + err.Error())
	}
}

func main() {
	var configPath = flag.String("config", "./app.ini", "配置文件路径")
	flag.Parse()
	loadConfig(*configPath)
	startSchedule()
	startService()
}
