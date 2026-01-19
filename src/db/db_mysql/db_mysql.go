package db_mysql

import (
	"loginServer/config"
	"loginServer/pkg/mysql"

	"gorm.io/gorm"
)

var DB *gorm.DB

func Start() {
	mysqlip := config.Config.GetString("mysql.ip")
	if mysqlip != "" {
		mysqlink := mysql.Link{
			User:     config.Config.GetString("mysql.user"),
			Password: config.Config.GetString("mysql.password"),
			Ip:       mysqlip,
			Port:     config.Config.GetString("mysql.port"),
			Db:       config.Config.GetString("mysql.db"),
		}
		DB = mysql.Start(mysqlink)
	}
}

// GameList 游戏服务器列表结构体
type GameList struct {
	ID        int64   `gorm:"column:id;primaryKey" json:"id"`
	ClusterID *int64  `gorm:"column:cluster_id" json:"cluster_id"`
	GameID    *int64  `gorm:"column:game_id" json:"game_id"`
	Name      *string `gorm:"column:name" json:"name"`
	IsShow    *int    `gorm:"column:is_show" json:"is_show"`
	State     *int64  `gorm:"column:state" json:"state"`
	IsNew     *int    `gorm:"column:is_new" json:"is_new"`
	LoginURL  *string `gorm:"column:login_url" json:"login_url"`
	Desc      *string `gorm:"column:desc" json:"desc"`
	Info      *string `gorm:"column:info" json:"info"`
}

// GetServerList 获取服务器列表（返回所有服务器）
func GetServerList() ([]GameList, error) {
	var servers []GameList
	err := DB.Find(&servers).Error
	return servers, err
}
