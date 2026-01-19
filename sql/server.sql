CREATE DATABASE IF NOT EXISTS `loginServer` DEFAULT CHARACTER
SET = utf8mb4 COLLATE = utf8mb4_general_ci;
USE loginServer;
CREATE TABLE IF NOT EXISTS `game_list` (
    `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
    `cluster_id` bigint NULL DEFAULT NULL COMMENT '所属集群ID',
    `game_id` bigint NULL DEFAULT NULL COMMENT '游戏服ID',
    `name` varchar(255) NULL DEFAULT NULL COMMENT '服务器名称',
    `is_show` tinyint NULL DEFAULT 1 COMMENT '是否展示：0否 1是',
    `state` bigint NULL DEFAULT NULL COMMENT '当前状态：1维护 2流畅 3爆满',
    `is_new` tinyint NULL DEFAULT 0 COMMENT '新服标记：0否 1是',
    `login_url` varchar(255) NULL DEFAULT NULL COMMENT '游戏服登录地址端口',
    `desc` varchar(255) NULL DEFAULT NULL COMMENT '描述',
    `info` text NULL DEFAULT NULL COMMENT '额外信息',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_game_list_cluster_game`(`cluster_id` ASC, `game_id` ASC),
    INDEX `idx_game_list_cluster_id`(`cluster_id` ASC),
    INDEX `idx_game_list_state`(`state` ASC)
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic COMMENT = '游戏服务器列表';