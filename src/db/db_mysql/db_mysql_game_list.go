package db_mysql

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GameList 单个游戏服结构（对应 game_list 表）
type GameList struct {
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
	err := DB.Find(&servers).Error
	return servers, err
}

// UpdateServerState 仅更新状态（维护/流畅等）
func UpdateServerState(servers []GameList) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		for _, server := range servers {
			if server.State == nil {
				continue
			}
			err := tx.Model(&GameList{}).
				Where("cluster_id = ? And game_id = ?", server.ClusterID, server.GameID).
				Update("state", *server.State).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpdateServerInfo 批量上报
func BatchUpdateServerInfo(servers []GameList) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		for _, server := range servers {
			if err := updateServerInfoWithTx(tx, server); err != nil {
				return err
			}
		}
		return nil
	})
}

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
	// 当无可更新字段时，用主键做 no-op 更新，避免 DoUpdates 为空导致 SQL 异常
	if len(cols) == 0 {
		cols = []string{"cluster_id"}
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "cluster_id"}, {Name: "game_id"}},
		DoUpdates: clause.AssignmentColumns(cols),
	}).Create(&server).Error
}
