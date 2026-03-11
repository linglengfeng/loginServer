package db_mysql

import (
	"errors"

	"gorm.io/gorm"
)

// PlayerHistoryItem JSON 数组里的每一个元素（单个游戏服记录）
type PlayerHistoryItem struct {
	ClusterID int   `json:"cluster_id"`
	GameID    int   `json:"game_id"`
	PlayerID  int64 `json:"player_id"`
}

// UserPlayerHistory 对应 user_player_history 表
type UserPlayerHistory struct {
	AccountID  string              `gorm:"column:account_id;primaryKey" json:"account_id"`
	State      int                 `gorm:"column:state; default:0" json:"state"` // 0正常账号 1白名单
	PlayerList []PlayerHistoryItem `gorm:"column:player_list;serializer:json" json:"player_list"`
	Info       *string             `gorm:"column:info" json:"info"`
}

// GetUserHistory 获取玩家的历史记录
func GetUserHistory(accountID string) (UserPlayerHistory, error) {
	var history UserPlayerHistory
	err := DB.Where("account_id = ?", accountID).First(&history).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return UserPlayerHistory{
			AccountID:  accountID,
			PlayerList: []PlayerHistoryItem{},
		}, nil
	}

	return history, err
}

// SetUserHistory 保存/更新玩家历史
func SetUserHistory(accountID string, newItem PlayerHistoryItem) error {
	var history UserPlayerHistory

	err := DB.Where("account_id = ?", accountID).First(&history).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		history = UserPlayerHistory{
			AccountID:  accountID,
			PlayerList: []PlayerHistoryItem{newItem},
		}
		return DB.Create(&history).Error
	}
	if err != nil {
		return err
	}

	history.PlayerList = mergeAndSortList(history.PlayerList, newItem)
	return DB.Save(&history).Error
}

func mergeAndSortList(list []PlayerHistoryItem, target PlayerHistoryItem) []PlayerHistoryItem {
	foundIdx := -1
	for i, item := range list {
		if item.ClusterID == target.ClusterID && item.GameID == target.GameID {
			foundIdx = i
			break
		}
	}
	if foundIdx != -1 {
		list = append(list[:foundIdx], list[foundIdx+1:]...)
	}
	list = append([]PlayerHistoryItem{target}, list...)
	return list
}

// SetUserState 修改用户账号状态
func SetUserState(accountID string, state int) error {
	return DB.Model(&UserPlayerHistory{}).
		Where("account_id = ?", accountID).
		Update("state", state).Error
}
