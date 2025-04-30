-- 关注社区表
CREATE TABLE IF NOT EXISTS community_follow (
    user_id      BIGINT NOT NULL,
    community_id BIGINT NOT NULL,
    create_time  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, community_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
