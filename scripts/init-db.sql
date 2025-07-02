-- NSFW-Go 数据库初始化脚本
-- 创建必要的扩展和设置

-- 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 设置时区
SET timezone = 'Asia/Shanghai';

-- 创建应用用户（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'nsfw') THEN
        CREATE ROLE nsfw WITH LOGIN PASSWORD 'nsfw123';
    END IF;
END
$$;

-- 授权
GRANT ALL PRIVILEGES ON DATABASE nsfw_db TO nsfw;
GRANT ALL ON SCHEMA public TO nsfw;

-- 输出初始化完成信息
\echo '数据库初始化完成' 