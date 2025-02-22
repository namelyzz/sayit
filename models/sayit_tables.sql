DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
                        `id` bigint(20) NOT NULL AUTO_INCREMENT,
                        `user_id` bigint(20) NOT NULL,
                        `username` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
                        `password` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
                        `email` varchar(64) COLLATE utf8mb4_general_ci,
                        `gender` tinyint(4) NOT NULL DEFAULT '0',
                        `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                        `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE
CURRENT_TIMESTAMP,
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `idx_username` (`username`) USING BTREE,
                        UNIQUE KEY `idx_user_id` (`user_id`) USING BTREE
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