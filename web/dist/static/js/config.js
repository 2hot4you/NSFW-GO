// 全局变量
let currentConfig = null;
let originalConfig = null; // 存储原始配置，用于保护敏感字段
let isLoading = false;

// 敏感字段列表（现在仅用于标记，不再隐藏）
const sensitiveFields = [
    'database-password',
    'redis-password', 
    'telegram-token',
    'security-jwt-secret',
    'security-password-salt',
    'notification-email-password',
    'torrent-jackett-api-key',
    'torrent-qbittorrent-password'
];

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    loadConfig();
    loadBackups();
});

// 标签页切换
function switchTab(tabName) {
    // 移除所有活动状态
    document.querySelectorAll('.config-tab').forEach(btn => btn.classList.remove('active'));
    document.querySelectorAll('.config-section').forEach(section => section.classList.remove('active'));
    
    // 设置当前标签页为活动状态
    event.target.classList.add('active');
    document.getElementById(tabName + '-section').classList.add('active');
}

// 加载配置
async function loadConfig() {
    if (isLoading) return;
    
    setLoading(true);
    try {
        const response = await fetch('/api/v1/config');
        const result = await response.json();
        
        if (result.success) {
            // 首先获取完整配置（包含敏感信息的占位符）
            currentConfig = result.data;
            
            // 存储原始配置用于恢复敏感字段的真实值
            originalConfig = JSON.parse(JSON.stringify(result.data));
            
            populateForm(currentConfig);
            showNotification('配置加载成功', 'success');
        } else {
            showNotification('加载配置失败: ' + result.message, 'error');
        }
    } catch (error) {
        showNotification('加载配置时发生错误: ' + error.message, 'error');
    }
    setLoading(false);
}

// 保存配置
async function saveConfig() {
    if (isLoading) return;
    
    // 先进行配置验证
    const isValid = await validateConfigSilently();
    if (!isValid) {
        showNotification('配置验证失败，请先验证配置后再保存', 'error');
        return;
    }
    
    // 显示确认弹框
    const confirmRestart = confirm('⚠️ 保存配置后将自动重启服务器以应用新配置。\n\n是否确定要继续？');
    if (!confirmRestart) {
        return;
    }
    
    setLoading(true);
    try {
        const config = collectFormData();
        
        const response = await fetch('/api/v1/config', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(config)
        });
        
        const result = await response.json();
        
        if (result.success) {
            currentConfig = config;
            showNotification('配置保存成功，正在重启服务器...', 'success');
            loadBackups(); // 刷新备份列表
            
            // 延迟重启服务器
            setTimeout(() => {
                restartServer();
            }, 1000);
        } else {
            if (result.errors) {
                showNotification('配置验证失败:\n' + result.errors.join('\n'), 'error');
            } else {
                showNotification('保存配置失败: ' + result.message, 'error');
            }
        }
    } catch (error) {
        showNotification('保存配置时发生错误: ' + error.message, 'error');
    }
    setLoading(false);
}

// 验证配置
async function validateConfig() {
    if (isLoading) return;
    
    setLoading(true);
    try {
        const config = collectFormData();
        
        const response = await fetch('/api/v1/config/validate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(config)
        });
        
        const result = await response.json();
        
        if (result.success) {
            showNotification('✅ 配置验证通过！所有设置正确。', 'success');
        } else {
            showNotification('❌ 配置验证失败:\n' + result.errors.join('\n'), 'error');
        }
    } catch (error) {
        showNotification('验证配置时发生错误: ' + error.message, 'error');
    }
    setLoading(false);
}

// 静默验证配置（不显示通知）
async function validateConfigSilently() {
    try {
        const config = collectFormData();
        
        const response = await fetch('/api/v1/config/validate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(config)
        });
        
        const result = await response.json();
        return result.success;
    } catch (error) {
        console.error('配置验证错误:', error);
        return false;
    }
}

