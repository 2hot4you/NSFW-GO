/**
 * 侧边栏导航逻辑
 * 处理侧边栏的渲染、折叠/展开以及移动端适配
 */
class SidebarNavigation {
    constructor() {
        this.sidebar = null;
        this.isCollapsed = localStorage.getItem('sidebarCollapsed') === 'true';
        this.currentPath = window.location.pathname;

        this.menuItems = [
            { icon: 'fas fa-home', label: '仪表盘', path: '/index.html', exact: true },
            { icon: 'fas fa-search', label: '搜索', path: '/search.html' },
            { icon: 'fas fa-folder-open', label: '本地媒体', path: '/local-movies.html' },
            { icon: 'fas fa-trophy', label: '排行榜', path: '/rankings.html' },
            { icon: 'fas fa-cogs', label: '系统配置', path: '/config.html' },
            // { icon: 'fas fa-download', label: '下载管理', path: '/downloads.html' },
            // { icon: 'fas fa-history', label: '系统日志', path: '/logs.html' }
        ];

        this.init();
    }

    init() {
        // 创建侧边栏元素
        this.createSidebar();

        // 绑定事件
        this.bindEvents();

        // 设置初始状态
        this.updateState();
    }

    createSidebar() {
        const sidebar = document.createElement('aside');
        sidebar.className = 'sidebar';
        sidebar.id = 'appSidebar';

        // Logo 区域
        const logoHtml = `
            <div class="sidebar-header">
                <div class="logo-icon">
                    <i class="fas fa-play-circle"></i>
                </div>
                <span class="logo-text">NSFW-GO</span>
            </div>
        `;

        // 导航菜单
        const menuHtml = `
            <nav class="sidebar-nav">
                ${this.menuItems.map(item => this.createMenuItem(item)).join('')}
            </nav>
        `;

        // 底部折叠按钮
        const footerHtml = `
            <div class="sidebar-footer">
                <button class="collapse-btn" id="sidebarCollapseBtn">
                    <i class="fas fa-chevron-left"></i>
                </button>
            </div>
        `;

        sidebar.innerHTML = logoHtml + menuHtml + footerHtml;

        // 插入到页面最前
        document.body.insertBefore(sidebar, document.body.firstChild);
        this.sidebar = sidebar;

        // 添加样式
        this.addStyles();
    }

    createMenuItem(item) {
        // 简单的路由匹配
        const isActive = item.exact
            ? (this.currentPath === item.path || (item.path === '/index.html' && this.currentPath === '/'))
            : this.currentPath.startsWith(item.path);

        return `
            <a href="${item.path}" class="nav-item ${isActive ? 'active' : ''}">
                <div class="nav-icon">
                    <i class="${item.icon}"></i>
                </div>
                <span class="nav-label">${item.label}</span>
                ${isActive ? '<div class="active-indicator"></div>' : ''}
            </a>
        `;
    }

    addStyles() {
        const style = document.createElement('style');
        style.textContent = `
            .sidebar {
                position: fixed;
                top: 0;
                left: 0;
                height: 100vh;
                width: var(--sidebar-width);
                background-color: var(--bg-sidebar);
                border-right: 1px solid var(--border-subtle);
                display: flex;
                flex-direction: column;
                transition: width var(--duration-normal) var(--ease-out);
                z-index: 40;
            }

            .sidebar.collapsed {
                width: var(--sidebar-collapsed-width);
            }

            .sidebar-header {
                height: var(--header-height);
                display: flex;
                align-items: center;
                padding: 0 1.5rem;
                border-bottom: 1px solid var(--border-subtle);
                overflow: hidden;
                white-space: nowrap;
            }

            .logo-icon {
                font-size: 1.5rem;
                color: var(--primary);
                min-width: 24px;
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .logo-text {
                margin-left: 1rem;
                font-weight: 700;
                font-size: 1.25rem;
                opacity: 1;
                transition: opacity var(--duration-fast);
            }

            .sidebar.collapsed .logo-text {
                opacity: 0;
                pointer-events: none;
            }

            .sidebar-nav {
                flex: 1;
                padding: 1.5rem 0.75rem;
                overflow-y: auto;
                display: flex;
                flex-direction: column;
                gap: 0.5rem;
            }

            .nav-item {
                display: flex;
                align-items: center;
                padding: 0.75rem;
                border-radius: var(--radius-md);
                color: var(--text-muted);
                position: relative;
                overflow: hidden;
                white-space: nowrap;
                transition: all var(--duration-fast);
            }

            .nav-item:hover {
                background-color: rgba(255, 255, 255, 0.05);
                color: var(--text-main);
            }

            .nav-item.active {
                background-color: var(--primary-dim);
                color: var(--primary);
            }

            .nav-icon {
                min-width: 24px;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 1.125rem;
            }

            .nav-label {
                margin-left: 1rem;
                font-weight: 500;
                opacity: 1;
                transition: opacity var(--duration-fast);
            }

            .sidebar.collapsed .nav-label {
                opacity: 0;
                pointer-events: none;
            }

            .sidebar-footer {
                padding: 1rem;
                border-top: 1px solid var(--border-subtle);
                display: flex;
                justify-content: flex-end;
            }

            .collapse-btn {
                background: transparent;
                border: none;
                color: var(--text-muted);
                cursor: pointer;
                padding: 0.5rem;
                border-radius: var(--radius-md);
                transition: all var(--duration-fast);
            }

            .collapse-btn:hover {
                background-color: rgba(255, 255, 255, 0.05);
                color: var(--text-main);
            }

            .sidebar.collapsed .collapse-btn {
                transform: rotate(180deg);
                margin: 0 auto;
            }
            
            /* 移动端适配 */
            @media (max-width: 768px) {
                .sidebar {
                    transform: translateX(-100%);
                    width: 100%;
                    max-width: 280px;
                }
                
                .sidebar.mobile-open {
                    transform: translateX(0);
                }
                
                .sidebar-footer {
                    display: none;
                }
            }
        `;
        document.head.appendChild(style);
    }

    bindEvents() {
        const collapseBtn = document.getElementById('sidebarCollapseBtn');
        if (collapseBtn) {
            collapseBtn.addEventListener('click', () => this.toggleCollapse());
        }
    }

    toggleCollapse() {
        this.isCollapsed = !this.isCollapsed;
        localStorage.setItem('sidebarCollapsed', this.isCollapsed);
        this.updateState();
    }

    updateState() {
        if (this.isCollapsed) {
            this.sidebar.classList.add('collapsed');
            document.body.classList.add('sidebar-collapsed');
            // 更新主内容区域的 margin
            const mainContent = document.querySelector('.main-content');
            if (mainContent) {
                mainContent.style.marginLeft = 'var(--sidebar-collapsed-width)';
            }
        } else {
            this.sidebar.classList.remove('collapsed');
            document.body.classList.remove('sidebar-collapsed');
            const mainContent = document.querySelector('.main-content');
            if (mainContent) {
                mainContent.style.marginLeft = 'var(--sidebar-width)';
            }
        }
    }
}

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    window.sidebarNav = new SidebarNavigation();
});
