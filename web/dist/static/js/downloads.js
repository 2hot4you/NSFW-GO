// 下载页面JavaScript
(function() {
    'use strict';

    let currentSearchQuery = '';
    let currentPage = 1;
    let isSearching = false;

    // 初始化页面
    function init() {
        setupEventListeners();
        checkUrlParams();
        loadQBittorrentStatus();
        setupAutoRefresh();
    }

    // 设置事件监听器
    function setupEventListeners() {
        // 搜索表单
        const searchForm = document.getElementById('torrentSearchForm');
        if (searchForm) {
            searchForm.addEventListener('submit', (e) => {
                e.preventDefault();
                performSearch();
            });
        }

        // 搜索输入框
        const searchInput = document.getElementById('searchInput');
        if (searchInput) {
            // 回车搜索
            searchInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    performSearch();
                }
            });
        }

        // 搜索按钮
        const searchBtn = document.getElementById('searchBtn');
        if (searchBtn) {
            searchBtn.addEventListener('click', performSearch);
        }

        // 刷新下载列表
        const refreshBtn = document.getElementById('refreshDownloads');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', loadQBittorrentStatus);
        }

        // 清空下载列表
        const clearBtn = document.getElementById('clearCompleted');
        if (clearBtn) {
            clearBtn.addEventListener('click', clearCompletedDownloads);
        }
    }

    // 检查URL参数
    function checkUrlParams() {
        const params = new URLSearchParams(window.location.search);
        const searchQuery = params.get('search');
        
        if (searchQuery) {
            const searchInput = document.getElementById('searchInput');
            if (searchInput) {
                searchInput.value = searchQuery;
                performSearch();
            }
        }
    }

    // 设置自动刷新
    function setupAutoRefresh() {
        // 每10秒刷新一次下载状态
        setInterval(() => {
            loadQBittorrentStatus(false);
        }, 10000);
    }

    // 执行搜索
    async function performSearch() {
        const searchInput = document.getElementById('searchInput');
        if (!searchInput) return;

        const query = searchInput.value.trim();
        if (!query) {
            showNotification('请输入搜索关键词', 'warning');
            return;
        }

        if (isSearching) {
            showNotification('正在搜索中，请稍候...', 'info');
            return;
        }

        currentSearchQuery = query;
        currentPage = 1;
        isSearching = true;

        // 更新搜索按钮状态
        const searchBtn = document.getElementById('searchBtn');
        if (searchBtn) {
            searchBtn.disabled = true;
            searchBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 搜索中...';
        }

        // 显示加载状态
        showSearchLoading();

        try {
            const response = await fetch(`/api/v1/torrents/search?query=${encodeURIComponent(query)}&page=${currentPage}`);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const result = await response.json();
            
            if (result.success && result.data) {
                displaySearchResults(result.data);
            } else {
                showSearchError(result.error || '搜索失败');
            }
        } catch (error) {
            console.error('搜索失败:', error);
            showSearchError('搜索失败: ' + error.message);
        } finally {
            isSearching = false;
            if (searchBtn) {
                searchBtn.disabled = false;
                searchBtn.innerHTML = '<i class="fas fa-search"></i> 搜索';
            }
        }
    }

    // 显示搜索结果
    function displaySearchResults(torrents) {
        const container = document.getElementById('searchResults');
        if (!container) return;

        if (!torrents || torrents.length === 0) {
            container.innerHTML = `
                <div class="text-center py-12">
                    <i class="fas fa-inbox text-6xl text-gray-600 mb-4"></i>
                    <p class="text-gray-500">没有找到相关种子</p>
                    <p class="text-gray-600 text-sm mt-2">尝试使用不同的关键词</p>
                </div>
            `;
            return;
        }

        // 生成种子列表
        const torrentList = torrents.map(torrent => createTorrentItem(torrent)).join('');
        
        container.innerHTML = `
            <div class="mb-4 text-gray-400">
                找到 ${torrents.length} 个种子
            </div>
            <div class="space-y-4">
                ${torrentList}
            </div>
        `;

        // 添加下载事件
        setupDownloadEvents();
    }

    // 创建种子项
    function createTorrentItem(torrent) {
        const sizeText = formatFileSize(torrent.size);
        const seedersClass = torrent.seeders > 10 ? 'text-green-400' : 
                           torrent.seeders > 0 ? 'text-yellow-400' : 'text-red-400';
        
        return `
            <div class="torrent-item bg-gray-800 rounded-lg p-4 hover:bg-gray-750 transition-colors">
                <div class="flex items-start justify-between">
                    <div class="flex-1">
                        <h3 class="text-white font-medium mb-2" title="${torrent.title}">
                            ${torrent.title}
                        </h3>
                        <div class="flex flex-wrap gap-4 text-sm text-gray-400">
                            <span><i class="fas fa-hdd"></i> ${sizeText}</span>
                            <span class="${seedersClass}">
                                <i class="fas fa-seedling"></i> ${torrent.seeders || 0}
                            </span>
                            <span class="text-blue-400">
                                <i class="fas fa-users"></i> ${torrent.leechers || 0}
                            </span>
                            <span><i class="fas fa-server"></i> ${torrent.indexer || 'Unknown'}</span>
                            <span><i class="far fa-clock"></i> ${formatDate(torrent.publish_date)}</span>
                        </div>
                    </div>
                    <div class="ml-4 flex gap-2">
                        <button class="download-torrent px-3 py-2 bg-green-500 text-white rounded hover:bg-green-600 transition-colors" 
                                data-magnet="${torrent.magnet_link || ''}"
                                data-link="${torrent.link || ''}"
                                data-title="${torrent.title}">
                            <i class="fas fa-download"></i> 下载
                        </button>
                        ${torrent.magnet_link ? `
                            <button class="copy-magnet px-3 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
                                    data-magnet="${torrent.magnet_link}">
                                <i class="fas fa-magnet"></i> 复制
                            </button>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;
    }

    // 设置下载事件
    function setupDownloadEvents() {
        // 下载按钮
        document.querySelectorAll('.download-torrent').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const magnet = e.currentTarget.dataset.magnet;
                const link = e.currentTarget.dataset.link;
                const title = e.currentTarget.dataset.title;
                
                if (magnet || link) {
                    await downloadTorrent(magnet || link, title);
                } else {
                    showNotification('无效的种子链接', 'error');
                }
            });
        });

        // 复制磁力链接
        document.querySelectorAll('.copy-magnet').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const magnet = e.currentTarget.dataset.magnet;
                copyToClipboard(magnet);
            });
        });
    }

    // 下载种子
    async function downloadTorrent(url, title) {
        try {
            const response = await fetch('/api/v1/torrents/download', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    url: url,
                    title: title
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const result = await response.json();
            
            if (result.success) {
                showNotification('种子已添加到下载队列', 'success');
                // 刷新下载列表
                setTimeout(() => {
                    loadQBittorrentStatus();
                }, 1000);
            } else {
                showNotification(result.error || '添加下载失败', 'error');
            }
        } catch (error) {
            console.error('添加下载失败:', error);
            showNotification('添加下载失败: ' + error.message, 'error');
        }
    }

    // 加载qBittorrent状态
    async function loadQBittorrentStatus(showLoading = true) {
        const container = document.getElementById('downloadsList');
        if (!container) return;

        if (showLoading) {
            container.innerHTML = `
                <div class="text-center py-12">
                    <i class="fas fa-spinner fa-spin text-4xl text-blue-500 mb-4"></i>
                    <p class="text-gray-400">加载下载列表中...</p>
                </div>
            `;
        }

        try {
            const response = await fetch('/api/v1/torrents/status');
            
            if (!response.ok) {
                if (response.status === 404) {
                    showDownloadsError('qBittorrent 服务未配置或未运行');
                    return;
                }
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const result = await response.json();
            
            if (result.success && result.data) {
                displayDownloads(result.data);
                updateDownloadStats(result.data);
            } else {
                showDownloadsError('无法获取下载状态');
            }
        } catch (error) {
            console.error('获取下载状态失败:', error);
            showDownloadsError('获取下载状态失败');
        }
    }

    // 显示下载列表
    function displayDownloads(torrents) {
        const container = document.getElementById('downloadsList');
        if (!container) return;

        // 过滤只显示PornDB标签的种子
        const porndbTorrents = (torrents || []).filter(torrent => {
            const tags = torrent.tags || '';
            return tags.includes('PornDB');
        });

        if (porndbTorrents.length === 0) {
            container.innerHTML = `
                <div class="text-center py-12">
                    <i class="fas fa-download text-6xl text-gray-600 mb-4"></i>
                    <p class="text-gray-500">暂无 PornDB 下载任务</p>
                </div>
            `;
            return;
        }

        const downloadItems = porndbTorrents.map(torrent => createDownloadItem(torrent)).join('');
        container.innerHTML = `
            <div class="space-y-4">
                ${downloadItems}
            </div>
        `;

        // 设置控制按钮事件
        setupDownloadControls();
    }

    // 创建下载项
    function createDownloadItem(torrent) {
        const progress = Math.round(torrent.progress * 100);
        const statusInfo = getStatusInfo(torrent.state);
        const speed = torrent.dlspeed > 0 ? formatFileSize(torrent.dlspeed) + '/s' : '-';
        const eta = torrent.eta > 0 && torrent.eta < 8640000 ? formatTime(torrent.eta) : '-';
        
        return `
            <div class="download-item bg-gray-800 rounded-lg p-4" data-hash="${torrent.hash}">
                <div class="flex justify-between items-start mb-3">
                    <div class="flex-1">
                        <h4 class="text-white font-medium mb-1">${torrent.name}</h4>
                        <div class="flex gap-4 text-sm text-gray-400">
                            <span>${formatFileSize(torrent.size)}</span>
                            <span class="${statusInfo.class}">${statusInfo.text}</span>
                            ${torrent.dlspeed > 0 ? `<span class="text-green-400"><i class="fas fa-download"></i> ${speed}</span>` : ''}
                            ${torrent.eta > 0 ? `<span><i class="far fa-clock"></i> ${eta}</span>` : ''}
                        </div>
                    </div>
                    <div class="flex gap-2">
                        ${torrent.state === 'pausedDL' ? `
                            <button class="resume-btn px-2 py-1 bg-green-500 text-white rounded hover:bg-green-600" data-hash="${torrent.hash}">
                                <i class="fas fa-play"></i>
                            </button>
                        ` : torrent.state !== 'completed' ? `
                            <button class="pause-btn px-2 py-1 bg-yellow-500 text-white rounded hover:bg-yellow-600" data-hash="${torrent.hash}">
                                <i class="fas fa-pause"></i>
                            </button>
                        ` : ''}
                        <button class="delete-btn px-2 py-1 bg-red-500 text-white rounded hover:bg-red-600" data-hash="${torrent.hash}">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
                ${torrent.state !== 'completed' ? `
                    <div class="relative">
                        <div class="bg-gray-700 rounded-full h-2 overflow-hidden">
                            <div class="bg-gradient-to-r from-blue-500 to-blue-600 h-full transition-all duration-300" 
                                 style="width: ${progress}%"></div>
                        </div>
                        <span class="absolute right-0 -top-5 text-xs text-gray-400">${progress}%</span>
                    </div>
                ` : ''}
            </div>
        `;
    }

    // 获取状态信息
    function getStatusInfo(state) {
        const states = {
            'downloading': { text: '下载中', class: 'text-blue-400' },
            'pausedDL': { text: '已暂停', class: 'text-yellow-400' },
            'queuedDL': { text: '排队中', class: 'text-gray-400' },
            'stalledDL': { text: '等待中', class: 'text-orange-400' },
            'completed': { text: '已完成', class: 'text-green-400' },
            'seeding': { text: '做种中', class: 'text-green-400' },
            'error': { text: '错误', class: 'text-red-400' }
        };
        return states[state] || { text: state, class: 'text-gray-400' };
    }

    // 设置下载控制
    function setupDownloadControls() {
        // 暂停按钮
        document.querySelectorAll('.pause-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const hash = e.currentTarget.dataset.hash;
                controlTorrent(hash, 'pause');
            });
        });

        // 继续按钮
        document.querySelectorAll('.resume-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const hash = e.currentTarget.dataset.hash;
                controlTorrent(hash, 'resume');
            });
        });

        // 删除按钮
        document.querySelectorAll('.delete-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const hash = e.currentTarget.dataset.hash;
                if (confirm('确定要删除这个下载任务吗？')) {
                    controlTorrent(hash, 'delete');
                }
            });
        });
    }

    // 控制种子
    async function controlTorrent(hash, action) {
        try {
            const response = await fetch(`/api/v1/torrents/${action}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ hash: hash })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const result = await response.json();
            
            if (result.success) {
                showNotification(`操作成功`, 'success');
                // 刷新列表
                setTimeout(() => {
                    loadQBittorrentStatus();
                }, 500);
            } else {
                showNotification(result.error || '操作失败', 'error');
            }
        } catch (error) {
            console.error('操作失败:', error);
            showNotification('操作失败: ' + error.message, 'error');
        }
    }

    // 清理已完成的下载
    async function clearCompletedDownloads() {
        if (!confirm('确定要清理所有已完成的下载任务吗？')) {
            return;
        }

        try {
            const response = await fetch('/api/v1/torrents/clear-completed', {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const result = await response.json();
            
            if (result.success) {
                showNotification('已清理完成的任务', 'success');
                loadQBittorrentStatus();
            } else {
                showNotification(result.error || '清理失败', 'error');
            }
        } catch (error) {
            console.error('清理失败:', error);
            showNotification('清理失败: ' + error.message, 'error');
        }
    }

    // 更新下载统计
    function updateDownloadStats(torrents) {
        if (!torrents) return;

        // 过滤只显示PornDB标签的种子
        const porndbTorrents = torrents.filter(torrent => {
            const tags = torrent.tags || '';
            return tags.includes('PornDB');
        });

        const downloading = porndbTorrents.filter(t => t.state === 'downloading').length;
        const completed = porndbTorrents.filter(t => t.state === 'completed' || t.state === 'seeding').length;
        const total = porndbTorrents.length;
        
        const totalSpeed = porndbTorrents.reduce((sum, t) => sum + (t.dlspeed || 0), 0);

        const statsContainer = document.getElementById('downloadStats');
        if (statsContainer) {
            statsContainer.innerHTML = `
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                    <div class="bg-gray-800 p-4 rounded-lg">
                        <div class="text-gray-400 text-sm">总任务</div>
                        <div class="text-2xl font-bold text-white">${total}</div>
                    </div>
                    <div class="bg-gray-800 p-4 rounded-lg">
                        <div class="text-gray-400 text-sm">下载中</div>
                        <div class="text-2xl font-bold text-blue-400">${downloading}</div>
                    </div>
                    <div class="bg-gray-800 p-4 rounded-lg">
                        <div class="text-gray-400 text-sm">已完成</div>
                        <div class="text-2xl font-bold text-green-400">${completed}</div>
                    </div>
                    <div class="bg-gray-800 p-4 rounded-lg">
                        <div class="text-gray-400 text-sm">下载速度</div>
                        <div class="text-xl font-bold text-white">${formatFileSize(totalSpeed)}/s</div>
                    </div>
                </div>
            `;
        }
    }

    // 显示搜索加载状态
    function showSearchLoading() {
        const container = document.getElementById('searchResults');
        if (!container) return;

        container.innerHTML = `
            <div class="text-center py-12">
                <i class="fas fa-spinner fa-spin text-4xl text-blue-500 mb-4"></i>
                <p class="text-gray-400">搜索种子中...</p>
            </div>
        `;
    }

    // 显示搜索错误
    function showSearchError(message) {
        const container = document.getElementById('searchResults');
        if (!container) return;

        container.innerHTML = `
            <div class="text-center py-12">
                <i class="fas fa-exclamation-triangle text-6xl text-red-500 mb-4"></i>
                <p class="text-gray-400">${message}</p>
            </div>
        `;
    }

    // 显示下载错误
    function showDownloadsError(message) {
        const container = document.getElementById('downloadsList');
        if (!container) return;

        container.innerHTML = `
            <div class="text-center py-12">
                <i class="fas fa-exclamation-circle text-6xl text-yellow-500 mb-4"></i>
                <p class="text-gray-400">${message}</p>
                <a href="/config.html" class="mt-4 inline-block px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
                    配置下载客户端
                </a>
            </div>
        `;
    }

    // 复制到剪贴板
    function copyToClipboard(text) {
        if (navigator.clipboard) {
            navigator.clipboard.writeText(text).then(() => {
                showNotification('已复制到剪贴板', 'success');
            }).catch(() => {
                fallbackCopy(text);
            });
        } else {
            fallbackCopy(text);
        }
    }

    // 降级复制方法
    function fallbackCopy(text) {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();
        
        try {
            document.execCommand('copy');
            showNotification('已复制到剪贴板', 'success');
        } catch (err) {
            showNotification('复制失败', 'error');
        }
        
        document.body.removeChild(textarea);
    }

    // 显示通知
    function showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `fixed top-4 right-4 p-4 rounded-lg shadow-lg z-50 ${
            type === 'success' ? 'bg-green-500' : 
            type === 'error' ? 'bg-red-500' : 
            type === 'warning' ? 'bg-yellow-500' :
            'bg-blue-500'
        } text-white`;
        notification.innerHTML = message;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 3000);
    }

    // 格式化文件大小
    function formatFileSize(bytes) {
        if (!bytes || bytes === 0) return '0 B';
        
        const units = ['B', 'KB', 'MB', 'GB', 'TB'];
        const k = 1024;
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + units[i];
    }

    // 格式化时间
    function formatTime(seconds) {
        if (!seconds || seconds <= 0) return '-';
        
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        
        if (hours > 0) {
            return `${hours}小时${minutes}分钟`;
        } else if (minutes > 0) {
            return `${minutes}分钟`;
        } else {
            return `${seconds}秒`;
        }
    }

    // 格式化日期
    function formatDate(dateStr) {
        if (!dateStr) return '未知';
        
        const date = new Date(dateStr);
        const now = new Date();
        const diff = now - date;
        
        if (diff < 86400000) {
            return '今天';
        } else if (diff < 172800000) {
            return '昨天';
        } else if (diff < 604800000) {
            return Math.floor(diff / 86400000) + '天前';
        } else {
            return date.toLocaleDateString('zh-CN');
        }
    }

    // 页面加载完成后初始化
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();