// 填充表单数据
function populateForm(config) {
    // 服务器配置
    setFieldValue('server-host', config.server.host);
    setFieldValue('server-port', config.server.port);
    setFieldValue('server-mode', config.server.mode);
    setFieldValue('server-read-timeout', config.server.read_timeout);
    setFieldValue('server-write-timeout', config.server.write_timeout);
    setFieldValue('server-enable-cors', config.server.enable_cors);
    setFieldValue('server-enable-swagger', config.server.enable_swagger);
    
    // 数据库配置
    setFieldValue('database-host', config.database.host);
    setFieldValue('database-port', config.database.port);
    setFieldValue('database-user', config.database.user);
    setFieldValue('database-password', config.database.password);
    setFieldValue('database-dbname', config.database.dbname);
    setFieldValue('database-sslmode', config.database.sslmode);
    setFieldValue('database-max-open-conns', config.database.max_open_conns);
    setFieldValue('database-max-idle-conns', config.database.max_idle_conns);
    setFieldValue('database-max-lifetime', config.database.max_lifetime);
    
    // Redis配置
    setFieldValue('redis-host', config.redis.host);
    setFieldValue('redis-port', config.redis.port);
    setFieldValue('redis-password', config.redis.password);
    setFieldValue('redis-db', config.redis.db);
    setFieldValue('redis-pool-size', config.redis.pool_size);
    setFieldValue('redis-min-idle-conns', config.redis.min_idle_conns);
    
    // Telegram配置
    setFieldValue('telegram-enabled', config.bot.enabled);
    setFieldValue('telegram-token', config.bot.token);
    setFieldValue('telegram-webhook-url', config.bot.webhook_url);
    populateArrayField('telegram-admin-ids', config.bot.admin_ids || [], 'addAdminId');
    
    // 爬虫配置
    populateArrayField('crawler-user-agents', config.crawler.user_agents || [], 'addUserAgent');
    setFieldValue('crawler-proxy-enabled', config.crawler.proxy_enabled);
    populateArrayField('crawler-proxy-list', config.crawler.proxy_list || [], 'addProxy');
    setFieldValue('crawler-request-delay', config.crawler.request_delay);
    setFieldValue('crawler-retry-count', config.crawler.retry_count);
    setFieldValue('crawler-timeout', config.crawler.timeout);
    setFieldValue('crawler-concurrent-max', config.crawler.concurrent_max);
    
    // 媒体库配置
    setFieldValue('media-base-path', config.media.base_path);
    setFieldValue('media-scan-interval', config.media.scan_interval);
    populateArrayField('media-supported-exts', config.media.supported_exts || [], 'addMediaExt');
    setFieldValue('media-min-file-size', config.media.min_file_size);
    setFieldValue('media-max-file-size', config.media.max_file_size);
    
    // 安全配置
    setFieldValue('security-jwt-secret', config.security.jwt_secret);
    setFieldValue('security-jwt-expiry', config.security.jwt_expiry);
    setFieldValue('security-password-salt', config.security.password_salt);
    setFieldValue('security-rate-limit-rps', config.security.rate_limit_rps);
    populateArrayField('security-allowed-ips', config.security.allowed_ips || [], 'addAllowedIP');
    setFieldValue('security-enable-auth', config.security.enable_auth);
    
    // 通知配置
    setFieldValue('notification-telegram-enabled', config.notifications.telegram.enabled);
    setFieldValue('notification-telegram-chat-id', config.notifications.telegram.chat_id);
    setFieldValue('notification-email-enabled', config.notifications.email.enabled);
    setFieldValue('notification-email-smtp-host', config.notifications.email.smtp_host);
    setFieldValue('notification-email-smtp-port', config.notifications.email.smtp_port);
    setFieldValue('notification-email-username', config.notifications.email.username);
    setFieldValue('notification-email-password', config.notifications.email.password);
    setFieldValue('notification-email-from', config.notifications.email.from);
    populateArrayField('notification-email-to', config.notifications.email.to || [], 'addEmailTo');
    
    // 日志配置
    setFieldValue('log-level', config.log.level);
    setFieldValue('log-format', config.log.format);
    setFieldValue('log-output', config.log.output);
    setFieldValue('log-filename', config.log.filename);
    setFieldValue('log-max-size', config.log.max_size);
    setFieldValue('log-max-backups', config.log.max_backups);
    setFieldValue('log-max-age', config.log.max_age);
    setFieldValue('log-compress', config.log.compress);
    
    // 开发环境配置
    setFieldValue('dev-enable-debug-routes', config.dev.enable_debug_routes);
    setFieldValue('dev-enable-profiling', config.dev.enable_profiling);
    setFieldValue('dev-auto-reload', config.dev.auto_reload);
    
    // 种子下载配置
    if (config.torrent) {
        setFieldValue('torrent-jackett-host', config.torrent.jackett.host);
        setFieldValue('torrent-jackett-api-key', config.torrent.jackett.api_key);
        setFieldValue('torrent-jackett-timeout', config.torrent.jackett.timeout);
        setFieldValue('torrent-jackett-retry-count', config.torrent.jackett.retry_count);
        
        setFieldValue('torrent-qbittorrent-host', config.torrent.qbittorrent.host);
        setFieldValue('torrent-qbittorrent-username', config.torrent.qbittorrent.username);
        setFieldValue('torrent-qbittorrent-password', config.torrent.qbittorrent.password);
        setFieldValue('torrent-qbittorrent-timeout', config.torrent.qbittorrent.timeout);
        setFieldValue('torrent-qbittorrent-retry-count', config.torrent.qbittorrent.retry_count);
        setFieldValue('torrent-qbittorrent-download-dir', config.torrent.qbittorrent.download_dir);
        
        setFieldValue('torrent-search-max-results', config.torrent.search.max_results);
        setFieldValue('torrent-search-min-seeders', config.torrent.search.min_seeders);
        setFieldValue('torrent-search-sort-by-size', config.torrent.search.sort_by_size);
    }
}

