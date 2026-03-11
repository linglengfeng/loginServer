package db

import (
	"fmt"
	"loginServer/src/db/db_mysql"
	"loginServer/src/db/db_redis"
)

// Start 初始化 MySQL 和 Redis 连接
func Start() error {
	if err := db_mysql.Start(); err != nil {
		return fmt.Errorf("mysql start failed: %w", err)
	}
	if err := db_redis.Start(); err != nil {
		return fmt.Errorf("redis start failed: %w", err)
	}
	return nil
}

// Close 关闭 MySQL 和 Redis 连接，应在程序退出时调用（不记录错误日志，避免日志系统已关闭）
func Close() {
	_ = db_mysql.Close()
	_ = db_redis.Close()
}
