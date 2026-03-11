package db

import (
	"context"
	"encoding/json"
	"loginServer/src/db/db_mysql"
	"loginServer/src/db/db_redis"
	"loginServer/src/log"
	"time"
)

// CreateLoginNotice 创建公告
func CreateLoginNotice(notice db_mysql.LoginNotice) error {
	return db_mysql.CreateLoginNotice(notice)
}

// DeleteLoginNotice 删除公告
func DeleteLoginNotice(id uint64) error {
	return db_mysql.DeleteLoginNotice(id)
}

// BatchDeleteLoginNotice 批量删除公告
func BatchDeleteLoginNotice(ids []uint64) error {
	return db_mysql.BatchDeleteLoginNotice(ids)
}

// UpdateLoginNotice 更新公告
func UpdateLoginNotice(notice db_mysql.LoginNotice) error {
	return db_mysql.UpdateLoginNotice(notice)
}

// FindLoginNotice 单条查询
func FindLoginNotice(id uint64) (db_mysql.LoginNotice, error) {
	return db_mysql.FindLoginNotice(id)
}

// GetLoginNoticeList 分页查询 (带搜索)
func GetLoginNoticeList(page, pageSize int, title string, noticeType int, isEnable *int) ([]db_mysql.LoginNotice, int64, error) {
	return db_mysql.GetLoginNoticeList(page, pageSize, title, noticeType, isEnable)
}

// LoadNotice 从数据库加载数据
func LoadNotice() ([]db_mysql.LoginNotice, error) {
	return db_mysql.LoadNotice()
}

// GetLoginNotice 获取公告（优先 Redis 缓存，按时间与优先级筛选）
func GetLoginNotice() ([]db_mysql.LoginNotice, error) {
	ctx := context.Background()
	noticeCache, err := db_redis.GetOrLoad(ctx, db_redis.CacheKeyLoginNotice, func() ([]db_mysql.LoginNotice, error) {
		return LoadNotice()
	})
	if err != nil {
		return nil, err
	}

	typeMap := make(map[int]db_mysql.LoginNotice)
	now := time.Now().Unix()
	for _, n := range noticeCache {
		if n.StartTime <= now && n.EndTime > now {
			notice, exists := typeMap[n.NoticeType]
			if !exists {
				typeMap[n.NoticeType] = n
			} else {
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
	if db_redis.DB == nil {
		return
	}
	ctx := context.Background()
	if b, err := json.Marshal(data); err == nil {
		if setErr := db_redis.Set(ctx, db_redis.CacheKeyLoginNotice, b); setErr != nil {
			log.Warn("SetNoticeList: redis set failed: %v", setErr)
		}
	} else {
		log.Warn("SetNoticeList: marshal failed: %v", err)
	}
}

// UpdateNoticeList 从数据库加载并更新公告缓存
func UpdateNoticeList() {
	noticeList, err := LoadNotice()
	if err != nil {
		log.Error("UpdateNoticeList: failed to load notice from database, err:%v", err)
		return
	}
	SetNoticeList(noticeList)
}
