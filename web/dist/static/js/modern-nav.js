// 现代化导航栏组件
(function() {
    'use strict';
    
    // 初始化导航栏
    function initNavigation() {
        const workMode = localStorage.getItem('workMode') === 'true';
        const currentPath = window.location.pathname;
        
        // 导航项配置
        const navItems = [
            { href: '/', icon: 'fas fa-home', text: '主页', path: ['/', '/index.html'] },
            { href: '/search.html', icon: 'fas fa-search', text: '搜索', path: ['/search.html'] },
            { href: '/local-movies.html', icon: 'fas fa-folder', text: '本地库', path: ['/local-movies.html'] },
            { href: '/rankings.html', icon: 'fas fa-trophy', text: '排行榜', path: ['/rankings.html'] },
            { href: '/downloads.html', icon: 'fas fa-download', text: '下载', path: ['/downloads.html'] },
            { href: '/logs.html', icon: 'fas fa-file-alt', text: '日志', path: ['/logs.html'] }
        ];
        
        // 生成导航链接HTML
        const navLinksHTML = navItems.map(item => {
            const isActive = item.path.some(p => currentPath === p || currentPath.endsWith(p));
            return `
                <a href="${item.href}" class="nav-link ${isActive ? 'active' : ''}">
                    <i class="${item.icon}"></i>
                    <span>${item.text}</span>
                </a>
            `;
        }).join('');
        
        // 创建导航栏HTML
        const navHTML = `
            <nav id="mainNav" class="main-navigation">
                <div class="nav-container">
                    <!-- Logo区域 -->
                    <div class="nav-brand">
                        <a href="/" class="brand-link">
                            <div class="brand-icon">
                                <i class="fas fa-video"></i>
                            </div>
                            <span class="brand-text">NSFW-Go</span>
                        </a>
                        <span class="brand-badge">智能影视库</span>
                    </div>
                    
                    <!-- 导航链接 -->
                    <div class="nav-links">
                        ${navLinksHTML}
                    </div>
                    
                    <!-- 右侧操作区 -->
                    <div class="nav-actions">
                        <!-- 快速搜索 -->
                        <button class="nav-btn nav-btn-search" onclick="openQuickSearch()" title="快速搜索 (Ctrl+K)">
                            <i class="fas fa-search"></i>
                            <span class="btn-text">搜索</span>
                        </button>
                        
                        <!-- 工作模式 -->
                        <button id="workModeBtn" class="nav-btn nav-btn-work ${workMode ? 'active' : ''}" 
                                onclick="toggleWorkMode()" 
                                title="${workMode ? '工作模式已开启' : '工作模式已关闭'}">
                            <i class="fas ${workMode ? 'fa-eye-slash' : 'fa-eye'}"></i>
                            <span class="btn-text">工作模式</span>
                        </button>
                        
                        <!-- 设置 -->
                        <a href="/config.html" class="nav-btn nav-btn-config" title="系统配置">
                            <i class="fas fa-cog"></i>
                        </a>
                        
                        <!-- 移动端菜单 -->
                        <button class="nav-btn nav-btn-mobile" onclick="toggleMobileMenu()">
                            <i class="fas fa-bars"></i>
                        </button>
                    </div>
                </div>
                
                <!-- 移动端菜单 -->
                <div id="mobileMenu" class="mobile-menu">
                    <div class="mobile-menu-content">
                        ${navLinksHTML}
                        <div class="mobile-menu-divider"></div>
                        <button class="mobile-work-btn ${workMode ? 'active' : ''}" onclick="toggleWorkMode()">
                            <i class="fas ${workMode ? 'fa-eye-slash' : 'fa-eye'}"></i>
                            <span>${workMode ? '工作模式：开' : '工作模式：关'}</span>
                        </button>
                    </div>
                </div>
            </nav>
        `;
        
        // 插入导航栏
        document.body.insertAdjacentHTML('afterbegin', navHTML);
        
        // 添加导航栏样式
        addNavigationStyles();
        
        // 应用工作模式
        if (workMode) {
            applyWorkMode();
        }
        
        // 初始化快捷键
        initKeyboardShortcuts();
    }
    
    // 添加导航栏样式
    function addNavigationStyles() {
        const style = document.createElement('style');
        style.id = 'modern-nav-styles';
        style.textContent = `
            /* 导航栏主样式 - 纯黑背景适配 */
            .main-navigation {
                position: sticky;
                top: 0;
                z-index: 1000;
                background: rgba(0, 0, 0, 0.95);
                backdrop-filter: blur(12px);
                border-bottom: 1px solid rgba(139, 92, 246, 0.3);
                box-shadow: 0 4px 20px rgba(139, 92, 246, 0.15);
            }
            
            .nav-container {
                width: 80%;
                max-width: 1400px;
                margin: 0 auto;
                padding: 0 clamp(16px, 3vw, 32px);
                height: 64px;
                display: flex;
                align-items: center;
                justify-content: space-between;
            }
            
            /* Logo区域 */
            .nav-brand {
                display: flex;
                align-items: center;
                gap: 0.75rem;
            }
            
            .brand-link {
                display: flex;
                align-items: center;
                gap: 0.75rem;
                text-decoration: none;
                transition: transform 0.2s;
            }
            
            .brand-link:hover {
                transform: scale(1.05);
            }
            
            .brand-icon {
                width: 40px;
                height: 40px;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                border-radius: 12px;
                display: flex;
                align-items: center;
                justify-content: center;
                color: white;
                font-size: 20px;
                box-shadow: 0 4px 6px rgba(102, 126, 234, 0.3);
                animation: float 3s ease-in-out infinite;
            }
            
            @keyframes float {
                0%, 100% { transform: translateY(0); }
                50% { transform: translateY(-3px); }
            }
            
            .brand-text {
                font-size: 1.5rem;
                font-weight: bold;
                background: linear-gradient(90deg, #fff 0%, #e0e7ff 100%);
                -webkit-background-clip: text;
                -webkit-text-fill-color: transparent;
                background-clip: text;
            }
            
            .brand-badge {
                padding: 0.25rem 0.75rem;
                background: rgba(99, 102, 241, 0.2);
                color: #a5b4fc;
                border-radius: 9999px;
                font-size: 0.75rem;
                border: 1px solid rgba(99, 102, 241, 0.3);
            }
            
            /* 导航链接 */
            .nav-links {
                display: none;
                gap: 0.5rem;
            }
            
            @media (min-width: 768px) {
                .nav-links {
                    display: flex;
                }
            }
            
            .nav-link {
                padding: 0.5rem 1rem;
                color: #9ca3af;
                text-decoration: none;
                border-radius: 8px;
                transition: all 0.2s;
                display: flex;
                align-items: center;
                gap: 0.5rem;
                font-size: 0.95rem;
            }
            
            .nav-link:hover {
                background: rgba(99, 102, 241, 0.1);
                color: #fff;
            }
            
            .nav-link.active {
                background: rgba(99, 102, 241, 0.2);
                color: #818cf8;
                font-weight: 500;
            }
            
            /* 操作按钮 */
            .nav-actions {
                display: flex;
                align-items: center;
                gap: 0.5rem;
            }
            
            .nav-btn {
                padding: 0.5rem 1rem;
                background: rgba(75, 85, 99, 0.3);
                color: #d1d5db;
                border: 1px solid rgba(255, 255, 255, 0.1);
                border-radius: 8px;
                cursor: pointer;
                transition: all 0.2s;
                display: flex;
                align-items: center;
                gap: 0.5rem;
                font-size: 0.9rem;
            }
            
            .nav-btn:hover {
                background: rgba(99, 102, 241, 0.2);
                border-color: rgba(99, 102, 241, 0.3);
                transform: translateY(-1px);
            }
            
            .nav-btn-work.active {
                background: linear-gradient(135deg, #10b981 0%, #059669 100%);
                border-color: #10b981;
                color: white;
            }
            
            .nav-btn-work.active:hover {
                background: linear-gradient(135deg, #059669 0%, #047857 100%);
            }
            
            .btn-text {
                display: none;
            }
            
            @media (min-width: 1024px) {
                .btn-text {
                    display: inline;
                }
            }
            
            .nav-btn-mobile {
                display: flex;
            }
            
            @media (min-width: 768px) {
                .nav-btn-mobile {
                    display: none;
                }
            }
            
            /* 移动端菜单 */
            .mobile-menu {
                display: none;
                position: absolute;
                top: 100%;
                left: 0;
                right: 0;
                background: rgba(17, 24, 39, 0.98);
                backdrop-filter: blur(12px);
                border-bottom: 1px solid rgba(255, 255, 255, 0.1);
                box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
            }
            
            .mobile-menu.show {
                display: block;
            }
            
            .mobile-menu-content {
                padding: 1rem;
                display: flex;
                flex-direction: column;
                gap: 0.5rem;
            }
            
            .mobile-menu .nav-link {
                width: 100%;
                padding: 0.75rem 1rem;
            }
            
            .mobile-menu-divider {
                height: 1px;
                background: rgba(255, 255, 255, 0.1);
                margin: 0.5rem 0;
            }
            
            .mobile-work-btn {
                padding: 0.75rem 1rem;
                background: rgba(75, 85, 99, 0.3);
                color: #d1d5db;
                border: 1px solid rgba(255, 255, 255, 0.1);
                border-radius: 8px;
                cursor: pointer;
                display: flex;
                align-items: center;
                justify-content: space-between;
                transition: all 0.2s;
            }
            
            .mobile-work-btn.active {
                background: linear-gradient(135deg, #10b981 0%, #059669 100%);
                color: white;
            }
            
            /* 工作模式指示器 */
            .work-mode-indicator {
                position: fixed;
                bottom: 20px;
                left: 20px;
                background: linear-gradient(135deg, #10b981 0%, #059669 100%);
                color: white;
                padding: 0.75rem 1.25rem;
                border-radius: 9999px;
                font-size: 0.875rem;
                box-shadow: 0 10px 15px -3px rgba(16, 185, 129, 0.3);
                display: flex;
                align-items: center;
                gap: 0.5rem;
                animation: pulse 2s infinite;
                z-index: 999;
            }
            
            @keyframes pulse {
                0%, 100% { opacity: 1; transform: scale(1); }
                50% { opacity: 0.9; transform: scale(0.98); }
            }
        `;
        document.head.appendChild(style);
    }
    
    // 切换工作模式
    function toggleWorkMode() {
        const currentMode = localStorage.getItem('workMode') === 'true';
        const newMode = !currentMode;
        localStorage.setItem('workMode', String(newMode));
        
        // 更新按钮状态
        const btn = document.getElementById('workModeBtn');
        const mobileBtn = document.querySelector('.mobile-work-btn');
        
        if (btn) {
            btn.classList.toggle('active', newMode);
            btn.title = newMode ? '工作模式已开启' : '工作模式已关闭';
            btn.querySelector('i').className = `fas ${newMode ? 'fa-eye-slash' : 'fa-eye'}`;
        }
        
        if (mobileBtn) {
            mobileBtn.classList.toggle('active', newMode);
            mobileBtn.innerHTML = `
                <i class="fas ${newMode ? 'fa-eye-slash' : 'fa-eye'}"></i>
                <span>${newMode ? '工作模式：开' : '工作模式：关'}</span>
            `;
        }
        
        // 应用/移除工作模式
        if (newMode) {
            applyWorkMode();
        } else {
            removeWorkMode();
        }
        
        // 触发自定义事件
        window.dispatchEvent(new CustomEvent('workModeChanged', { 
            detail: { enabled: newMode } 
        }));
    }
    
    // 应用工作模式
    function applyWorkMode() {
        document.body.classList.add('work-mode');
        
        // 添加模糊样式
        let styleEl = document.getElementById('work-mode-blur');
        if (!styleEl) {
            styleEl = document.createElement('style');
            styleEl.id = 'work-mode-blur';
            styleEl.textContent = `
                .work-mode img:not(.no-blur),
                .work-mode video,
                .work-mode .movie-cover,
                .work-mode .cover-image {
                    filter: blur(25px) brightness(0.3) !important;
                    transition: filter 0.3s;
                }
                
                .work-mode img:not(.no-blur):hover,
                .work-mode .movie-cover:hover,
                .work-mode .cover-image:hover {
                    filter: blur(15px) brightness(0.5) !important;
                }
                
                /* 导航栏图标不模糊 */
                .brand-icon,
                .nav-link i,
                .nav-btn i {
                    filter: none !important;
                }
            `;
            document.head.appendChild(styleEl);
        }
        
        // 显示指示器
        showWorkModeIndicator();
    }
    
    // 移除工作模式
    function removeWorkMode() {
        document.body.classList.remove('work-mode');
        
        const styleEl = document.getElementById('work-mode-blur');
        if (styleEl) {
            styleEl.remove();
        }
        
        hideWorkModeIndicator();
    }
    
    // 显示工作模式指示器
    function showWorkModeIndicator() {
        if (!document.getElementById('workModeIndicator')) {
            const indicator = document.createElement('div');
            indicator.id = 'workModeIndicator';
            indicator.className = 'work-mode-indicator';
            indicator.innerHTML = '🙈 工作模式已开启';
            document.body.appendChild(indicator);
        }
    }
    
    // 隐藏工作模式指示器
    function hideWorkModeIndicator() {
        const indicator = document.getElementById('workModeIndicator');
        if (indicator) {
            indicator.remove();
        }
    }
    
    // 切换移动端菜单
    function toggleMobileMenu() {
        const menu = document.getElementById('mobileMenu');
        if (menu) {
            menu.classList.toggle('show');
        }
    }
    
    // 快速搜索
    function openQuickSearch() {
        // 如果页面有自定义搜索函数，使用它
        if (typeof window.openSearchModal === 'function') {
            window.openSearchModal();
        } else {
            // 否则跳转到搜索页面
            window.location.href = '/search.html';
        }
    }
    
    // 初始化快捷键
    function initKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            // Ctrl/Cmd + K 打开搜索
            if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
                e.preventDefault();
                openQuickSearch();
            }
            
            // Ctrl/Cmd + W 切换工作模式
            if ((e.ctrlKey || e.metaKey) && e.key === 'w') {
                e.preventDefault();
                toggleWorkMode();
            }
        });
    }
    
    // 导出全局函数
    window.toggleWorkMode = toggleWorkMode;
    window.toggleMobileMenu = toggleMobileMenu;
    window.openQuickSearch = openQuickSearch;
    
    // DOM加载完成后初始化
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initNavigation);
    } else {
        initNavigation();
    }
})();