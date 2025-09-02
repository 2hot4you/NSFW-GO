// 🎯 NSFW-GO 导航栏组件
// 提供统一的导航栏和页面布局

class Navigation {
    constructor() {
        this.currentPage = this.getCurrentPage();
        this.workMode = localStorage.getItem('workMode') === 'true';
        this.init();
    }

    init() {
        // 渲染导航栏
        this.renderNavigation();
        // 设置活动页面
        this.setActivePage();
        // 初始化快捷键
        this.initKeyboardShortcuts();
        // 应用工作模式
        this.applyWorkMode();
    }

    getCurrentPage() {
        const path = window.location.pathname;
        if (path === '/' || path === '/index.html') return 'dashboard';
        if (path.includes('search')) return 'search';
        if (path.includes('local-movies')) return 'local';
        if (path.includes('rankings')) return 'rankings';
        if (path.includes('downloads')) return 'downloads';
        if (path.includes('config')) return 'config';
        if (path.includes('test-connections')) return 'test';
        return 'dashboard';
    }

    renderNavigation() {
        // 检查是否已有背景，避免重复
        const hasBackground = document.querySelector('.hero-gradient');
        
        const navHTML = `
            ${!hasBackground ? `
            <!-- 动态背景 -->
            <div class="fixed inset-0 -z-10">
                <div class="absolute inset-0 bg-dark"></div>
                <div class="hero-gradient absolute inset-0"></div>
            </div>
            ` : ''}

            <!-- 页面加载进度条 -->
            <div id="pageProgress" class="fixed top-0 left-0 right-0 h-1 z-50">
                <div class="h-full bg-gradient-to-r from-primary via-secondary to-primary" style="width: 0%; transition: width 0.3s;"></div>
            </div>

            <!-- 导航栏 -->
            <nav class="glass-card sticky top-0 z-40" style="border-radius: 0; border-left: 0; border-right: 0;">
                <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div class="flex-between h-16">
                        <!-- Logo和标题 -->
                        <div class="flex items-center gap-4">
                            <a href="/" class="flex items-center gap-2">
                                <div class="stat-icon float-animation" style="width: 40px; height: 40px;">
                                    <i class="fas fa-video"></i>
                                </div>
                                <span class="text-2xl font-bold glow-text">NSFW-Go</span>
                            </a>
                            <span class="badge badge-primary hidden sm:block">智能影视库</span>
                        </div>
                        
                        <!-- 主导航 -->
                        <div class="hidden md:flex items-center gap-2">
                            <a href="/" class="nav-link ${this.currentPage === 'dashboard' ? 'active' : ''}" data-page="dashboard">
                                <i class="fas fa-home"></i>
                                <span>主页</span>
                            </a>
                            <a href="/search.html" class="nav-link ${this.currentPage === 'search' ? 'active' : ''}" data-page="search">
                                <i class="fas fa-search"></i>
                                <span>搜索</span>
                            </a>
                            <a href="/local-movies.html" class="nav-link ${this.currentPage === 'local' ? 'active' : ''}" data-page="local">
                                <i class="fas fa-folder"></i>
                                <span>本地库</span>
                            </a>
                            <a href="/rankings.html" class="nav-link ${this.currentPage === 'rankings' ? 'active' : ''}" data-page="rankings">
                                <i class="fas fa-trophy"></i>
                                <span>排行榜</span>
                            </a>
                            <a href="/downloads.html" class="nav-link ${this.currentPage === 'downloads' ? 'active' : ''}" data-page="downloads">
                                <i class="fas fa-download"></i>
                                <span>下载</span>
                            </a>
                        </div>
                        
                        <!-- 右侧操作 -->
                        <div class="flex items-center gap-3">
                            <!-- 搜索按钮 -->
                            <button 
                                onclick="navigation.openQuickSearch()"
                                class="btn btn-secondary hidden sm:flex"
                                data-tooltip="🔍 快速搜索 (⌘K)"
                            >
                                <i class="fas fa-search"></i>
                                <span class="hidden lg:block">搜索</span>
                            </button>
                            
                            <!-- 工作模式切换 -->
                            <button 
                                id="workModeToggle"
                                onclick="navigation.toggleWorkMode()"
                                class="btn ${this.workMode ? 'btn-success' : 'btn-secondary'}"
                                data-tooltip="${this.workMode ? '🙈 工作模式已开启' : '👁️ 工作模式已关闭'}"
                            >
                                <i class="fas ${this.workMode ? 'fa-eye-slash' : 'fa-eye'}"></i>
                                <span class="hidden lg:block">工作模式</span>
                            </button>
                            
                            <!-- 设置按钮 -->
                            <a href="/config.html" class="btn btn-secondary" data-tooltip="⚙️ 系统配置">
                                <i class="fas fa-cogs"></i>
                            </a>
                            
                            <!-- 移动端菜单 -->
                            <button onclick="navigation.toggleMobileMenu()" class="btn btn-secondary md:hidden">
                                <i class="fas fa-bars"></i>
                            </button>
                        </div>
                    </div>
                </div>
                
                <!-- 移动端菜单 -->
                <div id="mobileMenu" class="hidden md:hidden border-t border-white/10">
                    <div class="px-4 py-3 space-y-2">
                        <a href="/" class="mobile-nav-link ${this.currentPage === 'dashboard' ? 'active' : ''}">
                            <i class="fas fa-home"></i>
                            <span>主页</span>
                        </a>
                        <a href="/search.html" class="mobile-nav-link ${this.currentPage === 'search' ? 'active' : ''}">
                            <i class="fas fa-search"></i>
                            <span>搜索</span>
                        </a>
                        <a href="/local-movies.html" class="mobile-nav-link ${this.currentPage === 'local' ? 'active' : ''}">
                            <i class="fas fa-folder"></i>
                            <span>本地库</span>
                        </a>
                        <a href="/rankings.html" class="mobile-nav-link ${this.currentPage === 'rankings' ? 'active' : ''}">
                            <i class="fas fa-trophy"></i>
                            <span>排行榜</span>
                        </a>
                        <a href="/downloads.html" class="mobile-nav-link ${this.currentPage === 'downloads' ? 'active' : ''}">
                            <i class="fas fa-download"></i>
                            <span>下载</span>
                        </a>
                        <div class="pt-3 border-t border-white/10">
                            <button 
                                onclick="navigation.toggleWorkMode()"
                                class="mobile-nav-link justify-between w-full"
                            >
                                <span>
                                    <i class="fas ${this.workMode ? 'fa-eye-slash' : 'fa-eye'}"></i>
                                    工作模式
                                </span>
                                <span class="badge ${this.workMode ? 'badge-success' : 'badge-secondary'}">
                                    ${this.workMode ? '开启' : '关闭'}
                                </span>
                            </button>
                        </div>
                    </div>
                </div>
            </nav>

            <!-- 快速搜索模态框 -->
            <div id="quickSearchModal" class="modal hidden">
                <div class="modal-content max-w-2xl">
                    <div class="modal-header">
                        <h3 class="text-xl font-semibold">🔍 快速搜索</h3>
                        <button onclick="navigation.closeQuickSearch()" class="modal-close">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                    <div class="modal-body">
                        <div class="input-group">
                            <input 
                                type="text" 
                                id="quickSearchInput" 
                                placeholder="输入影片名称或番号..."
                                onkeyup="navigation.handleQuickSearch(event)"
                                autofocus
                            >
                        </div>
                        <div id="quickSearchResults" class="mt-4 space-y-2">
                            <!-- 搜索结果 -->
                        </div>
                    </div>
                </div>
            </div>
        `;

        // 插入导航栏
        document.body.insertAdjacentHTML('afterbegin', navHTML);
    }

