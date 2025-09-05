-- 创建排行榜下载任务表
CREATE TABLE ranking_download_tasks (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    code VARCHAR(50) NOT NULL,
    title VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    torrent_url VARCHAR(2000),
    torrent_hash VARCHAR(100),
    progress DECIMAL(3,2) DEFAULT 0,
    error_msg VARCHAR(1000),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    file_size BIGINT DEFAULT 0,
    downloaded_size BIGINT DEFAULT 0,
    source VARCHAR(50) DEFAULT 'manual',
    rank_type VARCHAR(20)
);

-- 创建排行榜下载任务索引
CREATE UNIQUE INDEX idx_ranking_download_tasks_code ON ranking_download_tasks(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_ranking_download_tasks_status ON ranking_download_tasks(status);
CREATE INDEX idx_ranking_download_tasks_source ON ranking_download_tasks(source);
CREATE INDEX idx_ranking_download_tasks_rank_type ON ranking_download_tasks(rank_type);
CREATE INDEX idx_ranking_download_tasks_created_at ON ranking_download_tasks(created_at);
CREATE INDEX idx_ranking_download_tasks_deleted_at ON ranking_download_tasks(deleted_at);

-- 创建订阅配置表
CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    rank_type VARCHAR(20) NOT NULL,
    enabled BOOLEAN DEFAULT FALSE,
    hourly_limit INTEGER DEFAULT 10,
    daily_limit INTEGER DEFAULT 50,
    last_run_at TIMESTAMP,
    last_check_at TIMESTAMP,
    total_downloads INTEGER DEFAULT 0,
    success_downloads INTEGER DEFAULT 0
);

-- 创建订阅配置索引
CREATE UNIQUE INDEX idx_subscriptions_rank_type ON subscriptions(rank_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_subscriptions_enabled ON subscriptions(enabled);
CREATE INDEX idx_subscriptions_deleted_at ON subscriptions(deleted_at);

-- 创建订阅限制记录表
CREATE TABLE subscription_limits (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    rank_type VARCHAR(20) NOT NULL,
    limit_type VARCHAR(20) NOT NULL,
    count INTEGER DEFAULT 0,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL
);

-- 创建订阅限制记录索引
CREATE INDEX idx_subscription_limits_rank_type ON subscription_limits(rank_type);
CREATE INDEX idx_subscription_limits_limit_type ON subscription_limits(limit_type);
CREATE INDEX idx_subscription_limits_period_start ON subscription_limits(period_start);
CREATE INDEX idx_subscription_limits_period_end ON subscription_limits(period_end);
CREATE INDEX idx_subscription_limits_deleted_at ON subscription_limits(deleted_at);

-- 插入默认订阅配置
INSERT INTO subscriptions (rank_type, enabled, hourly_limit, daily_limit, created_at, updated_at) VALUES
    ('daily', FALSE, 10, 50, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('weekly', FALSE, 10, 50, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('monthly', FALSE, 10, 50, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);