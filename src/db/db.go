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

// UpdateServerState 仅更新状态（维护/流畅等）
func UpdateServerState(clusterID int64, gameID int64, newState *int) error {
	return db_mysql.UpdateServerState(clusterID, gameID, newState)
}

// UpdateServerInfo 上报/注册服务器信息
func BatchUpdateServerInfo(serverReqs []*db_mysql.GameList) error {
	return db_mysql.BatchUpdateServerInfo(serverReqs)
}

// GetUserHistory 获取用户服务器列表
func GetUserHistory(accountID string) (*db_mysql.UserPlayerHistory, error) {
	return db_mysql.GetUserHistory(accountID)
}

// SetUserHistory 保存/更新玩家历史
func SetUserHistory(accountID string, newItem *db_mysql.PlayerHistoryItem) error {
	return db_mysql.SetUserHistory(accountID, *newItem)
}
