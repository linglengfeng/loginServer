package request

import (
	"fmt"
	"loginServer/src/db"
	"loginServer/src/db/db_mysql"
	"loginServer/src/log"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	CacheKeyServerList  = "serverList"  // 服务器列表缓存键
	CacheKeyServer      = "server"      // 服务器缓存键
	CacheKeyLoginNotice = "loginNotice" // 登录服公告缓存键
	CacheKeyWhitelist   = "whitelist"   // IP白名单缓存键前缀
)

var (
	globalCacheInstance = cache.New(cache.NoExpiration, 0)
)

// ========== 基础函数 ==========

// GetFromCacheWithLoader 从缓存获取数据，缓存未命中时使用加载器
func GetFromCacheWithLoader(key string, loader func() (any, error)) (any, error) {
	data, exists := globalCacheInstance.Get(key)
	if exists {
		return data, nil
	}

	data, err := loader()
	if err != nil {
		return nil, err
	}

	globalCacheInstance.Set(key, data, cache.NoExpiration)
	return data, nil
}

// ========== 业务函数 ==========

// ========== 服务器列表缓存 ==========

// GetServerList 获取服务器列表
func GetServerList() ([]db_mysql.GameList, error) {
	data, err := GetFromCacheWithLoader(CacheKeyServerList, loadServerList)
	if err != nil {
		log.Error("GetServerList: failed to get server list from cache: %v", err)
		return nil, err
	}

	gameList, ok := data.([]db_mysql.GameList)
	if !ok {
		log.Error("GetServerList: invalid cache data type, expected []db_mysql.GameList, got %T", data)
		return nil, fmt.Errorf("invalid cache data format")
	}

	return gameList, nil
}

// UpdateCacheServerList 更新服务器列表缓存
func UpdateCacheServerList(updates []db_mysql.GameList) {
	var serverKeyList []db_mysql.GameList

	if listData, exists := globalCacheInstance.Get(CacheKeyServerList); exists {
		if list, ok := listData.([]db_mysql.GameList); ok {
			serverKeyList = make([]db_mysql.GameList, len(list))
			copy(serverKeyList, list)
		}
	}

	if serverKeyList == nil {
		serverKeyList = make([]db_mysql.GameList, 0)
	}

	listChanged := false
	serverMap := make(map[string]int)

	for i, server := range serverKeyList {
		key := genServerKey(server.ClusterID, server.GameID)
		serverMap[key] = i
	}

	for _, update := range updates {
		key := genServerKey(update.ClusterID, update.GameID)
		serverData, exists := globalCacheInstance.Get(key)

		if exists {
			serverInfo, ok := serverData.(db_mysql.GameList)
			if ok {
				updatedServer := serverInfo

				if update.Name != nil {
					if updatedServer.Name == nil {
						nameCopy := *update.Name
						updatedServer.Name = &nameCopy
					} else {
						*updatedServer.Name = *update.Name
					}
				}
				if update.Addr != nil {
					if updatedServer.Addr == nil {
						addrCopy := *update.Addr
						updatedServer.Addr = &addrCopy
					} else {
						*updatedServer.Addr = *update.Addr
					}
				}
				if update.Info != nil {
					if updatedServer.Info == nil {
						infoCopy := *update.Info
						updatedServer.Info = &infoCopy
					} else {
						*updatedServer.Info = *update.Info
					}
				}
				if update.Port != nil {
					if updatedServer.Port == nil {
						portCopy := *update.Port
						updatedServer.Port = &portCopy
					} else {
						*updatedServer.Port = *update.Port
					}
				}
				if update.State != nil {
					if updatedServer.State == nil {
						stateCopy := *update.State
						updatedServer.State = &stateCopy
					} else {
						*updatedServer.State = *update.State
					}
				}
				if update.IsShow != nil {
					if updatedServer.IsShow == nil {
						isShowCopy := *update.IsShow
						updatedServer.IsShow = &isShowCopy
					} else {
						*updatedServer.IsShow = *update.IsShow
					}
				}
				if update.IsNew != nil {
					if updatedServer.IsNew == nil {
						isNewCopy := *update.IsNew
						updatedServer.IsNew = &isNewCopy
					} else {
						*updatedServer.IsNew = *update.IsNew
					}
				}
				if update.Desc != nil {
					if updatedServer.Desc == nil {
						descCopy := *update.Desc
						updatedServer.Desc = &descCopy
					} else {
						*updatedServer.Desc = *update.Desc
					}
				}

				globalCacheInstance.Set(key, updatedServer, cache.NoExpiration)

				if idx, found := serverMap[key]; found {
					serverKeyList[idx] = updatedServer
				} else {
					serverMap[key] = len(serverKeyList)
					serverKeyList = append(serverKeyList, updatedServer)
				}
				listChanged = true
			}
		} else {
			globalCacheInstance.Set(key, update, cache.NoExpiration)
			serverKeyList = append(serverKeyList, update)
			listChanged = true
		}
	}

	if listChanged {
		globalCacheInstance.Set(CacheKeyServerList, serverKeyList, cache.NoExpiration)
	}
}

