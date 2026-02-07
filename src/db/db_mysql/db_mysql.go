package db_mysql

import (
	"errors"
	"loginServer/config"
	"loginServer/pkg/mysql"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DB *gorm.DB

func Start() error {
	mysqlip := config.Config.GetString("mysql.ip")
	if mysqlip != "" {
		mysqlink := mysql.Link{
			User:     config.Config.GetString("mysql.user"),
			Password: config.Config.GetString("mysql.password"),
			Ip:       mysqlip,
			Port:     config.Config.GetString("mysql.port"),
			Db:       config.Config.GetString("mysql.db"),
		}
		db, err := mysql.Start(mysqlink)
		if err != nil {
			return err
		}
		DB = db
	}
	return nil
}

// GameList 单个游戏服结构
type GameList struct {
	// 注意：复合主键两个字段都要加上 primaryKey
	ClusterID int64   `gorm:"column:cluster_id;primaryKey;autoIncrement:false" json:"cluster_id"`
	GameID    int64   `gorm:"column:game_id;primaryKey;autoIncrement:false" json:"game_id"`
	Name      *string `gorm:"column:name" json:"name"`
	State     *int    `gorm:"column:state;default:1" json:"state"`     // 1维护 2流畅 3爆满
	IsShow    *int    `gorm:"column:is_show;default:1" json:"is_show"` // 0否 1是
	IsNew     *int    `gorm:"column:is_new;default:0" json:"is_new"`   // 0否 1是
	Addr      *string `gorm:"column:addr" json:"addr"`
	Port      *int    `gorm:"column:port" json:"port"`
	Desc      *string `gorm:"column:desc" json:"desc"`
	Info      *string `gorm:"column:info" json:"info"`
}

// GetServerList 获取服务器列表
func GetServerList() ([]GameList, error) {
	var servers []GameList
	// 按照主键顺序读取，无需额外的 ORDER BY 也会自动按 cluster_id, game_id 排序
	err := DB.Find(&servers).Error
	return servers, err
}

// UpdateServerState 仅更新状态（维护/流畅等）
func UpdateServerState(servers []GameList) error {
	// SQL: UPDATE game_list SET state = ? WHERE cluster_id = ? AND game_id = ?
	// GORM 会自动识别 GameList 结构体里的复合主键值作为 WHERE 条件
	return DB.Transaction(func(tx *gorm.DB) error {
		for _, server := range servers {
			if server.State == nil {
				continue
			}
			err := tx.Model(&GameList{}).
				Where("cluster_id = ? And game_id = ?", server.ClusterID, server.GameID).
				Update("state", *server.State).Error
			if err != nil {
				return err // 只要有一条出错，回滚整个事务
			}
		}
		return nil
	})
}

// BatchUpdateServerInfo 批量上报
func BatchUpdateServerInfo(servers []GameList) error {
	// 开启事务
	return DB.Transaction(func(tx *gorm.DB) error {
		for _, server := range servers {
			// 复用 tx 执行单条 Upsert
			if err := updateServerInfoWithTx(tx, server); err != nil {
				return err // 只要有一条失败，回滚整个事务
			}
		}
		return nil
	})
}

// 私有辅助函数：使用传入的 tx 执行单条更新
func updateServerInfoWithTx(tx *gorm.DB, server GameList) error {
	var cols []string
	if server.Name != nil {
		cols = append(cols, "name")
	}
	if server.Addr != nil {
		cols = append(cols, "addr")
	}
	if server.Port != nil {
		cols = append(cols, "port")
	}
	if server.State != nil {
		cols = append(cols, "state")
	}
	if server.IsShow != nil {
		cols = append(cols, "is_show")
	}
	if server.IsNew != nil {
		cols = append(cols, "is_new")
	}
	if server.Desc != nil {
		cols = append(cols, "desc")
	}
	if server.Info != nil {
		cols = append(cols, "info")
	}
	return tx.Clauses(clause.OnConflict{
		// 1. 冲突检测：复合主键 (cluster_id, game_id)
		Columns: []clause.Column{{Name: "cluster_id"}, {Name: "game_id"}},
		// 2. 冲突处理 (Upsert)：
		DoUpdates: clause.AssignmentColumns(cols),
	}).Create(&server).Error
}

// PlayerHistoryItem JSON 数组里的每一个元素（单个游戏服记录）
type PlayerHistoryItem struct {
	ClusterID int   `json:"cluster_id"`
	GameID    int   `json:"game_id"`
	PlayerID  int64 `json:"player_id"`
}

// UserPlayerHistory 对应数据库表
type UserPlayerHistory struct {
	AccountID  string              `gorm:"column:account_id;primaryKey" json:"account_id"`
	State      int                 `gorm:"column:state; default:0" json:"state"` //0 正常账号  1 白名单
	PlayerList []PlayerHistoryItem `gorm:"column:player_list;serializer:json" json:"player_list"`
	Info       *string             `gorm:"column:info" json:"info"`
}

// GetUserHistory 获取玩家的历史记录
func GetUserHistory(accountID string) (UserPlayerHistory, error) {
	var history UserPlayerHistory

	// 根据 AccountID 查询
	err := DB.Where("account_id = ?", accountID).First(&history).Error

	// 如果记录不存在，返回一个空的结构体（方便前端处理，不要返回 nil）
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return UserPlayerHistory{
			AccountID:  accountID,
			PlayerList: []PlayerHistoryItem{},
		}, nil
	}

	return history, err
}

