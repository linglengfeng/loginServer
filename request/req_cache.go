package request

import (
	"loginServer/src/db"
	"loginServer/src/db/db_mysql"
	"loginServer/src/log"
	"sync"
)

const (
	CacheKeyServerList = "serverList" // 服务器列表缓存键
)

// cacheItem 缓存项
type cacheItem struct {
	data any
}

// cache 缓存结构（线程安全）
type cache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

var (
	// cacheInstance 全局缓存实例
	cacheInstance = &cache{
		items: make(map[string]*cacheItem),
	}
)

// Get 从缓存获取数据
func (c *cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	return item.data, true
}

// Set 设置缓存数据
func (c *cache) Set(key string, data any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &cacheItem{data: data}
}

// GetWithLoader 从缓存获取数据，如果缓存没有则使用加载器从数据源获取并设置到缓存
// 使用 double-check 模式避免竞态条件
func (c *cache) GetWithLoader(key string, loader func() (any, error)) (any, error) {
	// 第一次检查（读锁）
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if exists {
		return item.data, nil
	}

	// 缓存不存在，获取写锁
	c.mu.Lock()
	// 第二次检查（double-check），防止在获取写锁期间其他 goroutine 已经设置了缓存
	item, exists = c.items[key]
	if exists {
		c.mu.Unlock()
		return item.data, nil
	}

	// 使用加载器从数据源获取
	data, err := loader()
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}

	// 设置到缓存
	c.items[key] = &cacheItem{data: data}
	c.mu.Unlock()

	return data, nil
}

// GetFromCache 从缓存获取数据
func GetFromCache(key string) (any, bool) {
	return cacheInstance.Get(key)
}

// SetToCache 设置数据到缓存
func SetToCache(key string, data any) {
	cacheInstance.Set(key, data)
}

// GetFromCacheWithLoader 从缓存获取数据，如果缓存没有则使用加载器
func GetFromCacheWithLoader(key string, loader func() (any, error)) (any, error) {
	return cacheInstance.GetWithLoader(key, loader)
}

// ========== 服务器列表缓存封装 ==========

// loadServerList 加载服务器列表的数据加载器
func loadServerList() (any, error) {
	servers, err := db.GetServerList()
	if err != nil {
		log.Error("loadServerList from database failed, err:%v", err)
		return nil, err
	}
	return servers, nil
}

// GetServerList 获取服务器列表（优先从缓存，缓存没有则从数据库）
func GetServerList() ([]db_mysql.GameList, error) {
	data, err := GetFromCacheWithLoader(CacheKeyServerList, loadServerList)
	if err != nil {
		return nil, err
	}

	servers, ok := data.([]db_mysql.GameList)
	if !ok {
		log.Error("GetServerList cache data type error")
		return []db_mysql.GameList{}, nil
	}

	return servers, nil
}

// SetServerList 设置服务器列表到缓存
func SetServerList(data []db_mysql.GameList) {
	SetToCache(CacheKeyServerList, data)
}