// 收集表单数据
function collectFormData() {
    const formData = {
        server: {
            host: getFieldValue('server-host'),
            port: parseInt(getFieldValue('server-port')) || 8080,
            mode: getFieldValue('server-mode'),
            read_timeout: getFieldValue('server-read-timeout'),
            write_timeout: getFieldValue('server-write-timeout'),
            enable_cors: getFieldValue('server-enable-cors'),
            enable_swagger: getFieldValue('server-enable-swagger')
        },
        database: {
            host: getFieldValue('database-host'),
            port: parseInt(getFieldValue('database-port')) || 5432,
            user: getFieldValue('database-user'),
            password: getFieldValue('database-password'),
            dbname: getFieldValue('database-dbname'),
            sslmode: getFieldValue('database-sslmode'),
            max_open_conns: parseInt(getFieldValue('database-max-open-conns')) || 25,
            max_idle_conns: parseInt(getFieldValue('database-max-idle-conns')) || 10,
            max_lifetime: parseInt(getFieldValue('database-max-lifetime')) || 3600
        },
        redis: {
            host: getFieldValue('redis-host'),
            port: parseInt(getFieldValue('redis-port')) || 6379,
            password: getFieldValue('redis-password'),
            db: parseInt(getFieldValue('redis-db')) || 0,
            pool_size: parseInt(getFieldValue('redis-pool-size')) || 10,
            min_idle_conns: parseInt(getFieldValue('redis-min-idle-conns')) || 5
        },
        bot: {
            enabled: getFieldValue('telegram-enabled'),
            token: getFieldValue('telegram-token'),
            webhook_url: getFieldValue('telegram-webhook-url'),
            admin_ids: collectArrayField('telegram-admin-ids').map(id => parseInt(id)).filter(id => !isNaN(id))
        },
        crawler: {
            user_agents: collectArrayField('crawler-user-agents'),
            proxy_enabled: getFieldValue('crawler-proxy-enabled'),
            proxy_list: collectArrayField('crawler-proxy-list'),
            request_delay: getFieldValue('crawler-request-delay'),
            retry_count: parseInt(getFieldValue('crawler-retry-count')) || 3,
            timeout: getFieldValue('crawler-timeout'),
            concurrent_max: parseInt(getFieldValue('crawler-concurrent-max')) || 5
        },
        media: {
            base_path: getFieldValue('media-base-path'),
            scan_interval: parseInt(getFieldValue('media-scan-interval')) || 24,
            supported_exts: collectArrayField('media-supported-exts'),
            min_file_size: parseInt(getFieldValue('media-min-file-size')) || 100,
            max_file_size: parseInt(getFieldValue('media-max-file-size')) || 10240
        },
        security: {
            jwt_secret: getFieldValue('security-jwt-secret'),
            jwt_expiry: getFieldValue('security-jwt-expiry'),
            password_salt: getFieldValue('security-password-salt'),
            rate_limit_rps: parseInt(getFieldValue('security-rate-limit-rps')) || 100,
            allowed_ips: collectArrayField('security-allowed-ips'),
            enable_auth: getFieldValue('security-enable-auth')
        },
        notifications: {
            telegram: {
                enabled: getFieldValue('notification-telegram-enabled'),
                chat_id: getFieldValue('notification-telegram-chat-id')
            },
            email: {
                enabled: getFieldValue('notification-email-enabled'),
                smtp_host: getFieldValue('notification-email-smtp-host'),
                smtp_port: parseInt(getFieldValue('notification-email-smtp-port')) || 587,
                username: getFieldValue('notification-email-username'),
                password: getFieldValue('notification-email-password'),
                from: getFieldValue('notification-email-from'),
                to: collectArrayField('notification-email-to')
            }
        },
        log: {
            level: getFieldValue('log-level'),
            format: getFieldValue('log-format'),
            output: getFieldValue('log-output'),
            filename: getFieldValue('log-filename'),
            max_size: parseInt(getFieldValue('log-max-size')) || 100,
            max_backups: parseInt(getFieldValue('log-max-backups')) || 7,
            max_age: parseInt(getFieldValue('log-max-age')) || 30,
            compress: getFieldValue('log-compress')
        },
        dev: {
            enable_debug_routes: getFieldValue('dev-enable-debug-routes'),
            enable_profiling: getFieldValue('dev-enable-profiling'),
            auto_reload: getFieldValue('dev-auto-reload')
        },
        sites: {
            javdb: {
                base_url: "https://javdb.com",
                search_path: "/search",
                detail_selector: ".movie-panel",
                rate_limit: "1s"
            },
            javlibrary: {
                base_url: "http://www.javlibrary.com",
                search_path: "/cn/vl_searchbyid.php",
                rate_limit: "2s"
            },
            javbus: {
                base_url: "https://www.javbus.com",
                search_path: "/search",
                rate_limit: "1s"
            }
        },
        download: {
            max_concurrent: 3,
            retry_count: 3,
            retry_delay: "5m",
            speed_limit: 0,
            temp_dir: "/tmp/nsfw-downloads",
            completed_dir: "/MediaCenter/NSFW/Hub/#Downloads"
        },
        torrent: {
            jackett: {
                host: getFieldValue('torrent-jackett-host'),
                api_key: getFieldValue('torrent-jackett-api-key'),
                timeout: getFieldValue('torrent-jackett-timeout'),
                retry_count: parseInt(getFieldValue('torrent-jackett-retry-count')) || 3
            },
            qbittorrent: {
                host: getFieldValue('torrent-qbittorrent-host'),
                username: getFieldValue('torrent-qbittorrent-username'),
                password: getFieldValue('torrent-qbittorrent-password'),
                timeout: getFieldValue('torrent-qbittorrent-timeout'),
                retry_count: parseInt(getFieldValue('torrent-qbittorrent-retry-count')) || 3,
                download_dir: getFieldValue('torrent-qbittorrent-download-dir')
            },
            search: {
                max_results: parseInt(getFieldValue('torrent-search-max-results')) || 20,
                min_seeders: parseInt(getFieldValue('torrent-search-min-seeders')) || 1,
                sort_by_size: getFieldValue('torrent-search-sort-by-size')
            }
        }
    };
    
    return formData;
}

