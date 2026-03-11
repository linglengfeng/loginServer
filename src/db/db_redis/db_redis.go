package db_redis

import (
	"context"
	"encoding/json"
	"fmt"
	"loginServer/config"
	inredis "loginServer/pkg/redis"
	"loginServer/src/log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var DB *redis.Client

// Start 初始化 Redis 连接
func Start() error {
	redisip := config.Config.GetString("redis.ip")
	if redisip == "" {
		return fmt.Errorf("redis.ip is required")
	}
	redislink := inredis.Link{
		Password: config.Config.GetString("redis.password"),
		Db:       config.Config.GetInt("redis.db"),
		Ip:       redisip,
		Port:     config.Config.GetString("redis.port"),
	}
	DB = inredis.Start(redislink)
	if err := DB.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

// Close 关闭 Redis 连接，防止连接泄漏
func Close() error {
	if DB == nil {
		return nil
	}
	return DB.Close()
}

// 缓存键常量（Redis 惯例：冒号分层 + 业务前缀避免多项目冲突）
const (
	CacheKeyServerList      = "loginServer:server_list"   // 服务器列表
	CacheKeyServer          = "loginServer:server"        // 服务器单项前缀（GenServerKey 用）
	CacheKeyLoginNotice     = "loginServer:notice_login"  // 登录服公告
	CacheKeyIPWhitelistSCAN = "loginServer:ipwhitelist:"  // IP 白名单 key 前缀（GenIPWhitelistKey、SCAN 用）
	CacheKeyUserHistory     = "loginServer:user_history:" // 用户历史 key 前缀（GenUserHistoryKey 用）

	// UserHistoryTTL 用户历史缓存过期时间
	UserHistoryTTL = 3 * 24 * time.Hour
)

// ========== Key 生成 ==========

// GenServerKey 生成服务器列表单项缓存键（格式：loginServer:server:clusterID:gameID）
func GenServerKey(clusterID, gameID int64) string {
	return fmt.Sprintf("%s:%d:%d", CacheKeyServer, clusterID, gameID)
}

// GenIPWhitelistKey 生成 IP 白名单缓存键（格式：loginServer:ipwhitelist:apiGroup）
func GenIPWhitelistKey(apiGroup string) string {
	return CacheKeyIPWhitelistSCAN + strings.ToLower(apiGroup)
}

// GenUserHistoryKey 生成用户历史缓存键（格式：loginServer:user_history:accountID）
func GenUserHistoryKey(accountID string) string {
	return CacheKeyUserHistory + accountID
}

// ========== 通用 Redis 读写 ==========

// Get 从 Redis 读取 key，返回字节；key 不存在返回 redis.Nil
func Get(ctx context.Context, key string) ([]byte, error) {
	if DB == nil {
		return nil, redis.Nil
	}
	return DB.Get(ctx, key).Bytes()
}

// Set 写入 Redis，value 为字节。ttl 为可选过期时间，不传或传 0 表示永不过期
func Set(ctx context.Context, key string, value []byte, ttl ...time.Duration) error {
	if DB == nil {
		return nil
	}
	var d time.Duration
	if len(ttl) > 0 {
		d = ttl[0]
	}
	return DB.Set(ctx, key, value, d).Err()
}

// Del 删除指定 key，用于缓存失效
func Del(ctx context.Context, key string) error {
	if DB == nil {
		return nil
	}
	return DB.Del(ctx, key).Err()
}

// GetOrLoad 从 Redis 获取数据，未命中时通过 loader 加载并写回 Redis（JSON 序列化）
// ttl 为可选过期时间，不传或传 0 表示永不过期
func GetOrLoad[T any](ctx context.Context, key string, loader func() (T, error), ttl ...time.Duration) (T, error) {
	var zero T
	if DB == nil {
		return loader()
	}

	data, err := DB.Get(ctx, key).Bytes()
	if err == nil {
		var v T
		unmarshalErr := json.Unmarshal(data, &v)
		if unmarshalErr == nil {
			return v, nil
		}
		log.Warn("redis unmarshal key=%s failed: %v", key, unmarshalErr)
	}
	if err != nil && err != redis.Nil {
		log.Warn("redis get key=%s failed: %v", key, err)
	}

	v, loadErr := loader()
	if loadErr != nil {
		return zero, loadErr
	}

	var d time.Duration
	if len(ttl) > 0 {
		d = ttl[0]
	}
	if b, marshalErr := json.Marshal(v); marshalErr == nil {
		if setErr := DB.Set(ctx, key, b, d).Err(); setErr != nil {
			log.Warn("redis set key=%s failed: %v", key, setErr)
		}
	}
	return v, nil
}
