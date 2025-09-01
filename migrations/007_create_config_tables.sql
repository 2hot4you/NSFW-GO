-- 创建配置分类表
CREATE TABLE IF NOT EXISTS config_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建配置模板表
CREATE TABLE IF NOT EXISTS config_templates (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    type VARCHAR(20) NOT NULL,
    default_value TEXT,
    category VARCHAR(50),
    required BOOLEAN DEFAULT FALSE,
    is_secret BOOLEAN DEFAULT FALSE,
    validation_rule TEXT,
    options TEXT,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建配置存储表
CREATE TABLE IF NOT EXISTS config_store (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    type VARCHAR(20) NOT NULL,
    description VARCHAR(500),
    category VARCHAR(50),
    is_secret BOOLEAN DEFAULT FALSE,
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建配置备份表
CREATE TABLE IF NOT EXISTS config_backups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    config_data TEXT NOT NULL,
    version VARCHAR(20),
    created_by VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_config_store_category ON config_store(category);
CREATE INDEX IF NOT EXISTS idx_config_templates_category ON config_templates(category);

-- 插入默认配置分类
INSERT INTO config_categories (name, display_name, description, sort_order) VALUES
('server', '服务器配置', 'HTTP服务器相关配置', 1),
('database', '数据库配置', '数据库连接和性能配置', 2),
('redis', 'Redis配置', 'Redis缓存配置', 3),
('media', '媒体库配置', '本地媒体库扫描配置', 4),
('crawler', '爬虫配置', '网站爬取相关配置', 5),
('security', '安全配置', '安全和认证相关配置', 6),
('bot', 'Bot配置', 'Telegram Bot配置', 7),
('torrent', '种子下载配置', '种子搜索和下载配置', 8),
('notifications', '通知配置', '邮件和消息通知配置', 9),
('log', '日志配置', '日志记录相关配置', 10),
('sites', '站点配置', '各个影视网站的配置', 11),
('download', '下载配置', '文件下载相关配置', 12),
('dev', '开发配置', '开发和调试相关配置', 99)
ON CONFLICT (name) DO NOTHING;

-- 插入配置模板
INSERT INTO config_templates (key, name, description, type, default_value, category, required) VALUES
-- 服务器配置
('server.host', '服务器主机', '服务器绑定的IP地址', 'string', '"0.0.0.0"', 'server', true),
('server.port', '服务器端口', '服务器监听端口', 'int', '8080', 'server', true),
('server.mode', '运行模式', '服务器运行模式：debug/release/test', 'string', '"debug"', 'server', true),
('server.read_timeout', '读取超时', '请求读取超时时间', 'string', '"30s"', 'server', false),
('server.write_timeout', '写入超时', '响应写入超时时间', 'string', '"30s"', 'server', false),
('server.enable_cors', '启用CORS', '是否启用跨域资源共享', 'bool', 'true', 'server', false),
('server.enable_swagger', '启用Swagger', '是否启用API文档', 'bool', 'true', 'server', false),

-- 数据库配置
('database.host', '数据库主机', 'PostgreSQL数据库主机地址', 'string', '"localhost"', 'database', true),
('database.port', '数据库端口', 'PostgreSQL数据库端口', 'int', '5432', 'database', true),
('database.user', '数据库用户', '数据库用户名', 'string', '"nsfw"', 'database', true),
('database.password', '数据库密码', '数据库密码', 'string', '"nsfw123"', 'database', true),
('database.dbname', '数据库名称', '数据库名称', 'string', '"nsfw_db"', 'database', true),
('database.sslmode', 'SSL模式', 'SSL连接模式', 'string', '"disable"', 'database', false),
('database.max_open_conns', '最大连接数', '数据库最大开放连接数', 'int', '25', 'database', false),
('database.max_idle_conns', '最大空闲连接数', '数据库最大空闲连接数', 'int', '10', 'database', false),
('database.max_lifetime', '连接最大生存时间', '连接最大生存时间(秒)', 'int', '3600', 'database', false),

-- Redis配置
('redis.host', 'Redis主机', 'Redis服务器主机地址', 'string', '"localhost"', 'redis', true),
('redis.port', 'Redis端口', 'Redis服务器端口', 'int', '6379', 'redis', true),
('redis.password', 'Redis密码', 'Redis连接密码', 'string', '""', 'redis', false),
('redis.db', 'Redis数据库', 'Redis数据库编号', 'int', '0', 'redis', false),
('redis.pool_size', '连接池大小', 'Redis连接池大小', 'int', '10', 'redis', false),
('redis.min_idle_conns', '最小空闲连接', 'Redis最小空闲连接数', 'int', '5', 'redis', false),

-- 媒体库配置
('media.base_path', '媒体库路径', '本地媒体库根目录路径', 'string', '""', 'media', false),
('media.scan_interval', '扫描间隔', '自动扫描间隔(小时)', 'int', '24', 'media', false),
('media.supported_exts', '支持的文件扩展名', '支持的视频文件扩展名列表', 'array', '[".mp4",".mkv",".avi",".mov",".wmv"]', 'media', false),
('media.min_file_size', '最小文件大小', '最小文件大小(MB)', 'int', '100', 'media', false),
('media.max_file_size', '最大文件大小', '最大文件大小(MB)', 'int', '10240', 'media', false)

ON CONFLICT (key) DO NOTHING;
