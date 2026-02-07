CREATE DATABASE IF NOT EXISTS `loginServer` DEFAULT CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci;
USE loginServer;
CREATE TABLE IF NOT EXISTS `game_list` (
    `cluster_id` bigint NOT NULL COMMENT '所属集群ID',
    `game_id` bigint NOT NULL COMMENT '游戏服ID',
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
    `state` tinyint NULL DEFAULT 0 COMMENT '当前状态：0正常账号  1白名单',
    `player_list` JSON NULL COMMENT '玩家游戏服列表数组，包含: cluster_id, game_id, player_id ...',
    `info` text NULL DEFAULT NULL COMMENT '玩家历史额外信息',
    PRIMARY KEY (`account_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic COMMENT = '玩家历史表';
CREATE TABLE IF NOT EXISTS `login_notice` (
    `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `notice_type` TINYINT(3) UNSIGNED NOT NULL DEFAULT '1' COMMENT '公告类型: 1-更新通知, 2-公平运营声明, 3-游戏圈邀请',
    `title` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '公告标题 (Tab页显示的文字)',
    `content` TEXT COMMENT '公告正文内容 (支持HTML或富文本标记)',
    `banner_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '顶部Banner图片的URL地址',
    `priority` INT(11) NOT NULL DEFAULT '0' COMMENT '优先级: 数值越大越靠前',
    `is_enable` TINYINT(1) NOT NULL DEFAULT '1' COMMENT '开关: 0-关闭, 1-开启',
    `start_time` BIGINT(20) NOT NULL COMMENT '开始展示时间',
    `end_time` BIGINT(20) NOT NULL COMMENT '结束展示时间',
    `operator` VARCHAR(32) DEFAULT '' COMMENT '操作人 (GM账号)',
    `created_at` BIGINT(20) NOT NULL COMMENT '创建时间',
    `updated_at` BIGINT(20) NOT NULL COMMENT '最后更新时间',
    `info` text NULL DEFAULT NULL COMMENT '额外信息',
    PRIMARY KEY (`id`),
    KEY `idx_type_time` (`notice_type`, `start_time`, `end_time`) USING BTREE COMMENT '用于快速筛选当前有效的某类公告'
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '登录服公告配置表';
CREATE TABLE IF NOT EXISTS `ip_whitelist` (
    `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `api_group` VARCHAR(64) NOT NULL COMMENT 'API分组名称（如: sgame, adminServer）',
    `ip` VARCHAR(64) NOT NULL COMMENT 'IP地址或CIDR（如: 127.0.0.1 或 192.168.1.0/24）',
    `created_at` BIGINT(20) NOT NULL COMMENT '创建时间',
    `updated_at` BIGINT(20) NOT NULL COMMENT '最后更新时间',
    `info` TEXT NULL DEFAULT NULL COMMENT '额外信息',
    PRIMARY KEY (`id`),
    KEY `idx_api_group` (`api_group`) USING BTREE COMMENT '用于快速查询指定分组的白名单'
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'IP白名单配置表';