// SetUserHistory 保存/更新玩家历史
// 场景：玩家登录游戏服成功、或者升级时调用
// 逻辑：取出旧数据 -> 更新列表 -> 覆盖写回数据库
func SetUserHistory(accountID string, newItem PlayerHistoryItem) error {
	var history UserPlayerHistory

	// 先查询是否存在记录
	err := DB.Where("account_id = ?", accountID).First(&history).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果是新用户：直接创建
		history = UserPlayerHistory{
			AccountID:  accountID,
			PlayerList: []PlayerHistoryItem{newItem},
		}
		return DB.Create(&history).Error
	} else if err != nil {
		// 数据库报错
		return err
	}

	// 如果是老用户更新切片
	// 逻辑：如果这个服已经在列表里，更新信息；如果不在，追加。
	history.PlayerList = mergeAndSortList(history.PlayerList, newItem)

	// 2. 写回数据库 (Save 会自动处理 UPDATE)
	// GORM 自动将 Slice -> JSON
	return DB.Save(&history).Error
}

// mergeAndSortList 列表去重、更新、排序
func mergeAndSortList(list []PlayerHistoryItem, target PlayerHistoryItem) []PlayerHistoryItem {
	foundIdx := -1
	// 1. 寻找是否已存在该服记录
	for i, item := range list {
		if item.ClusterID == target.ClusterID && item.GameID == target.GameID {
			foundIdx = i
			break
		}
	}
	// 2. 更新列表
	if foundIdx != -1 {
		// 如果找到了：更新该条目，并把它移动到数组最前面（最近登录的在最前）
		// 删除旧的位置
		list = append(list[:foundIdx], list[foundIdx+1:]...)
	}
	// 将最新的插入到头部 (Prepend)
	// 这样前端拿到数组时，第一个就是最近玩的服
	list = append([]PlayerHistoryItem{target}, list...)
	return list
}

// SetUserState 修改用户账号状态
func SetUserState(accountID string, state int) error {
	// 根据 accountID 更新
	err := DB.Model(&UserPlayerHistory{}).
		Where("account_id = ?", accountID).
		Update("state", state).Error

	if err != nil {
		return err
	}

	return nil
}

// LoginNotice 登录服公告结构
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

// LoadNotice 从数据库加载数据到缓存
// 只获取 [已开启] 且 [未过期] 的数据
func LoadNotice() ([]LoginNotice, error) {
	var list []LoginNotice

	// 获取当前时间戳
	now := time.Now().Unix()

	// SQL 逻辑：
	// 1. is_enable = 1 (必须是开启的)
	// 2. end_time > now (结束时间必须在未来，意味着还没过期)
	// 注意：这里不判断 start_time，因为我们允许把“明天开始”的公告也缓存在内存里，
	//      等到了明天时间一到，GetValidNoticesFromCache 就能自动把它刷出来，而不需要重新查库。
	err := DB.Where("is_enable = ? AND end_time > ?", 1, now).
		Order("priority DESC, id DESC"). // 按优先级排序，方便后续处理
		Find(&list).Error

	if err != nil {
		return nil, err
	}

	return list, err
}