// 设置字段值
function setFieldValue(fieldId, value) {
    const field = document.getElementById(fieldId);
    if (!field) return;
    
    if (field.type === 'checkbox') {
        field.checked = !!value;
    } else {
        field.value = value || '';
    }
}

// 获取字段值
function getFieldValue(fieldId) {
    const field = document.getElementById(fieldId);
    if (!field) return '';
    
    if (field.type === 'checkbox') {
        return field.checked;
    }
    return field.value;
}

// 获取字段值（现在直接返回当前值，不再处理占位符）
function getFieldValueWithFallback(fieldId, originalValue) {
    const currentValue = getFieldValue(fieldId);
    
    // 如果当前值为空，则使用原始值
    if (currentValue === '') {
        return originalValue || '';
    }
    
    return currentValue;
}

// 填充数组字段
function populateArrayField(containerId, values, addFunctionName) {
    const container = document.getElementById(containerId);
    if (!container) return;
    
    container.innerHTML = '';
    values.forEach(value => {
        addArrayItem(container, value, addFunctionName);
    });
}

// 收集数组字段
function collectArrayField(containerId) {
    const container = document.getElementById(containerId);
    if (!container) return [];
    
    const inputs = container.querySelectorAll('input');
    return Array.from(inputs).map(input => input.value).filter(value => value.trim() !== '');
}

