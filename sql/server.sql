CREATE DATABASE IF NOT EXISTS `loginServer` DEFAULT CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci;
USE loginServer;

CREATE TABLE IF NOT EXISTS `game_list` (
    `cluster_id` bigint NOT NULL  COMMENT '所属集群ID',
    `game_id` bigint NOT NULL COMMENT '游戏服ID',
    `cluster_name` varchar(255) NULL DEFAULT NULL COMMENT '集群名称',
    `name` varchar(255) NULL DEFAULT NULL COMMENT '服务器名称',
    `state` tinyint NULL DEFAULT 2 COMMENT '当前状态：1维护 2流畅 3爆满',
    `is_show` tinyint NULL DEFAULT 1 COMMENT '是否展示：0否 1是',
    `is_new` tinyint NULL DEFAULT 0 COMMENT '新服标记：0否 1是',
    `addr` varchar(255) NULL DEFAULT NULL COMMENT '游戏服登录地址',
    `port` int NULL DEFAULT NULL COMMENT '游戏服登录端口',
    `desc` varchar(255) NULL DEFAULT NULL COMMENT '描述',
    `info` text NULL DEFAULT NULL COMMENT '额外信息',
    -- 复合主键
    PRIMARY KEY (`cluster_id`, `game_id`)
    ) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '游戏服务器列表';

CREATE TABLE IF NOT EXISTS `user_player_history` (
    `account_id` varchar(255) NOT NULL COMMENT '玩家账号id',
    `player_list` JSON NULL COMMENT '玩家游戏服列表数组，包含: cluster_id, game_id, player_id ...',
    `info` text NULL DEFAULT NULL COMMENT '玩家历史额外信息',
    PRIMARY KEY (`account_id`)
    ) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic  COMMENT = '玩家历史表';