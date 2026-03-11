package db_mysql

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// LoginNotice 登录服公告结构（对应 login_notice 表）
type LoginNotice struct {
	ID         uint64 `gorm:"primaryKey;column:id" json:"id"`
	NoticeType int    `gorm:"column:notice_type" json:"notice_type"` // 1,2,3
	Title      string `gorm:"column:title" json:"title"`
	Content    string `gorm:"column:content" json:"content"`
	BannerURL  string `gorm:"column:banner_url" json:"banner_url"`
	Priority   int    `gorm:"column:priority" json:"priority"`
	IsEnable   int    `gorm:"column:is_enable" json:"is_enable"`
	StartTime  int64  `gorm:"column:start_time" json:"start_time"`
	EndTime    int64  `gorm:"column:end_time" json:"end_time"`
	Operator   string `gorm:"column:operator" json:"operator"`
	CreatedAt  int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at" json:"updated_at"`
	Info       string `gorm:"column:info" json:"info"`
}

// LoadNotice 从数据库加载数据（已开启且未过期）
func LoadNotice() ([]LoginNotice, error) {
	var list []LoginNotice
	now := time.Now().Unix()

	err := DB.Where("is_enable = ? AND end_time > ?", 1, now).
		Order("priority DESC, id DESC").
		Find(&list).Error

	return list, err
}

// CreateLoginNotice 创建公告
func CreateLoginNotice(notice LoginNotice) error {
	return DB.Create(&notice).Error
}

// DeleteLoginNotice 删除公告
func DeleteLoginNotice(id uint64) error {
	return DB.Delete(&LoginNotice{}, id).Error
}

// BatchDeleteLoginNotice 批量删除
func BatchDeleteLoginNotice(ids []uint64) error {
	return DB.Delete(&LoginNotice{}, ids).Error
}

// UpdateLoginNotice 更新公告
func UpdateLoginNotice(notice LoginNotice) error {
	return DB.Save(&notice).Error
}

// FindLoginNotice 单条查询
func FindLoginNotice(id uint64) (LoginNotice, error) {
	var notice LoginNotice
	err := DB.First(&notice, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return LoginNotice{}, nil
	}
	return notice, err
}

// GetLoginNoticeList 分页查询（带搜索）
func GetLoginNoticeList(page, pageSize int, title string, noticeType int, isEnable *int) ([]LoginNotice, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var list []LoginNotice
	var total int64

	tx := DB.Model(&LoginNotice{})

	if title != "" {
		tx = tx.Where("title LIKE ?", "%"+title+"%")
	}
	if noticeType != 0 {
		tx = tx.Where("notice_type = ?", noticeType)
	}
	if isEnable != nil {
		tx = tx.Where("is_enable = ?", *isEnable)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := tx.Order("id DESC").Limit(pageSize).Offset(offset).Find(&list).Error
	return list, total, err
}
