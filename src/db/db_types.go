package db

import "loginServer/src/db/db_mysql"

// 以下类型从 db_mysql 重新导出，供外部通过 db 包使用，避免直接依赖 db_mysql

// GameList 单个游戏服结构（对应 game_list 表）
type GameList = db_mysql.GameList

// PlayerHistoryItem JSON 数组里的每一个元素（单个游戏服记录）
type PlayerHistoryItem = db_mysql.PlayerHistoryItem

// LoginNotice 登录服公告结构（对应 login_notice 表）
type LoginNotice = db_mysql.LoginNotice
