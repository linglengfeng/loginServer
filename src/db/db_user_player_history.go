package db

import (
	"context"
	"loginServer/src/db/db_mysql"
	"loginServer/src/db/db_redis"
)

// GetUserHistory 获取用户服务器列表（优先 Redis 缓存，未命中从 MySQL 加载并写入缓存）
func GetUserHistory(accountID string) (db_mysql.UserPlayerHistory, error) {
	ctx := context.Background()
	key := db_redis.GenUserHistoryKey(accountID)
	return db_redis.GetOrLoad(ctx, key, func() (db_mysql.UserPlayerHistory, error) {
		return db_mysql.GetUserHistory(accountID)
	}, db_redis.UserHistoryTTL)
}

// SetUserHistory 保存/更新玩家历史（先写 MySQL，再失效缓存）
func SetUserHistory(accountID string, newItem db_mysql.PlayerHistoryItem) error {
	if err := db_mysql.SetUserHistory(accountID, newItem); err != nil {
		return err
	}
	ctx := context.Background()
	_ = db_redis.Del(ctx, db_redis.GenUserHistoryKey(accountID))
	return nil
}

// SetUserState 更新用户账号状态（先写 MySQL，再失效缓存）
func SetUserState(accountID string, state int) error {
	if err := db_mysql.SetUserState(accountID, state); err != nil {
		return err
	}
	ctx := context.Background()
	_ = db_redis.Del(ctx, db_redis.GenUserHistoryKey(accountID))
	return nil
}
