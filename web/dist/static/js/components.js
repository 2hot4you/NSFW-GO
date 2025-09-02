// 🎨 NSFW-GO 统一组件库
// 提供可复用的UI组件，确保整个应用的一致性

class NSFWComponents {
    constructor() {
        this.init();
    }

    init() {
        // 初始化全局事件监听
        this.initRippleEffect();
        this.initTooltips();
        this.initModals();
    }

    // 🌊 波纹效果
    initRippleEffect() {
        document.addEventListener('click', (e) => {
            const button = e.target.closest('.btn');
            if (!button) return;

            const ripple = document.createElement('span');
            const rect = button.getBoundingClientRect();
            const size = Math.max(rect.width, rect.height);
            const x = e.clientX - rect.left - size / 2;
            const y = e.clientY - rect.top - size / 2;

            ripple.style.width = ripple.style.height = size + 'px';
            ripple.style.left = x + 'px';
            ripple.style.top = y + 'px';
            ripple.classList.add('ripple');

            button.appendChild(ripple);
            setTimeout(() => ripple.remove(), 600);
        });
    }

    // 💬 提示框
    initTooltips() {
        document.querySelectorAll('[data-tooltip]').forEach(element => {
            element.addEventListener('mouseenter', (e) => {
                const tooltip = document.createElement('div');
                tooltip.className = 'tooltip';
                tooltip.textContent = e.target.dataset.tooltip;
                document.body.appendChild(tooltip);

                const rect = e.target.getBoundingClientRect();
                tooltip.style.left = rect.left + rect.width / 2 - tooltip.offsetWidth / 2 + 'px';
                tooltip.style.top = rect.top - tooltip.offsetHeight - 8 + 'px';
                
                setTimeout(() => tooltip.classList.add('show'), 10);
            });

            element.addEventListener('mouseleave', () => {
                document.querySelectorAll('.tooltip').forEach(t => t.remove());
            });
        });
    }

    // 🪟 模态框
    initModals() {
        document.querySelectorAll('[data-modal-trigger]').forEach(trigger => {
            trigger.addEventListener('click', () => {
                const modalId = trigger.dataset.modalTrigger;
                this.openModal(modalId);
            });
        });

        document.querySelectorAll('[data-modal-close]').forEach(closer => {
            closer.addEventListener('click', () => {
                this.closeModal();
            });
        });
    }

    // 📊 创建统计卡片
    createStatCard(title, value, icon, trend = null) {
        const trendHtml = trend ? `
            <div class="stat-trend ${trend.type}">
                <i class="fas fa-arrow-${trend.type === 'up' ? 'up' : 'down'}"></i>
                <span>${trend.value}%</span>
            </div>
        ` : '';

        return `
            <div class="stat-card fade-in">
                <div class="flex-between">
                    <div>
                        <div class="stat-label">${title}</div>
                        <div class="stat-value">${value}</div>
                        ${trendHtml}
                    </div>
                    <div class="stat-icon">
                        <i class="${icon}"></i>
                    </div>
                </div>
            </div>
        `;
    }

    // 🎴 创建内容卡片
    createContentCard(data) {
        const { title, subtitle, image, tags = [], actions = [] } = data;
        
        const tagsHtml = tags.map(tag => 
            `<span class="badge badge-primary">${tag}</span>`
        ).join('');
        
        const actionsHtml = actions.map(action => 
            `<button class="btn btn-${action.type || 'secondary'}" onclick="${action.onClick}">
                ${action.icon ? `<i class="${action.icon}"></i>` : ''}
                ${action.label}
            </button>`
        ).join('');

        return `
            <div class="glass-card content-card">
                ${image ? `
                    <div class="content-image">
                        <img src="${image}" alt="${title}" loading="lazy">
                    </div>
                ` : ''}
                <div class="content-body">
                    <h3 class="content-title">${title}</h3>
                    ${subtitle ? `<p class="content-subtitle text-muted">${subtitle}</p>` : ''}
                    ${tagsHtml ? `<div class="content-tags">${tagsHtml}</div>` : ''}
                    ${actionsHtml ? `<div class="content-actions">${actionsHtml}</div>` : ''}
                </div>
            </div>
        `;
    }

