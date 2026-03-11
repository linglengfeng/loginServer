package db

import (
	"context"
	"loginServer/src/db/db_mysql"
	"loginServer/src/db/db_redis"
)

// IPWhitelistCacheKeyPrefix IP 白名单 Redis key 前缀，用于从完整 key 解析分组名
const IPWhitelistCacheKeyPrefix = db_redis.CacheKeyIPWhitelistSCAN

// GetWhitelistFromCache 从 Redis 缓存获取指定分组的白名单（未命中则从 MySQL 加载并写入缓存）
func GetWhitelistFromCache(apiGroup string) []string {
	ctx := context.Background()
	key := db_redis.GenIPWhitelistKey(apiGroup)
	ips, err := db_redis.GetOrLoad(ctx, key, func() ([]string, error) {
		list, err := db_mysql.LoadWhitelist(apiGroup)
		if err != nil {
			return nil, err
		}
		result := make([]string, 0, len(list))
		for _, w := range list {
			result = append(result, w.IP)
		}
		return result, nil
	})
	if err != nil || ips == nil {
		return nil
	}
	result := make([]string, len(ips))
	copy(result, ips)
	return result
}

// SetWhitelistToCache 设置指定分组的白名单到 Redis 缓存
func SetWhitelistToCache(apiGroup string, ips []string) {
	ctx := context.Background()
	key := db_redis.GenIPWhitelistKey(apiGroup)
	_ = db_redis.SetIPWhitelist(ctx, key, ips)
}

// GetAllWhitelistItems 获取所有白名单缓存项（缓存为空时从 MySQL 加载并写入缓存）
func GetAllWhitelistItems() map[string][]string {
	ctx := context.Background()
	items := db_redis.ScanIPWhitelistKeys(ctx, db_redis.CacheKeyIPWhitelistSCAN)
	if len(items) == 0 {
		// 缓存为空，从 MySQL 加载并回填缓存
		list, err := db_mysql.LoadAllWhitelists()
		if err != nil {
			return nil
		}
		groupMap := make(map[string][]string)
		for _, w := range list {
			group := db_redis.GenIPWhitelistKey(w.APIGroup)
			groupMap[group] = append(groupMap[group], w.IP)
		}
		for group, ips := range groupMap {
			_ = db_redis.SetIPWhitelist(ctx, group, ips)
			items[group] = ips
		}
	}
	return items
}

// LoadWhitelist 从数据库加载指定分组的白名单
func LoadWhitelist(apiGroup string) ([]db_mysql.IPWhitelist, error) {
	return db_mysql.LoadWhitelist(apiGroup)
}

// LoadAllWhitelists 从数据库加载所有分组的白名单
func LoadAllWhitelists() ([]db_mysql.IPWhitelist, error) {
	return db_mysql.LoadAllWhitelists()
}

// SetWhitelist 设置指定分组的IP白名单（完全替换）
func SetWhitelist(apiGroup string, ips []string) error {
	return db_mysql.SetWhitelist(apiGroup, ips)
}

// AddWhitelistIP 向指定分组添加IP
func AddWhitelistIP(apiGroup string, ip string) error {
	return db_mysql.AddWhitelistIP(apiGroup, ip)
}

// RemoveWhitelistIP 从指定分组删除IP
func RemoveWhitelistIP(apiGroup string, ip string) error {
	return db_mysql.RemoveWhitelistIP(apiGroup, ip)
}

// RemoveWhitelistGroup 删除整个分组的所有IP
func RemoveWhitelistGroup(apiGroup string) error {
	return db_mysql.RemoveWhitelistGroup(apiGroup)
}
