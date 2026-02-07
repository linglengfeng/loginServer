package db

import (
	"fmt"
	"loginServer/src/db/db_mysql"
	"loginServer/src/db/db_redis"
)

func Start() error {
	if err := db_mysql.Start(); err != nil {
		return fmt.Errorf("mysql start failed: %w", err)
	}
	db_redis.Start()
	return nil
}

// GetServerList 获取服务器列表
func GetServerList() ([]db_mysql.GameList, error) {
	return db_mysql.GetServerList()
}

// UpdateServerState 仅更新状态（维护/流畅等）
func UpdateServerState(serverReqs []db_mysql.GameList) error {
	return db_mysql.UpdateServerState(serverReqs)
}

// BatchUpdateServerInfo 上报/注册服务器信息
func BatchUpdateServerInfo(serverReqs []db_mysql.GameList) error {
	return db_mysql.BatchUpdateServerInfo(serverReqs)
}

// GetUserHistory 获取用户服务器列表
func GetUserHistory(accountID string) (db_mysql.UserPlayerHistory, error) {
	return db_mysql.GetUserHistory(accountID)
}

// SetUserHistory 保存/更新玩家历史
func SetUserHistory(accountID string, newItem db_mysql.PlayerHistoryItem) error {
	return db_mysql.SetUserHistory(accountID, newItem)
}

// SetUserState 更新用户账号状态
func SetUserState(accountID string, state int) error {
	return db_mysql.SetUserState(accountID, state)
}

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

// ========== IP白名单 ==========

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
