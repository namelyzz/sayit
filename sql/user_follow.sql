CREATE TABLE IF NOT EXISTS user_follow (
    follower_id  BIGINT NOT NULL,
    following_id BIGINT NOT NULL,
    create_time  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, following_id),
    KEY idx_following_id (following_id),
    KEY idx_follower_id (follower_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