// 添加数组项
function addArrayItem(container, value = '', addFunctionName) {
    const div = document.createElement('div');
    div.className = 'array-item';
    
    const input = document.createElement('input');
    input.type = 'text';
    input.className = 'form-input';
    input.value = value;
    
    const removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.className = 'remove-btn';
    removeBtn.textContent = '删除';
    removeBtn.onclick = () => div.remove();
    
    div.appendChild(input);
    div.appendChild(removeBtn);
    container.appendChild(div);
}

// 数组字段添加函数
function addAdminId() {
    addArrayItem(document.getElementById('telegram-admin-ids'));
}

function addUserAgent() {
    addArrayItem(document.getElementById('crawler-user-agents'));
}

function addProxy() {
    addArrayItem(document.getElementById('crawler-proxy-list'));
}

function addMediaExt() {
    addArrayItem(document.getElementById('media-supported-exts'));
}

function addAllowedIP() {
    addArrayItem(document.getElementById('security-allowed-ips'));
}

function addEmailTo() {
    addArrayItem(document.getElementById('notification-email-to'));
}

// 连接测试函数
async function testDatabaseConnection() {
    const host = getFieldValue('database-host');
    const port = getFieldValue('database-port');
    const user = getFieldValue('database-user');
    const password = getFieldValue('database-password');
    const dbname = getFieldValue('database-dbname');
    const sslmode = getFieldValue('database-sslmode');
    
    // 验证端口号
    const portNum = parseInt(port);
    if (isNaN(portNum) || portNum <= 0 || portNum > 65535) {
        const resultElement = document.getElementById('database-test-result');
        resultElement.style.display = 'block';
        resultElement.className = 'test-result error';
        resultElement.textContent = '✗ 端口号无效，应在1-65535之间';
        setTimeout(() => {
            resultElement.style.display = 'none';
        }, 5000);
        return;
    }
    
    await testConnection('database', {
        host: host,
        port: portNum,
        user: user,
        password: password,
        dbname: dbname,
        sslmode: sslmode,
        max_open_conns: parseInt(getFieldValue('database-max-open-conns')) || 25,
        max_idle_conns: parseInt(getFieldValue('database-max-idle-conns')) || 10,
        max_lifetime: parseInt(getFieldValue('database-max-lifetime')) || 3600
    }, 'database-test-result');
}

