package db_mysql

import (
	"errors"
	"loginServer/config"
	"loginServer/pkg/mysql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DB *gorm.DB

func Start() {
	mysqlip := config.Config.GetString("mysql.ip")
	if mysqlip != "" {
		mysqlink := mysql.Link{
			User:     config.Config.GetString("mysql.user"),
			Password: config.Config.GetString("mysql.password"),
			Ip:       mysqlip,
			Port:     config.Config.GetString("mysql.port"),
			Db:       config.Config.GetString("mysql.db"),
		}
		DB = mysql.Start(mysqlink)
	}
}

// GameList 单个游戏服结构
type GameList struct {
	// 注意：复合主键两个字段都要加上 primaryKey
	ClusterID   int64   `gorm:"column:cluster_id;primaryKey;autoIncrement:false" json:"cluster_id"`
	GameID      int64   `gorm:"column:game_id;primaryKey;autoIncrement:false" json:"game_id"`
	ClusterName *string `gorm:"column:cluster_name" json:"cluster_name"`
	Name        *string `gorm:"column:name" json:"name"`
	State       *int    `gorm:"column:state;default:1" json:"state"`     // 1维护 2流畅 3爆满
	IsShow      *int    `gorm:"column:is_show;default:1" json:"is_show"` // 0否 1是
	IsNew       *int    `gorm:"column:is_new;default:0" json:"is_new"`   // 0否 1是
	Addr        *string `gorm:"column:addr" json:"addr"`
	Port        *int    `gorm:"column:port" json:"port"`
	Desc        *string `gorm:"column:desc" json:"desc"`
	Info        *string `gorm:"column:info" json:"info"`
}

// GetServerList 获取服务器列表
func GetServerList() ([]GameList, error) {
	var servers []GameList
	// 按照主键顺序读取，无需额外的 ORDER BY 也会自动按 cluster_id, game_id 排序
	err := DB.Find(&servers).Error
	return servers, err
}

// UpdateServerState 仅更新状态（维护/流畅等）
func UpdateServerState(clusterID int64, gameID int64, newState *int) error {
	// SQL: UPDATE game_list SET state = ? WHERE cluster_id = ? AND game_id = ?
	// GORM 会自动识别 GameList 结构体里的复合主键值作为 WHERE 条件
	return DB.Model(&GameList{ClusterID: clusterID, GameID: gameID}).
		Update("state", newState).Error
}

// BatchUpdateServerInfo 批量上报
func BatchUpdateServerInfo(servers []*GameList) error {
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
func updateServerInfoWithTx(tx *gorm.DB, server *GameList) error {
	var cols []string
	if server.ClusterName != nil {
		cols = append(cols, "cluster_name")
	}
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
	}).Create(server).Error
}

// PlayerHistoryItem JSON 数组里的每一个元素（单个游戏服记录）
type PlayerHistoryItem struct {
	ClusterID  int     `json:"cluster_id"`
	GameID     int     `json:"game_id"`
	PlayerID   int64   `json:"player_id"`
	PlayerName *string `json:"player_name"`
	Level      *int    `json:"level"`
	Avatar     *int    `json:"avatar"`
	LoginTime  *int    `json:"login_time"` // 记录该角色具体什么时候登录的
}

// UserPlayerHistory 对应数据库表
type UserPlayerHistory struct {
	AccountID  string              `gorm:"column:account_id;primaryKey" json:"account_id"`
	PlayerList []PlayerHistoryItem `gorm:"column:player_list;serializer:json" json:"player_list"`
	Info       *string             `gorm:"column:info" json:"info"`
}

// GetUserHistory 获取玩家的历史记录
func GetUserHistory(accountID string) (*UserPlayerHistory, error) {
	var history UserPlayerHistory

	// 根据 AccountID 查询
	err := DB.Where("account_id = ?", accountID).First(&history).Error

	// 如果记录不存在，返回一个空的结构体（方便前端处理，不要返回 nil）
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &UserPlayerHistory{
			AccountID:  accountID,
			PlayerList: []PlayerHistoryItem{},
		}, nil
	}

	return &history, err
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
			if target.PlayerName == nil {
				target.PlayerName = item.PlayerName // 保持旧名字
			}
			if target.Level == nil {
				target.Level = item.Level // 保持旧等级
			}
			if target.Avatar == nil {
				target.Avatar = item.Avatar // 保持旧头像
			}
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
