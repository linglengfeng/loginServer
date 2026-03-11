package db_mysql

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// IPWhitelist IP 白名单结构（对应 ip_whitelist 表）
type IPWhitelist struct {
	ID        uint64 `gorm:"primaryKey;column:id" json:"id"`
	APIGroup  string `gorm:"column:api_group;not null;index:idx_api_group" json:"api_group"`
	IP        string `gorm:"column:ip;not null" json:"ip"`
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at" json:"updated_at"`
	Info      string `gorm:"column:info" json:"info"`
}

// LoadWhitelist 从数据库加载指定分组的白名单
func LoadWhitelist(apiGroup string) ([]IPWhitelist, error) {
	var list []IPWhitelist
	err := DB.Where("api_group = ?", apiGroup).Order("id ASC").Find(&list).Error
	return list, err
}

// LoadAllWhitelists 从数据库加载所有分组的白名单
func LoadAllWhitelists() ([]IPWhitelist, error) {
	var list []IPWhitelist
	err := DB.Order("api_group ASC, id ASC").Find(&list).Error
	return list, err
}

// SetWhitelist 设置指定分组的 IP 白名单（完全替换）
func SetWhitelist(apiGroup string, ips []string) error {
	now := time.Now().Unix()
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("api_group = ?", apiGroup).Delete(&IPWhitelist{}).Error; err != nil {
			return err
		}

		if len(ips) == 0 {
			return nil
		}

		whitelists := make([]IPWhitelist, 0, len(ips))
		for _, ip := range ips {
			whitelists = append(whitelists, IPWhitelist{
				APIGroup:  apiGroup,
				IP:        ip,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}

		return tx.Create(&whitelists).Error
	})
}

// AddWhitelistIP 向指定分组添加 IP（已存在则仅更新更新时间）
func AddWhitelistIP(apiGroup string, ip string) error {
	now := time.Now().Unix()

	var existing IPWhitelist
	err := DB.Where("api_group = ? AND ip = ?", apiGroup, ip).First(&existing).Error
	if err == nil {
		existing.UpdatedAt = now
		return DB.Save(&existing).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		whitelist := IPWhitelist{
			APIGroup:  apiGroup,
			IP:        ip,
			CreatedAt: now,
			UpdatedAt: now,
		}
		return DB.Create(&whitelist).Error
	}
	return err
}

// RemoveWhitelistIP 从指定分组删除 IP
func RemoveWhitelistIP(apiGroup string, ip string) error {
	return DB.Where("api_group = ? AND ip = ?", apiGroup, ip).Delete(&IPWhitelist{}).Error
}

// RemoveWhitelistGroup 删除整个分组的所有 IP
func RemoveWhitelistGroup(apiGroup string) error {
	return DB.Where("api_group = ?", apiGroup).Delete(&IPWhitelist{}).Error
}