    // 📋 创建数据表格
    createDataTable(columns, data, options = {}) {
        const { 
            sortable = true, 
            searchable = true, 
            pageable = true,
            pageSize = 10 
        } = options;

        const tableId = 'table-' + Date.now();
        
        const headerHtml = columns.map(col => `
            <th ${sortable ? 'class="sortable" data-sort="' + col.key + '"' : ''}>
                ${col.label}
                ${sortable ? '<i class="fas fa-sort"></i>' : ''}
            </th>
        `).join('');

        const bodyHtml = data.map(row => `
            <tr>
                ${columns.map(col => {
                    const value = this.getNestedValue(row, col.key);
                    const formatted = col.formatter ? col.formatter(value, row) : value;
                    return `<td>${formatted}</td>`;
                }).join('')}
            </tr>
        `).join('');

        const searchHtml = searchable ? `
            <div class="table-search">
                <input type="text" class="input-field" placeholder="搜索..." 
                       onkeyup="components.filterTable('${tableId}', this.value)">
            </div>
        ` : '';

        return `
            <div class="table-container">
                ${searchHtml}
                <table class="data-table" id="${tableId}">
                    <thead>
                        <tr>${headerHtml}</tr>
                    </thead>
                    <tbody>${bodyHtml}</tbody>
                </table>
                ${pageable ? this.createPagination(tableId, data.length, pageSize) : ''}
            </div>
        `;
    }

    // 📄 创建分页组件
    createPagination(tableId, totalItems, pageSize) {
        const totalPages = Math.ceil(totalItems / pageSize);
        
        return `
            <div class="pagination" data-table="${tableId}">
                <button class="btn btn-secondary" onclick="components.changePage('${tableId}', 'prev')">
                    <i class="fas fa-chevron-left"></i>
                </button>
                <span class="page-info">
                    第 <span class="current-page">1</span> / ${totalPages} 页
                </span>
                <button class="btn btn-secondary" onclick="components.changePage('${tableId}', 'next')">
                    <i class="fas fa-chevron-right"></i>
                </button>
            </div>
        `;
    }

    // 🔍 表格过滤
    filterTable(tableId, searchTerm) {
        const table = document.getElementById(tableId);
        const rows = table.querySelectorAll('tbody tr');
        
        rows.forEach(row => {
            const text = row.textContent.toLowerCase();
            row.style.display = text.includes(searchTerm.toLowerCase()) ? '' : 'none';
        });
    }

    // 📑 切换页面
    changePage(tableId, direction) {
        // 实现分页逻辑
        console.log(`Changing page for ${tableId}: ${direction}`);
    }