    setActivePage() {
        // 高亮当前页面
        document.querySelectorAll('.nav-link').forEach(link => {
            if (link.dataset.page === this.currentPage) {
                link.classList.add('active');
            }
        });
    }

    // 切换工作模式
    toggleWorkMode() {
        this.workMode = !this.workMode;
        localStorage.setItem('workMode', this.workMode);
        this.applyWorkMode();
        
        // 更新按钮状态
        const btn = document.getElementById('workModeToggle');
        if (btn) {
            btn.className = `btn ${this.workMode ? 'btn-success' : 'btn-secondary'}`;
            btn.setAttribute('data-tooltip', this.workMode ? '🙈 工作模式已开启' : '👁️ 工作模式已关闭');
            btn.innerHTML = `
                <i class="fas ${this.workMode ? 'fa-eye-slash' : 'fa-eye'}"></i>
                <span class="hidden lg:block">工作模式</span>
            `;
        }

        // 更新移动端按钮
        const mobileBtn = document.querySelector('#mobileMenu button[onclick*="toggleWorkMode"]');
        if (mobileBtn) {
            mobileBtn.innerHTML = `
                <span>
                    <i class="fas ${this.workMode ? 'fa-eye-slash' : 'fa-eye'}"></i>
                    工作模式
                </span>
                <span class="badge ${this.workMode ? 'badge-success' : 'badge-secondary'}">
                    ${this.workMode ? '开启' : '关闭'}
                </span>
            `;
        }

        // 触发自定义事件，让页面响应
        window.dispatchEvent(new CustomEvent('workModeChanged', { detail: { enabled: this.workMode } }));
    }

