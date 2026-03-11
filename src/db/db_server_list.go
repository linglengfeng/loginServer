package db

import (
	"context"
	"encoding/json"
	"loginServer/src/db/db_mysql"
	"loginServer/src/db/db_redis"
	"loginServer/src/log"
)

// GetServerList 获取服务器列表（优先 Redis 缓存，未命中回源 MySQL）
func GetServerList() ([]db_mysql.GameList, error) {
	ctx := context.Background()
	gameList, err := db_redis.GetOrLoad(ctx, db_redis.CacheKeyServerList, func() ([]db_mysql.GameList, error) {
		return db_mysql.GetServerList()
	})
	if err != nil {
		log.Error("GetServerList: failed to get server list from cache: %v", err)
		return nil, err
	}
	return gameList, nil
}

// UpdateCacheServerList 更新服务器列表缓存
func UpdateCacheServerList(updates []db_mysql.GameList) {
	if db_redis.DB == nil {
		return
	}
	ctx := context.Background()

	var serverKeyList []db_mysql.GameList
	if data, err := db_redis.Get(ctx, db_redis.CacheKeyServerList); err == nil {
		if unmarshalErr := json.Unmarshal(data, &serverKeyList); unmarshalErr != nil {
			log.Warn("UpdateCacheServerList: unmarshal server list failed: %v", unmarshalErr)
			serverKeyList = nil
		}
	}

	if serverKeyList == nil {
		serverKeyList = make([]db_mysql.GameList, 0)
	}

	listChanged := false
	serverMap := make(map[string]int)

	for i, server := range serverKeyList {
		key := db_redis.GenServerKey(server.ClusterID, server.GameID)
		serverMap[key] = i
	}

	for _, update := range updates {
		key := db_redis.GenServerKey(update.ClusterID, update.GameID)

		if idx, exists := serverMap[key]; exists {
			updatedServer := serverKeyList[idx]

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

			serverKeyList[idx] = updatedServer
			listChanged = true
		} else {
			serverMap[key] = len(serverKeyList)
			serverKeyList = append(serverKeyList, update)
			listChanged = true
		}
	}

	if listChanged {
		if b, err := json.Marshal(serverKeyList); err == nil {
			if setErr := db_redis.Set(ctx, db_redis.CacheKeyServerList, b); setErr != nil {
				log.Warn("UpdateCacheServerList: redis set server list failed: %v", setErr)
			}
		} else {
			log.Warn("UpdateCacheServerList: marshal server list failed: %v", err)
		}
	}
}

// UpdateServerState 仅更新状态（维护/流畅等）
func UpdateServerState(serverReqs []db_mysql.GameList) error {
	return db_mysql.UpdateServerState(serverReqs)
}

// BatchUpdateServerInfo 上报/注册服务器信息
func BatchUpdateServerInfo(serverReqs []db_mysql.GameList) error {
	return db_mysql.BatchUpdateServerInfo(serverReqs)
}
