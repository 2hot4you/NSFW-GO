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
            
            /* 确保交互按钮在隐私模式下仍然可用 */
            .privacy-overlay-div {
                pointer-events: none !important;
            }
            
            /* 确保下载按钮等交互元素在隐私模式下有更高的层级 */
            button, .download-btn, [onclick] {
                position: relative;
                z-index: 100000 !important;
            }
            
            /* 特别处理悬停显示的按钮 */
            .group:hover button,
            .group:hover .download-btn,
            .movie-card:hover button {
                z-index: 100001 !important;
                pointer-events: auto !important;
            }
        `;
        document.head.appendChild(style);
        console.log('工作模式样式已注入');
    }
    
    // 应用/移除图片蒙版
    function applyImageMasks(enable) {
        const images = document.querySelectorAll('img');
        console.log(`工作模式: ${enable ? '开启' : '关闭'}, 找到 ${images.length} 张图片`);
        
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
                        background: linear-gradient(135deg, 
                            rgba(99, 102, 241, 0.3) 0%, 
                            rgba(168, 85, 247, 0.4) 25%, 
                            rgba(59, 130, 246, 0.3) 50%, 
                            rgba(147, 51, 234, 0.4) 75%, 
                            rgba(79, 70, 229, 0.3) 100%) !important;
                        backdrop-filter: blur(8px) saturate(0.3) !important;
                        z-index: 10 !important;
                        pointer-events: none !important;
                        border-radius: inherit;
                        transition: all 0.3s ease !important;
                    `;
                    
                    // 添加毛玻璃效果和工作模式标识
                    const privacyIcon = document.createElement('div');
                    privacyIcon.style.cssText = `
                        position: absolute !important;
                        top: 50% !important;
                        left: 50% !important;
                        transform: translate(-50%, -50%) !important;
                        color: rgba(255, 255, 255, 0.8) !important;
                        font-size: 24px !important;
                        text-shadow: 0 2px 8px rgba(0, 0, 0, 0.5) !important;
                        z-index: 1 !important;
                        pointer-events: none !important;
                    `;
                    privacyIcon.innerHTML = '<i class="fas fa-briefcase"></i>';
                    overlay.appendChild(privacyIcon);
                    
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
        
        // 确保所有交互按钮在隐私模式下仍然可用
        if (enable) {
            ensureButtonsAccessible();
        }
    }
    
    // 确保按钮在隐私模式下可访问
    function ensureButtonsAccessible() {
        const buttons = document.querySelectorAll('button, .download-btn, [onclick]');
        buttons.forEach(button => {
            if (!button.style.zIndex || parseInt(button.style.zIndex) < 100000) {
                button.style.position = 'relative';
                button.style.zIndex = '100000';
                button.style.pointerEvents = 'auto';
            }
        });
    }
    
    // 监听DOM变化
    function startObserver() {
        if (observer) return;
        
        observer = new MutationObserver((mutations) => {
            if (!isPrivacyMode) return;
            
            mutations.forEach((mutation) => {
                mutation.addedNodes.forEach((node) => {
                    if (node.nodeType === Node.ELEMENT_NODE) {
                        const images = node.querySelectorAll ? node.querySelectorAll('img') : [];
                        images.forEach(img => {
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
                                    background: linear-gradient(135deg, 
                                        rgba(99, 102, 241, 0.3) 0%, 
                                        rgba(168, 85, 247, 0.4) 25%, 
                                        rgba(59, 130, 246, 0.3) 50%, 
                                        rgba(147, 51, 234, 0.4) 75%, 
                                        rgba(79, 70, 229, 0.3) 100%) !important;
                                    backdrop-filter: blur(8px) saturate(0.3) !important;
                                    z-index: 10 !important;
                                    pointer-events: none !important;
                                    border-radius: inherit;
                                    transition: all 0.3s ease !important;
                                `;
                                
                                // 添加毛玻璃效果和工作模式标识
                                const privacyIcon = document.createElement('div');
                                privacyIcon.style.cssText = `
                                    position: absolute !important;
                                    top: 50% !important;
                                    left: 50% !important;
                                    transform: translate(-50%, -50%) !important;
                                    color: rgba(255, 255, 255, 0.8) !important;
                                    font-size: 24px !important;
                                    text-shadow: 0 2px 8px rgba(0, 0, 0, 0.5) !important;
                                    z-index: 1 !important;
                                    pointer-events: none !important;
                                `;
                                privacyIcon.innerHTML = '<i class="fas fa-briefcase"></i>';
                                overlay.appendChild(privacyIcon);
                                
                                img.parentElement.insertBefore(overlay, img.nextSibling);
                                console.log('新图片已添加遮挡层:', img);
                            }, 100);
                        });
                        
                        // 确保新添加的按钮也可访问
                        const newButtons = node.querySelectorAll ? node.querySelectorAll('button, .download-btn, [onclick]') : [];
                        newButtons.forEach(button => {
                            button.style.position = 'relative';
                            button.style.zIndex = '100000';
                            button.style.pointerEvents = 'auto';
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
        console.log(`工作模式切换为: ${isPrivacyMode ? '开启' : '关闭'}`);
        
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
            btn.innerHTML = '<i class="fas fa-eye mr-2"></i><span class="hidden sm:inline">工作模式</span>';
            btn.style.backgroundColor = 'rgba(99, 102, 241, 0.2)';
            btn.style.color = '#818cf8';
            btn.title = '工作模式已开启，内容已模糊处理，但功能按钮仍可正常使用';
        } else {
            btn.innerHTML = '<i class="fas fa-briefcase mr-2"></i><span class="hidden sm:inline">工作模式</span>';
            btn.style.backgroundColor = 'rgba(75, 85, 99, 0.5)';
            btn.style.color = '#d1d5db';
            btn.title = '开启工作模式，在办公环境中安全使用';
        }
    }
    
    // 创建隐私模式按钮
    function createPrivacyButton() {
        const btn = document.createElement('button');
        btn.id = 'privacy-mode-btn';
        btn.className = 'privacy-btn px-3 py-2 rounded-lg text-sm font-medium transition-all duration-200';
        btn.title = '开启工作模式，在办公环境中安全使用';
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
        console.log('初始化工作模式，当前状态:', isPrivacyMode);
        
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