    // 应用工作模式
    applyWorkMode() {
        if (this.workMode) {
            document.body.classList.add('work-mode');
            // 创建或更新样式
            let styleEl = document.getElementById('work-mode-styles');
            if (!styleEl) {
                styleEl = document.createElement('style');
                styleEl.id = 'work-mode-styles';
                document.head.appendChild(styleEl);
            }
            styleEl.textContent = `
                /* 工作模式样式 - 模糊所有图片 */
                .work-mode img:not(.no-blur),
                .work-mode .movie-cover,
                .work-mode .cover-image,
                .work-mode .thumbnail {
                    filter: blur(20px) brightness(0.5);
                    transition: filter 0.3s;
                }
                
                /* 鼠标悬停时稍微减少模糊 */
                .work-mode img:not(.no-blur):hover,
                .work-mode .movie-cover:hover,
                .work-mode .cover-image:hover,
                .work-mode .thumbnail:hover {
                    filter: blur(10px) brightness(0.7);
                }
                
                /* 工作模式提示 */
                .work-mode::before {
                    content: '🙈 工作模式已开启';
                    position: fixed;
                    bottom: 20px;
                    left: 20px;
                    background: rgba(34, 197, 94, 0.9);
                    color: white;
                    padding: 8px 16px;
                    border-radius: 8px;
                    font-size: 14px;
                    z-index: 9999;
                    animation: pulse 2s infinite;
                }
                
                @keyframes pulse {
                    0%, 100% { opacity: 0.9; }
                    50% { opacity: 0.6; }
                }
            `;
        } else {
            document.body.classList.remove('work-mode');
            const styleEl = document.getElementById('work-mode-styles');
            if (styleEl) {
                styleEl.remove();
            }
        }
    }

    // 切换移动端菜单
    toggleMobileMenu() {
        const menu = document.getElementById('mobileMenu');
        menu.classList.toggle('hidden');
    }

    // 打开快速搜索
    openQuickSearch() {
        const modal = document.getElementById('quickSearchModal');
        modal.classList.remove('hidden');
        document.getElementById('quickSearchInput').focus();
    }

    // 关闭快速搜索
    closeQuickSearch() {
        const modal = document.getElementById('quickSearchModal');
        modal.classList.add('hidden');
        document.getElementById('quickSearchInput').value = '';
        document.getElementById('quickSearchResults').innerHTML = '';
    }

