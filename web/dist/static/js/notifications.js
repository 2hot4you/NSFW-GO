// 通知系统
function showNotification(message, type = 'info', duration = 5000) {
    const container = document.getElementById('notification-container');
    if (!container) return;
    
    // 创建通知元素
    const notification = document.createElement('div');
    notification.className = `notification p-4 rounded-lg shadow-lg border-l-4 ${getNotificationClass(type)}`;
    
    notification.innerHTML = `
        <div class="flex items-center justify-between">
            <div class="flex items-center">
                <i class="fas ${getNotificationIcon(type)} mr-3"></i>
                <span>${message}</span>
            </div>
            <button class="ml-4 text-gray-400 hover:text-gray-600" onclick="this.parentElement.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        </div>
    `;
    
    container.appendChild(notification);
    
    // 显示动画
    setTimeout(() => {
        notification.classList.add('show');
    }, 100);
    
    // 自动隐藏
    if (duration > 0) {
        setTimeout(() => {
            hideNotification(notification);
        }, duration);
    }
    
    return notification;
}

function hideNotification(notification) {
    notification.classList.remove('show');
    setTimeout(() => {
        if (notification.parentElement) {
            notification.parentElement.removeChild(notification);
        }
    }, 300);
}

function getNotificationClass(type) {
    const classes = {
        'success': 'bg-green-50 border-green-400 text-green-800',
        'error': 'bg-red-50 border-red-400 text-red-800',
        'warning': 'bg-yellow-50 border-yellow-400 text-yellow-800',
        'info': 'bg-blue-50 border-blue-400 text-blue-800'
    };
    return classes[type] || classes.info;
}

function getNotificationIcon(type) {
    const icons = {
        'success': 'fa-check-circle',
        'error': 'fa-exclamation-circle',
        'warning': 'fa-exclamation-triangle',
        'info': 'fa-info-circle'
    };
    return icons[type] || icons.info;
} 