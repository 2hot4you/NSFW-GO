// 全局图片隐私蒙版功能
(function() {
    'use strict';
    
    let isPrivacyMode = false;
    let observer = null;
    
    // 创建CSS样式
    function injectStyles() {
        if (document.getElementById('privacy-mode-styles')) return;
        
        const style = document.createElement('style');
        style.id = 'privacy-mode-styles';
        style.textContent = `
            .privacy-mask-overlay {
                position: relative !important;
            }
            
            .privacy-mask-overlay::after {
                content: '';
                position: absolute !important;
                top: 0 !important;
                left: 0 !important;
                right: 0 !important;
                bottom: 0 !important;
                width: 100% !important;
                height: 100% !important;
                background: rgba(0, 0, 0, 0.8) !important;
                backdrop-filter: blur(12px) !important;
                z-index: 99999 !important;
                pointer-events: none !important;
                border-radius: inherit;
                display: block !important;
            }
            
            .privacy-btn {
                transition: all 0.2s ease;
            }
            
            .privacy-btn:hover {
                background-color: rgba(75, 85, 99, 0.8) !important;
            }
        `;
        document.head.appendChild(style);
        console.log('隐私模式样式已注入');
    }
    
    // 应用/移除图片蒙版
    function applyImageMasks(enable) {
        const images = document.querySelectorAll('img');
        console.log(`隐私模式: ${enable ? '开启' : '关闭'}, 找到 ${images.length} 张图片`);
        
        images.forEach((img, index) => {
            if (enable) {
                // 检查是否已经有遮挡层
                if (!img.nextElementSibling || !img.nextElementSibling.classList.contains('privacy-overlay-div')) {
                    // 确保图片有相对定位的父容器
                    if (img.parentElement.style.position !== 'relative') {
                        img.parentElement.style.position = 'relative';
                    }
                    
                    // 创建遮挡div
                    const overlay = document.createElement('div');
                    overlay.className = 'privacy-overlay-div';
                    overlay.style.cssText = `
                        position: absolute !important;
                        top: 0 !important;
                        left: 0 !important;
                        right: 0 !important;
                        bottom: 0 !important;
                        width: 100% !important;
                        height: 100% !important;
                        background: rgba(0, 0, 0, 0.85) !important;
                        backdrop-filter: blur(15px) !important;
                        z-index: 99999 !important;
                        pointer-events: none !important;
                        border-radius: inherit;
                    `;
                    
                    // 将遮挡层插入到图片后面
                    img.parentElement.insertBefore(overlay, img.nextSibling);
                    console.log(`图片 ${index + 1}: 已添加遮挡层 - ${img.src || img.alt || '未知'}`);
                }
            } else {
                // 移除遮挡层
                const overlay = img.nextElementSibling;
                if (overlay && overlay.classList.contains('privacy-overlay-div')) {
                    overlay.remove();
                    console.log(`图片 ${index + 1}: 已移除遮挡层 - ${img.src || img.alt || '未知'}`);
                }
            }
        });
    }
    
    // 监听DOM变化，自动为新图片添加蒙版
    function startObserver() {
        if (observer) observer.disconnect();
        
        observer = new MutationObserver((mutations) => {
            if (!isPrivacyMode) return;
            
            mutations.forEach((mutation) => {
                mutation.addedNodes.forEach((node) => {
                    if (node.nodeType === Node.ELEMENT_NODE) {
                        // 检查新添加的图片
                        const images = node.tagName === 'IMG' ? [node] : node.querySelectorAll('img');
                        images.forEach(img => {
                            if (!img.nextElementSibling || !img.nextElementSibling.classList.contains('privacy-overlay-div')) {
                                // 为新图片添加遮挡层
                                setTimeout(() => {
                                    if (img.parentElement.style.position !== 'relative') {
                                        img.parentElement.style.position = 'relative';
                                    }
                                    
                                    const overlay = document.createElement('div');
                                    overlay.className = 'privacy-overlay-div';
                                    overlay.style.cssText = `
                                        position: absolute !important;
                                        top: 0 !important;
                                        left: 0 !important;
                                        right: 0 !important;
                                        bottom: 0 !important;
                                        width: 100% !important;
                                        height: 100% !important;
                                        background: rgba(0, 0, 0, 0.85) !important;
                                        backdrop-filter: blur(15px) !important;
                                        z-index: 99999 !important;
                                        pointer-events: none !important;
                                        border-radius: inherit;
                                    `;
                                    
                                    img.parentElement.insertBefore(overlay, img.nextSibling);
                                    console.log('新图片已添加遮挡层:', img);
                                }, 100);
                            }
                        });
                    }
                });
            });
        });
        
        observer.observe(document.body, {
            childList: true,
            subtree: true
        });
    }
    
    // 切换隐私模式
    function togglePrivacyMode() {
        isPrivacyMode = !isPrivacyMode;
        console.log(`隐私模式切换为: ${isPrivacyMode ? '开启' : '关闭'}`);
        
        localStorage.setItem('privacy-mode', isPrivacyMode ? 'on' : 'off');
        
        applyImageMasks(isPrivacyMode);
        updateButtonState();
        
        if (isPrivacyMode) {
            startObserver();
        } else if (observer) {
            observer.disconnect();
        }
    }
    
    // 更新按钮状态
    function updateButtonState() {
        const btn = document.getElementById('privacy-mode-btn');
        if (!btn) return;
        
        if (isPrivacyMode) {
            btn.innerHTML = '<i class="fas fa-eye mr-2"></i><span class="hidden sm:inline">关闭隐私</span>';
            btn.style.backgroundColor = 'rgba(239, 68, 68, 0.2)';
            btn.style.color = '#f87171';
        } else {
            btn.innerHTML = '<i class="fas fa-eye-slash mr-2"></i><span class="hidden sm:inline">隐私模式</span>';
            btn.style.backgroundColor = 'rgba(75, 85, 99, 0.5)';
            btn.style.color = '#d1d5db';
        }
    }
    
    // 创建隐私模式按钮
    function createPrivacyButton() {
        const btn = document.createElement('button');
        btn.id = 'privacy-mode-btn';
        btn.className = 'privacy-btn px-3 py-2 rounded-lg text-sm font-medium transition-all duration-200';
        btn.title = '在公共场所使用时，开启隐私模式可为图片添加蒙版';
        btn.onclick = togglePrivacyMode;
        
        return btn;
    }
    
    // 插入按钮到导航栏
    function insertPrivacyButton() {
        // 避免重复插入
        if (document.getElementById('privacy-mode-btn')) return;
        
        // 查找导航栏容器（兼容多种结构）
        const navSelectors = [
            'nav .flex.items-center.space-x-4',
            'nav .flex.items-center.space-x-6', 
            '#nav-right-privacy-anchor',
            'nav .max-w-7xl .flex:last-child',
            'nav [class*="space-x"]'
        ];
        
        let targetContainer = null;
        for (const selector of navSelectors) {
            targetContainer = document.querySelector(selector);
            if (targetContainer) break;
        }
        
        // 如果找不到合适容器，直接插入到nav中
        if (!targetContainer) {
            targetContainer = document.querySelector('nav');
            if (!targetContainer) return;
        }
        
        const btn = createPrivacyButton();
        targetContainer.appendChild(btn);
        updateButtonState();
    }
    
    // 初始化隐私模式
    function initPrivacyMode() {
        // 注入样式
        injectStyles();
        
        // 读取保存的状态
        isPrivacyMode = localStorage.getItem('privacy-mode') === 'on';
        console.log('初始化隐私模式，当前状态:', isPrivacyMode);
        
        // 插入按钮
        insertPrivacyButton();
        
        // 延迟应用初始状态，等待页面图片加载
        setTimeout(() => {
            if (isPrivacyMode) {
                applyImageMasks(true);
                startObserver();
            }
            updateButtonState();
        }, 1000);
        
        // 再次延迟检查，确保所有异步图片都被处理
        setTimeout(() => {
            if (isPrivacyMode) {
                applyImageMasks(true);
            }
        }, 3000);
    }
    
    // 页面加载完成后初始化
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initPrivacyMode);
    } else {
        initPrivacyMode();
    }
    
    // 导出到全局，便于调试
    window.PrivacyMode = {
        toggle: togglePrivacyMode,
        isEnabled: () => isPrivacyMode,
        refresh: () => {
            applyImageMasks(isPrivacyMode);
            updateButtonState();
        }
    };
})(); 