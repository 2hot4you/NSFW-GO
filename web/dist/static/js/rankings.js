// 排行榜页面JavaScript
(function() {
    'use strict';

    // 初始化页面
    function init() {
        loadRankings();
        setupEventListeners();
        setupAutoRefresh();
    }

    // 设置事件监听器
    function setupEventListeners() {
        // 刷新按钮
        const refreshBtn = document.getElementById('refreshRankings');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                loadRankings(true);
            });
        }

        // 爬取按钮
        const crawlBtn = document.getElementById('crawlRankings');
        if (crawlBtn) {
            crawlBtn.addEventListener('click', crawlRankings);
        }

        // 排行榜类型切换
        const rankTypeButtons = document.querySelectorAll('[data-rank-type]');
        rankTypeButtons.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const rankType = e.target.dataset.rankType;
                rankTypeButtons.forEach(b => b.classList.remove('active'));
                e.target.classList.add('active');
                loadRankings(false, rankType);
            });
        });
    }

    // 设置自动刷新
    function setupAutoRefresh() {
        // 每5分钟自动刷新一次
        setInterval(() => {
            loadRankings(false);
        }, 5 * 60 * 1000);
    }

    // 加载排行榜数据
    async function loadRankings(showLoading = true, rankType = 'daily') {
        if (showLoading) {
            showLoadingState();
        }

        try {
            const response = await fetch(`/api/v1/rankings?rank_type=${rankType}`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const result = await response.json();
            
            if (result.success && result.data) {
                displayRankings(result.data, rankType);
                updateStats(result);
            } else {
                showError('暂无排行榜数据');
            }
        } catch (error) {
            console.error('加载排行榜失败:', error);
            showError('加载排行榜失败: ' + error.message);
        }
    }

    // 显示排行榜数据
    function displayRankings(rankings, rankType) {
        const container = document.getElementById('rankingsGrid');
        if (!container) {
            console.error('找不到rankingsGrid容器');
            return;
        }

        if (!rankings || rankings.length === 0) {
            container.innerHTML = `
                <div class="col-span-full text-center py-12">
                    <i class="fas fa-chart-bar text-6xl text-gray-600 mb-4"></i>
                    <p class="text-gray-500">暂无${getRankTypeText(rankType)}排行榜数据</p>
                    <button onclick="crawlRankings()" class="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
                        立即爬取
                    </button>
                </div>
            `;
            return;
        }

        // 生成排行榜卡片
        const cards = rankings.map(item => createRankingCard(item)).join('');
        container.innerHTML = cards;

        // 添加卡片点击事件
        setupCardClickEvents();
    }

    // 创建排行榜卡片
    function createRankingCard(item) {
        const positionBadge = getPositionBadge(item.position);
        const localBadge = item.local_exists 
            ? '<span class="absolute top-2 right-2 bg-green-500 text-white text-xs px-2 py-1 rounded z-10">本地已存在</span>'
            : '';

        return `
            <div class="ranking-card bg-gray-800 rounded-lg overflow-hidden hover:shadow-xl transition-all duration-300 cursor-pointer" data-code="${item.code}">
                <div class="relative" style="aspect-ratio: 2/3;">
                    ${positionBadge}
                    ${localBadge}
                    <div class="absolute inset-0 bg-gray-700">
                        <img src="${item.cover_url || '/static/img/no-cover.jpg'}" 
                             alt="${item.title}" 
                             class="w-full h-full object-cover object-center"
                             loading="lazy"
                             onerror="this.src='/static/img/no-cover.jpg'">
                    </div>
                </div>
                <div class="p-4">
                    <div class="flex items-center justify-between mb-2">
                        <span class="text-blue-400 font-mono text-sm">${item.code}</span>
                        <span class="text-gray-500 text-xs">
                            <i class="far fa-clock"></i>
                            ${formatTime(item.crawled_at)}
                        </span>
                    </div>
                    <h3 class="text-white text-sm font-medium line-clamp-2" title="${item.title}">
                        ${item.title}
                    </h3>
                    <div class="mt-3 flex justify-between items-center">
                        <button class="search-btn text-blue-400 hover:text-blue-300 text-sm" data-code="${item.code}">
                            <i class="fas fa-search"></i> 搜索种子
                        </button>
                        ${!item.local_exists ? `
                            <button class="download-btn text-green-400 hover:text-green-300 text-sm" data-code="${item.code}">
                                <i class="fas fa-download"></i> 下载
                            </button>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;
    }

    // 获取排名徽章
    function getPositionBadge(position) {
        let badgeClass = 'absolute top-2 left-2 text-white text-lg font-bold px-3 py-1 rounded-full ';
        let icon = '';

        if (position === 1) {
            badgeClass += 'bg-gradient-to-r from-yellow-400 to-yellow-600';
            icon = '🥇';
        } else if (position === 2) {
            badgeClass += 'bg-gradient-to-r from-gray-300 to-gray-500';
            icon = '🥈';
        } else if (position === 3) {
            badgeClass += 'bg-gradient-to-r from-orange-400 to-orange-600';
            icon = '🥉';
        } else if (position <= 10) {
            badgeClass += 'bg-gradient-to-r from-purple-500 to-purple-700';
            icon = position;
        } else {
            badgeClass += 'bg-gray-700';
            icon = position;
        }

        return `<span class="${badgeClass}">${icon}</span>`;
    }

    // 设置卡片点击事件
    function setupCardClickEvents() {
        // 搜索按钮
        document.querySelectorAll('.search-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const code = e.currentTarget.dataset.code;
                searchTorrents(code);
            });
        });

        // 下载按钮
        document.querySelectorAll('.download-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const code = e.currentTarget.dataset.code;
                window.location.href = `/downloads.html?search=${code}`;
            });
        });

        // 卡片点击查看详情
        document.querySelectorAll('.ranking-card').forEach(card => {
            card.addEventListener('click', (e) => {
                if (!e.target.closest('.search-btn') && !e.target.closest('.download-btn')) {
                    const code = card.dataset.code;
                    viewDetails(code);
                }
            });
        });
    }

    // 搜索种子
    async function searchTorrents(code) {
        window.location.href = `/downloads.html?search=${code}`;
    }

    // 查看详情
    async function viewDetails(code) {
        try {
            const response = await fetch(`/api/v1/search/javdb?keyword=${code}`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const result = await response.json();
            if (result.success && result.data && result.data.length > 0) {
                showDetailsModal(result.data[0]);
            }
        } catch (error) {
            console.error('获取详情失败:', error);
            showNotification('获取详情失败', 'error');
        }
    }

    // 显示详情模态框
    function showDetailsModal(movie) {
        // 这里可以实现一个详情模态框
        console.log('显示影片详情:', movie);
        // TODO: 实现详情模态框
    }

    // 手动爬取排行榜
    async function crawlRankings() {
        const btn = document.getElementById('crawlRankings');
        if (btn) {
            btn.disabled = true;
            btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 爬取中...';
        }

        try {
            const response = await fetch('/api/v1/rankings/crawl', {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const result = await response.json();
            
            if (result.success) {
                showNotification('排行榜爬取成功！', 'success');
                // 延迟1秒后刷新
                setTimeout(() => {
                    loadRankings(true);
                }, 1000);
            } else {
                showNotification(result.error || '爬取失败', 'error');
            }
        } catch (error) {
            console.error('爬取失败:', error);
            showNotification('爬取失败: ' + error.message, 'error');
        } finally {
            if (btn) {
                btn.disabled = false;
                btn.innerHTML = '<i class="fas fa-sync"></i> 手动爬取';
            }
        }
    }

    // 更新统计信息
    function updateStats(result) {
        const statsContainer = document.getElementById('rankingsStats');
        if (!statsContainer) return;

        const stats = `
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
                <div class="bg-gray-800 p-4 rounded-lg">
                    <div class="text-gray-400 text-sm">总数量</div>
                    <div class="text-2xl font-bold text-white">${result.count || 0}</div>
                </div>
                <div class="bg-gray-800 p-4 rounded-lg">
                    <div class="text-gray-400 text-sm">本地已存在</div>
                    <div class="text-2xl font-bold text-green-400">
                        ${result.data ? result.data.filter(item => item.local_exists).length : 0}
                    </div>
                </div>
                <div class="bg-gray-800 p-4 rounded-lg">
                    <div class="text-gray-400 text-sm">最后更新</div>
                    <div class="text-sm text-white">
                        ${result.data && result.data[0] ? formatTime(result.data[0].crawled_at) : '未知'}
                    </div>
                </div>
            </div>
        `;
        statsContainer.innerHTML = stats;
    }

    // 显示加载状态
    function showLoadingState() {
        const container = document.getElementById('rankingsGrid');
        if (!container) return;

        container.innerHTML = `
            <div class="col-span-full flex justify-center items-center py-12">
                <div class="text-center">
                    <i class="fas fa-spinner fa-spin text-4xl text-blue-500 mb-4"></i>
                    <p class="text-gray-400">加载排行榜中...</p>
                </div>
            </div>
        `;
    }

    // 显示错误信息
    function showError(message) {
        const container = document.getElementById('rankingsGrid');
        if (!container) return;

        container.innerHTML = `
            <div class="col-span-full text-center py-12">
                <i class="fas fa-exclamation-triangle text-6xl text-red-500 mb-4"></i>
                <p class="text-gray-400">${message}</p>
                <button onclick="location.reload()" class="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
                    重试
                </button>
            </div>
        `;
    }

    // 显示通知
    function showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `fixed top-4 right-4 p-4 rounded-lg shadow-lg z-50 ${
            type === 'success' ? 'bg-green-500' : 
            type === 'error' ? 'bg-red-500' : 
            'bg-blue-500'
        } text-white`;
        notification.innerHTML = message;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 3000);
    }

    // 格式化时间
    function formatTime(timestamp) {
        if (!timestamp) return '未知';
        
        const date = new Date(timestamp);
        const now = new Date();
        const diff = now - date;
        
        if (diff < 60000) {
            return '刚刚';
        } else if (diff < 3600000) {
            return Math.floor(diff / 60000) + '分钟前';
        } else if (diff < 86400000) {
            return Math.floor(diff / 3600000) + '小时前';
        } else {
            return date.toLocaleDateString('zh-CN', {
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit'
            });
        }
    }

    // 获取排行榜类型文本
    function getRankTypeText(rankType) {
        const types = {
            'daily': '日榜',
            'weekly': '周榜',
            'monthly': '月榜'
        };
        return types[rankType] || '日榜';
    }

    // 页面加载完成后初始化
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    // 导出全局函数
    window.crawlRankings = crawlRankings;
})();