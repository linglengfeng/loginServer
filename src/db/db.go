package db

import (
	"loginServer/src/db/db_mysql"
	"loginServer/src/db/db_redis"
)

func Start() {
	db_mysql.Start()
	db_redis.Start()
}

// GetServerList 获取服务器列表
func GetServerList() ([]db_mysql.GameList, error) {
	return db_mysql.GetServerList()
}
