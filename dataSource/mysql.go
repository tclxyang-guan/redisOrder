package dataSource

import (
	"baiwan/models"
	"baiwan/sysconfig"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"sync"
)

var mysqldb *gorm.DB

var once sync.Once

func initDb() {
	params := sysconfig.SysConfig.Mysql.UserName + ":" + sysconfig.SysConfig.Mysql.Password +
		"@(" + sysconfig.SysConfig.Mysql.Ip + ":" + sysconfig.SysConfig.Mysql.Port + ")/" + sysconfig.SysConfig.Mysql.DbName + "?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open("mysql", params)
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	db.AutoMigrate(
		models.Order{},
		models.Shop{},
		models.Store{},
	)
	db.DB().SetMaxOpenConns(10)
	db.DB().SetMaxIdleConns(20)
	// 启用Logger，显示详细日志
	db.LogMode(true)
	db.InstantSet("gorm:auto_preload", true)
	mysqldb = db
	log.Print("数据库初始化成功")
}

func GetDB() *gorm.DB {
	once.Do(initDb)
	return mysqldb
}