async function testRedisConnection() {
    const host = getFieldValue('redis-host');
    const port = getFieldValue('redis-port');
    const password = getFieldValue('redis-password');
    const db = getFieldValue('redis-db');
    
    // 验证端口号
    const portNum = parseInt(port);
    if (isNaN(portNum) || portNum <= 0 || portNum > 65535) {
        const resultElement = document.getElementById('redis-test-result');
        resultElement.style.display = 'block';
        resultElement.className = 'test-result error';
        resultElement.textContent = '✗ 端口号无效，应在1-65535之间';
        setTimeout(() => {
            resultElement.style.display = 'none';
        }, 5000);
        return;
    }
    
    await testConnection('redis', {
        host: host,
        port: portNum,
        password: password,
        db: parseInt(db) || 0,
        pool_size: parseInt(getFieldValue('redis-pool-size')) || 10,
        min_idle_conns: parseInt(getFieldValue('redis-min-idle-conns')) || 5
    }, 'redis-test-result');
}

async function testTelegramConnection() {
    const token = getFieldValue('telegram-token');
    
    await testConnection('telegram', {
        enabled: getFieldValue('telegram-enabled'),
        token: token,
        webhook_url: getFieldValue('telegram-webhook-url'),
        admin_ids: collectArrayField('telegram-admin-ids').map(id => parseInt(id)).filter(id => !isNaN(id))
    }, 'telegram-test-result');
}

async function testEmailConnection() {
    const smtpPort = getFieldValue('notification-email-smtp-port');
    const password = getFieldValue('notification-email-password');
    
    // 验证端口号
    const portNum = parseInt(smtpPort);
    if (isNaN(portNum) || portNum <= 0 || portNum > 65535) {
        const resultElement = document.getElementById('email-test-result');
        resultElement.style.display = 'block';
        resultElement.className = 'test-result error';
        resultElement.textContent = '✗ SMTP端口号无效，应在1-65535之间';
        setTimeout(() => {
            resultElement.style.display = 'none';
        }, 5000);
        return;
    }
    
    await testConnection('email', {
        enabled: getFieldValue('notification-email-enabled'),
        smtp_host: getFieldValue('notification-email-smtp-host'),
        smtp_port: portNum,
        username: getFieldValue('notification-email-username'),
        password: password,
        from: getFieldValue('notification-email-from'),
        to: collectArrayField('notification-email-to')
    }, 'email-test-result');
}

async function testJackettConnection() {
    const apiKey = getFieldValue('torrent-jackett-api-key');
    
    await testConnection('jackett', {
        host: getFieldValue('torrent-jackett-host'),
        api_key: apiKey,
        timeout: getFieldValue('torrent-jackett-timeout'),
        retry_count: parseInt(getFieldValue('torrent-jackett-retry-count')) || 3
    }, 'jackett-test-result');
}

async function testQBittorrentConnection() {
    const password = getFieldValue('torrent-qbittorrent-password');
    
    await testConnection('qbittorrent', {
        host: getFieldValue('torrent-qbittorrent-host'),
        username: getFieldValue('torrent-qbittorrent-username'),
        password: password,
        timeout: getFieldValue('torrent-qbittorrent-timeout'),
        retry_count: parseInt(getFieldValue('torrent-qbittorrent-retry-count')) || 3,
        download_dir: getFieldValue('torrent-qbittorrent-download-dir')
    }, 'qbittorrent-test-result');
}