    // 处理快速搜索
    async handleQuickSearch(event) {
        if (event.key === 'Escape') {
            this.closeQuickSearch();
            return;
        }

        const query = event.target.value.trim();
        if (query.length < 2) {
            document.getElementById('quickSearchResults').innerHTML = '';
            return;
        }

        // 搜索本地和在线
        try {
            const [localResponse, onlineResponse] = await Promise.all([
                fetch(`/api/v1/local/movies?search=${encodeURIComponent(query)}&limit=5`),
                fetch(`/api/v1/search/javdb?keyword=${encodeURIComponent(query)}&limit=5`)
            ]);

            const localData = await localResponse.json();
            const onlineData = await onlineResponse.json();

            let resultsHTML = '';

            // 本地结果
            if (localData.success && localData.data && localData.data.length > 0) {
                resultsHTML += '<div class="mb-4"><h4 class="text-sm text-gray-400 mb-2">📁 本地库</h4>';
                localData.data.forEach(movie => {
                    resultsHTML += `
                        <a href="/local-movies.html?search=${movie.code}" class="block p-2 hover:bg-gray-800 rounded">
                            <div class="flex items-center gap-3">
                                <img src="${movie.cover_url || '/static/img/no-cover.jpg'}" 
                                     class="w-12 h-16 object-cover rounded ${this.workMode ? '' : 'no-blur'}"
                                     onerror="this.src='/static/img/no-cover.jpg'">
                                <div>
                                    <div class="text-white">${movie.code}</div>
                                    <div class="text-sm text-gray-400 truncate">${movie.title || '无标题'}</div>
                                </div>
                            </div>
                        </a>
                    `;
                });
                resultsHTML += '</div>';
            }

            // 在线结果
            if (onlineData.success && onlineData.data && onlineData.data.length > 0) {
                resultsHTML += '<div><h4 class="text-sm text-gray-400 mb-2">🌐 在线搜索</h4>';
                onlineData.data.forEach(movie => {
                    resultsHTML += `
                        <a href="/search.html?q=${movie.code}" class="block p-2 hover:bg-gray-800 rounded">
                            <div class="flex items-center gap-3">
                                <img src="${movie.cover || '/static/img/no-cover.jpg'}" 
                                     class="w-12 h-16 object-cover rounded ${this.workMode ? '' : 'no-blur'}"
                                     onerror="this.src='/static/img/no-cover.jpg'">
                                <div>
                                    <div class="text-white">${movie.code}</div>
                                    <div class="text-sm text-gray-400 truncate">${movie.title || '无标题'}</div>
                                </div>
                            </div>
                        </a>
                    `;
                });
                resultsHTML += '</div>';
            }

            if (resultsHTML === '') {
                resultsHTML = '<div class="text-center text-gray-400 py-4">未找到相关结果</div>';
            }

            document.getElementById('quickSearchResults').innerHTML = resultsHTML;
        } catch (error) {
            console.error('快速搜索失败:', error);
        }
    }

    // 初始化键盘快捷键
    initKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            // Cmd/Ctrl + K 打开快速搜索
            if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
                e.preventDefault();
                this.openQuickSearch();
            }
            // Cmd/Ctrl + W 切换工作模式
            if ((e.metaKey || e.ctrlKey) && e.key === 'w') {
                e.preventDefault();
                this.toggleWorkMode();
            }
        });
    }

    // 显示页面加载进度
    showPageProgress(percent) {
        const progress = document.querySelector('#pageProgress > div');
        if (progress) {
            progress.style.width = `${percent}%`;
        }
    }

    // 完成页面加载
    completePageLoad() {
        this.showPageProgress(100);
        setTimeout(() => {
            this.showPageProgress(0);
        }, 500);
    }
}

// 等待DOM加载完成后初始化
let navigation;

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        navigation = new Navigation();
        window.navigation = navigation;
    });
} else {
    // DOM已经加载完成
    navigation = new Navigation();
    window.navigation = navigation;
}

// 页面加载完成
window.addEventListener('load', () => {
    if (navigation) {
        navigation.completePageLoad();
    }
});