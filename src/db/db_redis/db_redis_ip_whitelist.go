package db_redis

import (
	"context"
	"encoding/json"

	"loginServer/src/log"

	"github.com/redis/go-redis/v9"
)

// GetIPWhitelist 从 Redis 读取指定 key 的 IP 白名单（JSON []string），不存在返回 nil, nil
func GetIPWhitelist(ctx context.Context, key string) ([]string, error) {
	if DB == nil {
		return nil, nil
	}
	data, err := DB.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		log.Warn("GetIPWhitelist redis get key=%s failed: %v", key, err)
		return nil, err
	}
	var ips []string
	if unmarshalErr := json.Unmarshal(data, &ips); unmarshalErr != nil {
		log.Warn("GetIPWhitelist unmarshal key=%s failed: %v", key, unmarshalErr)
		return nil, unmarshalErr
	}
	return ips, nil
}

// SetIPWhitelist 将 IP 白名单写入 Redis（JSON）
func SetIPWhitelist(ctx context.Context, key string, ips []string) error {
	if DB == nil {
		return nil
	}
	b, err := json.Marshal(ips)
	if err != nil {
		log.Warn("SetIPWhitelist marshal key=%s failed: %v", key, err)
		return err
	}
	if setErr := DB.Set(ctx, key, b, 0).Err(); setErr != nil {
		log.Warn("SetIPWhitelist redis set key=%s failed: %v", key, setErr)
		return setErr
	}
	return nil
}

// ScanIPWhitelistKeys 扫描 prefix 开头的所有 key，读取并反序列化为 []string，返回 map[key][]string
func ScanIPWhitelistKeys(ctx context.Context, prefix string) map[string][]string {
	result := make(map[string][]string)
	if DB == nil {
		return result
	}
	var cursor uint64
	for {
		keys, nextCursor, err := DB.Scan(ctx, cursor, prefix+"*", 0).Result()
		if err != nil {
			log.Warn("ScanIPWhitelistKeys redis scan failed: %v", err)
			break
		}
		for _, key := range keys {
			ips, err := GetIPWhitelist(ctx, key)
			if err != nil || ips == nil {
				continue
			}
			result[key] = ips
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return result
}
