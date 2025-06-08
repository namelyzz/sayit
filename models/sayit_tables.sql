DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
                        `id` bigint(20) NOT NULL AUTO_INCREMENT,
                        `user_id` bigint(20) NOT NULL,
                        `username` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
                        `password` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
                        `signature` varchar(240) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
                        `email` varchar(64) COLLATE utf8mb4_general_ci,
                        `gender` tinyint(4) NOT NULL DEFAULT '0',
                        `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                        `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE
CURRENT_TIMESTAMP,
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `idx_username` (`username`) USING BTREE,
                        UNIQUE KEY `idx_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

DROP TABLE IF EXISTS `user_follow`;
CREATE TABLE `user_follow` (
                               `follower_id` bigint(20) NOT NULL,
                               `following_id` bigint(20) NOT NULL,
                               `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                               PRIMARY KEY (`follower_id`, `following_id`),
                               KEY `idx_following_id` (`following_id`),
                               KEY `idx_follower_id` (`follower_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

DROP TABLE IF EXISTS `community`;
CREATE TABLE `community` (
                             `id` int(11) NOT NULL AUTO_INCREMENT,
                             `community_id` int(10) unsigned NOT NULL,
                             `community_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL,
                             `introduction` varchar(256) COLLATE utf8mb4_general_ci NOT NULL,
                             `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `idx_community_id` (`community_id`),
                             UNIQUE KEY `idx_community_name` (`community_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


INSERT INTO `community` (`community_id`, `community_name`, `introduction`) VALUES
                                                                               (1, 'GolangStudy', 'Go语言学习交流社区，从入门到精通，分享学习心得和项目经验'),
                                                                               (2, 'KamenRiderFaiz', '假面骑士Faiz粉丝聚集地，讨论剧情、角色、变身器和相关周边'),
                                                                               (3, 'A_Stock', 'A股投资交流社区，分享股市分析、投资策略和市场动态'),
                                                                               (4, 'EnglishSpeaking', '英语口语练习社区，提供口语技巧、发音指导和实战练习机会'),
                                                                               (5, 'WoodworkingDIY', '木工DIY爱好者社区，分享木工技巧、工具使用和创意作品'),
                                                                               (6, 'AnimeLovers', '动漫爱好者天堂，讨论新番推荐、经典回顾和二次元文化'),
                                                                               (7, 'HomeCook', '家常美食制作社区，分享菜谱、烹饪技巧和厨房好物推荐'),
                                                                               (8, 'FitnessBeginner', '健身新手互助社区，提供训练计划、饮食建议和进步分享'),
                                                                               (9, 'DigitalNomad', '数字游民生活方式社区，分享远程工作、旅行经验和装备推荐'),
                                                                               (10, 'PlantParents', '植物养护交流社区，分享种植经验、病虫害防治和绿植搭配');

DROP TABLE IF EXISTS `post`;
CREATE TABLE `post` (
                        `id` bigint(20) NOT NULL AUTO_INCREMENT,
                        `post_id` bigint(20) NOT NULL COMMENT '帖子id',
                        `title` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '标题',
                        `content` varchar(8192) COLLATE utf8mb4_general_ci NOT NULL COMMENT '内容',
                        `author_id` bigint(20) NOT NULL COMMENT '作者的用户id',
                        `community_id` bigint(20) NOT NULL COMMENT '所属社区',
                        `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '帖子状态',
                        `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                        `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `idx_post_id` (`post_id`),
                        KEY `idx_author_id` (`author_id`),
                        KEY `idx_community_id` (`community_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

DROP TABLE IF EXISTS `outbox_events`;
CREATE TABLE `outbox_events` (
                                 `id` bigint(20) NOT NULL AUTO_INCREMENT,
                                 `event_type` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
                                 `aggregate_id` bigint(20) NOT NULL,
                                 `payload` json NOT NULL,
                                 `retry_count` int NOT NULL DEFAULT 0,
                                 `next_retry_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                 `last_error` varchar(512) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
                                 `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                                 `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                 PRIMARY KEY (`id`),
                                 KEY `idx_next_retry_at` (`next_retry_at`),
                                 KEY `idx_event_type` (`event_type`),
                                 KEY `idx_aggregate_id` (`aggregate_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

DROP TABLE IF EXISTS `comment`;
CREATE TABLE `comment` (
                           `id` bigint(20) NOT NULL AUTO_INCREMENT,
                           `comment_id` bigint(20) NOT NULL COMMENT 'comment id (snowflake)',
                           `post_id` bigint(20) NOT NULL COMMENT 'post id',
                           `author_id` bigint(20) NOT NULL COMMENT 'author user id',
                           `parent_id` bigint(20) NOT NULL DEFAULT 0 COMMENT 'parent comment id (0=top level)',
                           `root_id` bigint(20) NOT NULL DEFAULT 0 COMMENT 'root comment id (0=self is root)',
                           `content` varchar(1024) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'content',
                           `like_count` bigint(20) NOT NULL DEFAULT 0 COMMENT 'like count (denormalized)',
                           `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT 'status: 1=normal, 2=deleted',
                           `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                           `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                           PRIMARY KEY (`id`),
                           UNIQUE KEY `idx_comment_id` (`comment_id`),
                           KEY `idx_post_id` (`post_id`),
                           KEY `idx_parent_id` (`parent_id`),
                           KEY `idx_root_id` (`root_id`),
                           KEY `idx_author_id` (`author_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

DROP TABLE IF EXISTS `comment_like`;
CREATE TABLE `comment_like` (
                                `comment_id` bigint(20) NOT NULL COMMENT 'comment id',
                                `user_id` bigint(20) NOT NULL COMMENT 'user id',
                                `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                                PRIMARY KEY (`comment_id`, `user_id`),
                                 KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

DROP TABLE IF EXISTS `notifications`;
CREATE TABLE `notifications` (
                                 `id` bigint(20) NOT NULL AUTO_INCREMENT,
                                 `notification_id` bigint(20) NOT NULL COMMENT 'notification id (snowflake)',
                                 `recipient_id` bigint(20) NOT NULL COMMENT 'notification receiver user id',
                                 `actor_id` bigint(20) NOT NULL COMMENT 'action actor user id',
                                 `type` varchar(32) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'notification type',
                                 `post_id` bigint(20) DEFAULT NULL COMMENT 'related post id',
                                 `comment_id` bigint(20) DEFAULT NULL COMMENT 'related comment id',
                                 `parent_id` bigint(20) DEFAULT NULL COMMENT 'parent comment id',
                                 `direction` tinyint(4) DEFAULT NULL COMMENT 'vote direction: 1=upvote, -1=downvote',
                                 `title` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'notification title',
                                 `content` varchar(512) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'notification content',
                                 `link` varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'frontend link',
                                 `is_read` tinyint(4) NOT NULL DEFAULT 0 COMMENT '0=unread, 1=read',
                                 `dedupe_key` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'idempotency key',
                                 `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                                 `read_time` timestamp NULL DEFAULT NULL,
                                 PRIMARY KEY (`id`),
                                 UNIQUE KEY `uk_notification_id` (`notification_id`),
                                 UNIQUE KEY `uk_dedupe_key` (`dedupe_key`),
                                 KEY `idx_recipient_read_time` (`recipient_id`, `is_read`, `create_time`),
                                 KEY `idx_recipient_time` (`recipient_id`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
