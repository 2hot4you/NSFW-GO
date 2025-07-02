-- +migrate Up
-- 创建rankings表用于存储排行榜数据
CREATE TABLE IF NOT EXISTS rankings (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    cover_url VARCHAR(1000),
    rank_type VARCHAR(20) NOT NULL,
    position INTEGER NOT NULL,
    crawled_at TIMESTAMP NOT NULL,
    local_exists BOOLEAN DEFAULT FALSE,
    last_checked TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_rankings_code ON rankings(code);
CREATE INDEX IF NOT EXISTS idx_rankings_rank_type ON rankings(rank_type);
CREATE INDEX IF NOT EXISTS idx_rankings_position ON rankings(position);
CREATE INDEX IF NOT EXISTS idx_rankings_crawled_at ON rankings(crawled_at);
CREATE INDEX IF NOT EXISTS idx_rankings_local_exists ON rankings(local_exists);
CREATE INDEX IF NOT EXISTS idx_rankings_deleted_at ON rankings(deleted_at);

-- 创建组合索引优化查询
CREATE INDEX IF NOT EXISTS idx_rankings_type_position ON rankings(rank_type, position);
CREATE INDEX IF NOT EXISTS idx_rankings_type_crawled ON rankings(rank_type, crawled_at DESC);

-- +migrate Down
-- 删除rankings表
DROP TABLE IF EXISTS rankings; 