// 测试连接的通用函数
async function testConnection(type, data, resultElementId) {
    const resultElement = document.getElementById(resultElementId);
    resultElement.style.display = 'block';
    resultElement.className = 'test-result';
    resultElement.textContent = '测试中...';
    
    try {
        const response = await fetch('/api/v1/config/test', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                type: type,
                data: data
            })
        });
        
        const result = await response.json();
        
        // 检查外层API调用是否成功
        if (result.success && result.data) {
            // 检查实际的连接测试结果
            if (result.data.success) {
                resultElement.className = 'test-result success';
                resultElement.textContent = '✓ ' + (result.data.message || '连接成功') + 
                    (result.data.latency ? ` (${result.data.latency}ms)` : '');
            } else {
                resultElement.className = 'test-result error';
                resultElement.textContent = '✗ 连接失败: ' + (result.data.message || '未知错误') +
                    (result.data.latency ? ` (${result.data.latency}ms)` : '');
            }
        } else {
            resultElement.className = 'test-result error';
            resultElement.textContent = '✗ 测试失败: ' + (result.message || '未知错误');
        }
    } catch (error) {
        resultElement.className = 'test-result error';
        resultElement.textContent = '✗ 测试失败: ' + error.message;
    }
    
    // 5秒后隐藏结果
    setTimeout(() => {
        resultElement.style.display = 'none';
    }, 5000);
}

// 加载备份列表
async function loadBackups() {
    try {
        const response = await fetch('/api/v1/config/backups');
        const result = await response.json();
        
        if (result.success) {
            populateBackupList(result.data || []);
        }
    } catch (error) {
        console.error('加载备份列表失败:', error);
    }
}

// 填充备份列表
function populateBackupList(backups) {
    const container = document.getElementById('backup-list');
    if (!container) return;
    
    container.innerHTML = '';
    
    if (backups.length === 0) {
        container.innerHTML = '<div style="color: #fff; text-align: center; padding: 20px;">暂无备份文件</div>';
        return;
    }
    
    backups.forEach(backup => {
        const div = document.createElement('div');
        div.className = 'backup-item';
        
        const nameSpan = document.createElement('span');
        nameSpan.className = 'backup-name';
        nameSpan.textContent = backup;
        
        const restoreBtn = document.createElement('button');
        restoreBtn.className = 'restore-btn';
        restoreBtn.textContent = '恢复';
        restoreBtn.onclick = () => restoreBackup(backup);
        
        div.appendChild(nameSpan);
        div.appendChild(restoreBtn);
        container.appendChild(div);
    });
}

// 恢复备份
async function restoreBackup(backupName) {
    if (!confirm(`确定要恢复备份 "${backupName}" 吗？这将覆盖当前配置。`)) {
        return;
    }
    
    try {
        const response = await fetch(`/api/v1/config/restore/${backupName}`, {
            method: 'POST'
        });
        
        const result = await response.json();
        
        if (result.success) {
            showNotification('备份恢复成功，正在重新加载配置...', 'success');
            setTimeout(() => {
                loadConfig();
            }, 1000);
        } else {
            showNotification('恢复备份失败: ' + result.message, 'error');
        }
    } catch (error) {
        showNotification('恢复备份时发生错误: ' + error.message, 'error');
    }
}

// 设置加载状态
function setLoading(loading) {
    isLoading = loading;
    const container = document.querySelector('.config-container');
    if (container) {
        if (loading) {
            container.classList.add('loading');
        } else {
            container.classList.remove('loading');
        }
    }
}

// 显示通知
function showNotification(message, type) {
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.classList.add('show');
    }, 100);
    
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    }, 3000);
}

// 返回上一页
function goBack() {
    window.history.back();
}

// 重启服务器
async function restartServer() {
    try {
        showNotification('正在重启服务器，请稍等...', 'success');
        
        const response = await fetch('/api/v1/system/restart', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (response.ok) {
            showNotification('服务器重启指令已发送', 'success');
            
            // 等待服务器重启完成
            setTimeout(() => {
                showNotification('服务器正在重启中，页面将在5秒后自动刷新...', 'success');
                setTimeout(() => {
                    window.location.reload();
                }, 5000);
            }, 2000);
        } else {
            showNotification('重启服务器失败', 'error');
        }
    } catch (error) {
        showNotification('重启服务器时发生错误: ' + error.message, 'error');
    }
} 