    // 🔔 显示通知
    showNotification(message, type = 'info', duration = 3000) {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type} fade-in`;
        
        const iconMap = {
            success: 'fa-check-circle',
            error: 'fa-exclamation-circle',
            warning: 'fa-exclamation-triangle',
            info: 'fa-info-circle'
        };

        notification.innerHTML = `
            <div class="flex gap-2">
                <i class="fas ${iconMap[type]}"></i>
                <div>
                    <div class="notification-message">${message}</div>
                </div>
                <button class="notification-close" onclick="this.parentElement.parentElement.remove()">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `;

        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.classList.add('fade-out');
            setTimeout(() => notification.remove(), 300);
        }, duration);
    }

    // 🪟 打开模态框
    openModal(modalId) {
        const modal = document.getElementById(modalId);
        if (!modal) return;
        
        modal.classList.add('modal-open');
        document.body.style.overflow = 'hidden';
        
        // 点击背景关闭
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.closeModal();
            }
        });
    }

    // 🪟 关闭模态框
    closeModal() {
        const modal = document.querySelector('.modal-open');
        if (!modal) return;
        
        modal.classList.remove('modal-open');
        document.body.style.overflow = '';
    }

    // 📊 创建图表容器
    createChartContainer(chartId, title) {
        return `
            <div class="glass-card chart-container">
                <h3 class="chart-title">${title}</h3>
                <div id="${chartId}" class="chart-content"></div>
            </div>
        `;
    }

    // 🔄 创建加载状态
    createLoadingState(message = '加载中...') {
        return `
            <div class="loading-state flex-center">
                <div class="loader"></div>
                <span class="loading-message">${message}</span>
            </div>
        `;
    }

    // 📭 创建空状态
    createEmptyState(message, icon = 'fa-inbox', action = null) {
        const actionHtml = action ? `
            <button class="btn btn-primary mt-3" onclick="${action.onClick}">
                ${action.icon ? `<i class="${action.icon}"></i>` : ''}
                ${action.label}
            </button>
        ` : '';

        return `
            <div class="empty-state flex-center">
                <div class="text-center">
                    <i class="fas ${icon} empty-icon"></i>
                    <p class="empty-message">${message}</p>
                    ${actionHtml}
                </div>
            </div>
        `;
    }

    // 🎯 创建进度条
    createProgressBar(value, max = 100, label = '') {
        const percentage = (value / max) * 100;
        
        return `
            <div class="progress-container">
                ${label ? `<div class="progress-label">${label}</div>` : ''}
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${percentage}%">
                        <span class="progress-value">${value}/${max}</span>
                    </div>
                </div>
            </div>
        `;
    }

    // 🏷️ 创建标签输入
    createTagInput(inputId, tags = []) {
        const tagsHtml = tags.map(tag => `
            <span class="tag-item">
                ${tag}
                <i class="fas fa-times tag-remove" data-tag="${tag}"></i>
            </span>
        `).join('');

        return `
            <div class="tag-input-container" id="${inputId}">
                <div class="tag-list">${tagsHtml}</div>
                <input type="text" class="tag-input" placeholder="输入标签后按回车">
            </div>
        `;
    }

    // 🎨 创建颜色选择器
    createColorPicker(inputId, defaultColor = '#8b5cf6') {
        return `
            <div class="color-picker-container">
                <input type="color" id="${inputId}" value="${defaultColor}" class="color-input">
                <span class="color-preview" style="background: ${defaultColor}"></span>
            </div>
        `;
    }

    // 📤 创建文件上传
    createFileUpload(uploadId, options = {}) {
        const { 
            multiple = false, 
            accept = '*', 
            maxSize = 10 
        } = options;

        return `
            <div class="file-upload-container" id="${uploadId}">
                <input type="file" 
                       class="file-input" 
                       ${multiple ? 'multiple' : ''} 
                       accept="${accept}">
                <div class="file-upload-area">
                    <i class="fas fa-cloud-upload-alt"></i>
                    <p>拖拽文件到此处或点击上传</p>
                    <small>最大文件大小: ${maxSize}MB</small>
                </div>
                <div class="file-list"></div>
            </div>
        `;
    }

    // 🔧 工具函数：获取嵌套值
    getNestedValue(obj, path) {
        return path.split('.').reduce((current, key) => current?.[key], obj);
    }

    // 📐 创建响应式网格
    createGrid(items, columns = 3, renderer) {
        const gridClass = `grid-cols-${columns}`;
        
        return `
            <div class="grid ${gridClass}">
                ${items.map(item => `
                    <div class="grid-item">
                        ${renderer(item)}
                    </div>
                `).join('')}
            </div>
        `;
    }

    // 🎬 添加动画类
    addAnimation(element, animationClass, duration = 500) {
        if (typeof element === 'string') {
            element = document.querySelector(element);
        }
        
        if (!element) return;
        
        element.classList.add(animationClass);
        setTimeout(() => {
            element.classList.remove(animationClass);
        }, duration);
    }

    // 📱 检测设备类型
    getDeviceType() {
        const width = window.innerWidth;
        if (width < 768) return 'mobile';
        if (width < 1024) return 'tablet';
        return 'desktop';
    }

    // 🎨 主题切换
    toggleTheme() {
        const currentTheme = localStorage.getItem('theme') || 'dark';
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        
        document.documentElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
        
        this.showNotification(`已切换到${newTheme === 'dark' ? '深色' : '浅色'}主题`, 'success');
    }

    // 🔧 初始化所有组件
    initializeAll() {
        this.initRippleEffect();
        this.initTooltips();
        this.initModals();
        
        // 添加平滑滚动
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function (e) {
                e.preventDefault();
                const target = document.querySelector(this.getAttribute('href'));
                if (target) {
                    target.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }
            });
        });
        
        // 添加懒加载
        if ('IntersectionObserver' in window) {
            const lazyImages = document.querySelectorAll('img[loading="lazy"]');
            const imageObserver = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const img = entry.target;
                        img.src = img.dataset.src || img.src;
                        imageObserver.unobserve(img);
                    }
                });
            });
            
            lazyImages.forEach(img => imageObserver.observe(img));
        }
    }
}

// 创建全局实例
const components = new NSFWComponents();

// 导出给其他脚本使用
window.NSFWComponents = NSFWComponents;
window.components = components;

// DOM加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    components.initializeAll();
});