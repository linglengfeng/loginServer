package mysql

import (
	"fmt"
	stdlog "log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Link struct {
	User     string
	Password string
	Ip       string
	Port     string
	Db       string
}

func Start(mysqllink Link) (*gorm.DB, error) {
	user := mysqllink.User
	password := mysqllink.Password
	ip := mysqllink.Ip
	port := mysqllink.Port
	db := mysqllink.Db
	dsn := user + ":" + password + "@tcp(" + ip + ":" + port + ")/" + db + "?charset=utf8mb4&parseTime=True&loc=Local"
	gormLogger := logger.New(
		stdlog.New(os.Stdout, "", stdlog.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
	mysqlconfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: gormLogger,
	}
	dbcli, err := gorm.Open(mysql.Open(dsn), mysqlconfig)

	if err != nil {
		// 避免把包含密码的 DSN 输出到日志
		return nil, fmt.Errorf("无法连接到数据库：%w", err)
	}
	return dbcli, nil
}