// CreateLoginNotice 创建
func CreateLoginNotice(notice LoginNotice) error {
	return DB.Create(&notice).Error
}

// DeleteLoginNotice 删除
func DeleteLoginNotice(id uint64) error {
	return DB.Delete(&LoginNotice{}, id).Error
}

// BatchDeleteLoginNotice 批量删除
func BatchDeleteLoginNotice(ids []uint64) error {
	return DB.Delete(&LoginNotice{}, ids).Error
}

// UpdateLoginNotice 更新
func UpdateLoginNotice(notice LoginNotice) error {
	// 使用 Updates 更新非零值，或者使用 map 更新指定字段
	// 这里GVA 传过来的是完整对象，直接 Save 即可
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

// GetLoginNoticeList 分页查询 (带搜索)
func GetLoginNoticeList(page, pageSize int, title string, noticeType int, isEnable *int) ([]LoginNotice, int64, error) {
	var list []LoginNotice
	var total int64

	tx := DB.Model(&LoginNotice{})

	// 构造查询条件
	if title != "" {
		tx = tx.Where("title LIKE ?", "%"+title+"%")
	}
	if noticeType != 0 {
		tx = tx.Where("notice_type = ?", noticeType)
	}
	if isEnable != nil {
		tx = tx.Where("is_enable = ?", *isEnable)
	}

	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = tx.Order("id DESC").Limit(pageSize).Offset(offset).Find(&list).Error

	return list, total, err
}

// ========== IP白名单 ==========

// IPWhitelist IP白名单结构
type IPWhitelist struct {
	ID        uint64 `gorm:"primaryKey;column:id" json:"id"`
	APIGroup  string `gorm:"column:api_group;not null;index:idx_api_group" json:"api_group"` // API分组名称
	IP        string `gorm:"column:ip;not null" json:"ip"`                                   // IP地址或CIDR
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`                            // 创建时间
	UpdatedAt int64  `gorm:"column:updated_at" json:"updated_at"`                            // 更新时间
	Info      string `gorm:"column:info" json:"info"`                                        // 额外信息
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

// SetWhitelist 设置指定分组的IP白名单（完全替换）
// 先删除该分组的所有IP，然后插入新的IP列表
func SetWhitelist(apiGroup string, ips []string) error {
	now := time.Now().Unix()
	return DB.Transaction(func(tx *gorm.DB) error {
		// 1. 删除该分组的所有现有IP
		if err := tx.Where("api_group = ?", apiGroup).Delete(&IPWhitelist{}).Error; err != nil {
			return err
		}

		// 2. 如果IP列表为空，直接返回（表示清空白名单）
		if len(ips) == 0 {
			return nil
		}

		// 3. 批量插入新的IP列表
		whitelists := make([]IPWhitelist, 0, len(ips))
		for _, ip := range ips {
			whitelists = append(whitelists, IPWhitelist{
				APIGroup:  apiGroup,
				IP:        ip,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}

		if err := tx.Create(&whitelists).Error; err != nil {
			return err
		}

		return nil
	})
}

// AddWhitelistIP 向指定分组添加IP（如果已存在则忽略）
func AddWhitelistIP(apiGroup string, ip string) error {
	now := time.Now().Unix()

	// 检查是否已存在
	var existing IPWhitelist
	err := DB.Where("api_group = ? AND ip = ?", apiGroup, ip).First(&existing).Error
	if err == nil {
		// 已存在，只更新更新时间
		existing.UpdatedAt = now
		return DB.Save(&existing).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 不存在，创建新记录
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

// RemoveWhitelistIP 从指定分组删除IP
func RemoveWhitelistIP(apiGroup string, ip string) error {
	return DB.Where("api_group = ? AND ip = ?", apiGroup, ip).Delete(&IPWhitelist{}).Error
}

// RemoveWhitelistGroup 删除整个分组的所有IP
func RemoveWhitelistGroup(apiGroup string) error {
	return DB.Where("api_group = ?", apiGroup).Delete(&IPWhitelist{}).Error
}
