package db_mysql

import (
	"fmt"
	"loginServer/config"
	"loginServer/pkg/mysql"

	"gorm.io/gorm"
)

var DB *gorm.DB

// Start 初始化 MySQL 连接
func Start() error {
	mysqlip := config.Config.GetString("mysql.ip")
	if mysqlip == "" {
		return fmt.Errorf("mysql.ip is required")
	}
	mysqlink := mysql.Link{
		User:     config.Config.GetString("mysql.user"),
		Password: config.Config.GetString("mysql.password"),
		Ip:       mysqlip,
		Port:     config.Config.GetString("mysql.port"),
		Db:       config.Config.GetString("mysql.db"),
	}
	db, err := mysql.Start(mysqlink)
	if err != nil {
		return err
	}
	DB = db
	return nil
}

// Close 关闭 MySQL 连接，防止连接泄漏
func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
