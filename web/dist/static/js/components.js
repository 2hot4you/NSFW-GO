/**
 * NSFW-GO Shared Components
 * 适配 Cinematic Minimalism 设计系统
 */

const components = {
    // 🔔 通知消息
    showNotification(message, type = 'info') {
        const container = document.getElementById('notification-container') || this.createNotificationContainer();

        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;

        const iconMap = {
            success: 'fa-check-circle',
            error: 'fa-times-circle',
            warning: 'fa-exclamation-triangle',
            info: 'fa-info-circle'
        };

        notification.innerHTML = `
            <div class="notification-icon">
                <i class="fas ${iconMap[type] || 'fa-info-circle'}"></i>
            </div>
            <div class="notification-content">${message}</div>
            <button class="notification-close" onclick="this.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        `;

        // 样式注入 (如果尚未存在)
        if (!document.getElementById('notification-styles')) {
            const style = document.createElement('style');
            style.id = 'notification-styles';
            style.textContent = `
                #notification-container {
                    position: fixed;
                    top: 20px;
                    right: 20px;
                    z-index: 9999;
                    display: flex;
                    flex-direction: column;
                    gap: 10px;
                }
                .notification {
                    background: var(--bg-card);
                    border: 1px solid var(--border-subtle);
                    border-left: 4px solid var(--primary);
                    border-radius: var(--radius-md);
                    padding: 1rem;
                    min-width: 300px;
                    display: flex;
                    align-items: center;
                    gap: 12px;
                    box-shadow: var(--shadow-elevated);
                    animation: slideInRight 0.3s ease-out;
                    color: var(--text-main);
                }
                .notification-success { border-left-color: var(--success); }
                .notification-error { border-left-color: var(--danger); }
                .notification-warning { border-left-color: var(--warning); }
                
                .notification-icon { font-size: 1.25rem; }
                .notification-success .notification-icon { color: var(--success); }
                .notification-error .notification-icon { color: var(--danger); }
                .notification-warning .notification-icon { color: var(--warning); }
                
                .notification-content { flex: 1; font-size: 0.875rem; }
                
                .notification-close {
                    background: none;
                    border: none;
                    color: var(--text-muted);
                    cursor: pointer;
                    padding: 4px;
                }
                .notification-close:hover { color: var(--text-main); }
                
                @keyframes slideInRight {
                    from { transform: translateX(100%); opacity: 0; }
                    to { transform: translateX(0); opacity: 1; }
                }
            `;
            document.head.appendChild(style);
        }

        container.appendChild(notification);

        // 自动移除
        setTimeout(() => {
            notification.style.opacity = '0';
            notification.style.transform = 'translateX(100%)';
            notification.style.transition = 'all 0.3s ease';
            setTimeout(() => notification.remove(), 300);
        }, 3000);
    },

    createNotificationContainer() {
        const container = document.createElement('div');
        container.id = 'notification-container';
        document.body.appendChild(container);
        return container;
    },

    // 🖼️ 模态框
    createModal(id, title, contentHtml) {
        const modal = document.createElement('div');
        modal.className = 'modal-overlay';
        modal.id = id;

        modal.innerHTML = `
            <div class="modal-content">
                <div class="flex items-center justify-between mb-4 pb-4 border-b border-subtle">
                    <h3 class="text-lg font-bold">${title}</h3>
                    <button class="text-muted hover:text-white" onclick="components.closeModal('${id}')">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="modal-body">
                    ${contentHtml}
                </div>
            </div>
        `;

        document.body.appendChild(modal);
        return modal;
    },

    openModal(id) {
        const modal = document.getElementById(id);
        if (modal) {
            modal.classList.add('active');
        }
    },

    closeModal(id) {
        const modal = document.getElementById(id);
        if (modal) {
            modal.classList.remove('active');
            // 可选：延迟移除 DOM
        }
    },

    // 📊 加载状态
    createLoadingState(message = '加载中...') {
        return `
            <div class="flex flex-col items-center justify-center py-12 text-muted">
                <i class="fas fa-spinner fa-spin text-3xl mb-4"></i>
                <p>${message}</p>
            </div>
        `;
    },

    // 📭 空状态
    createEmptyState(message, icon = 'fa-inbox', action = null) {
        return `
            <div class="flex flex-col items-center justify-center py-12 text-muted">
                <div class="w-16 h-16 bg-input rounded-full flex items-center justify-center mb-4">
                    <i class="fas ${icon} text-2xl"></i>
                </div>
                <p class="mb-4">${message}</p>
                ${action ? `
                    <button class="btn btn-primary" onclick="${action.onClick}">
                        <i class="${action.icon}"></i> ${action.label}
                    </button>
                ` : ''}
            </div>
        `;
    }
};

// 导出到全局
window.components = components;