// ========== 游戏公告缓存 ==========

// GetLoginNotice 获取公告
func GetLoginNotice() ([]db_mysql.LoginNotice, error) {
	data, err := GetFromCacheWithLoader(CacheKeyLoginNotice, loadLoginNotice)
	if err != nil {
		return nil, err
	}

	noticeCache, ok := data.([]db_mysql.LoginNotice)
	if !ok {
		log.Error("GetLoginNoticeList cache data type error")
		return []db_mysql.LoginNotice{}, nil
	}

	typeMap := make(map[int]db_mysql.LoginNotice)
	now := time.Now().Unix()
	for _, n := range noticeCache {
		if n.StartTime <= now && n.EndTime > now {
			notice, exists := typeMap[n.NoticeType]
			if !exists {
				typeMap[n.NoticeType] = n
			} else {
				// 选取优先级更高的；优先级相同则选 ID 更大的（更“新”）
				if notice.Priority < n.Priority {
					typeMap[n.NoticeType] = n
				} else if notice.Priority == n.Priority && notice.ID < n.ID {
					typeMap[n.NoticeType] = n
				}
			}
		}
	}

	valid := make([]db_mysql.LoginNotice, 0, len(typeMap))
	for _, v := range typeMap {
		valid = append(valid, v)
	}

	return valid, nil
}

// SetNoticeList 设置公告列表到缓存
func SetNoticeList(data []db_mysql.LoginNotice) {
	globalCacheInstance.Set(CacheKeyLoginNotice, data, cache.NoExpiration)
}

// UpdateNoticeList 更新缓存
func UpdateNoticeList() {
	noticeList, err := db.LoadNotice()
	if err != nil {
		log.Error("UpdateNoticeList: failed to load notice from database, err:%v", err)
		return
	}
	globalCacheInstance.Set(CacheKeyLoginNotice, noticeList, cache.NoExpiration)
}

// ========== 辅助函数 ==========

// genServerKey 生成组合 Key
func genServerKey(clusterID, gameID int64) string {
	return fmt.Sprintf(CacheKeyServer+"_%d_%d", clusterID, gameID)
}

// loadServerList 加载服务器列表
func loadServerList() (any, error) {
	servers, err := db.GetServerList()
	if err != nil {
		log.Error("loadServerList from database failed, err:%v", err)
		return nil, err
	}

	// 将每个服务器单独缓存
	for _, server := range servers {
		key := genServerKey(server.ClusterID, server.GameID)
		globalCacheInstance.Set(key, server, cache.NoExpiration)
	}

	globalCacheInstance.Set(CacheKeyServerList, servers, cache.NoExpiration)
	return servers, nil
}

// loadLoginNotice 加载公告
func loadLoginNotice() (any, error) {
	noticeList, err := db.LoadNotice()
	if err != nil {
		log.Error("loadLoginNotice from database failed, err:%v", err)
		return nil, err
	}
	return noticeList, nil
}

// ========== IP白名单缓存 ==========

// genWhitelistKey 生成白名单缓存键
func genWhitelistKey(apiGroup string) string {
	return CacheKeyWhitelist + "_" + strings.ToLower(apiGroup)
}

// getWhitelistFromCache 从缓存获取指定分组的白名单
func getWhitelistFromCache(apiGroup string) []string {
	key := genWhitelistKey(apiGroup)
	data, exists := globalCacheInstance.Get(key)
	if !exists {
		return nil
	}

	ips, ok := data.([]string)
	if !ok {
		return nil
	}

	// 返回副本，避免外部修改
	result := make([]string, len(ips))
	copy(result, ips)
	return result
}

// setWhitelistToCache 设置指定分组的白名单到缓存
func setWhitelistToCache(apiGroup string, ips []string) {
	key := genWhitelistKey(apiGroup)
	globalCacheInstance.Set(key, ips, cache.NoExpiration)
}

// getAllWhitelistItems 获取所有白名单缓存项（内部使用）
func getAllWhitelistItems() map[string]cache.Item {
	items := globalCacheInstance.Items()
	result := make(map[string]cache.Item)
	prefix := CacheKeyWhitelist + "_"

	for key, item := range items {
		if strings.HasPrefix(key, prefix) {
			result[key] = item
		}
	}
	return result
}
