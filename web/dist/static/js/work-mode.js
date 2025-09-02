// 工作模式导航栏组件
(function() {
    'use strict';
    
    // 等待DOM加载
    function init() {
        console.log('初始化工作模式导航栏...');
        
        // 获取当前工作模式状态
        const workMode = localStorage.getItem('workMode') === 'true';
        
        // 创建导航栏HTML
        const navHTML = `
            <nav id="workModeNav" style="
                background: linear-gradient(90deg, #1f2937 0%, #374151 100%);
                padding: 1rem;
                position: sticky;
                top: 0;
                z-index: 9999;
                box-shadow: 0 2px 4px rgba(0,0,0,0.1);
                border-bottom: 1px solid rgba(255,255,255,0.1);
            ">
                <div style="
                    max-width: 1280px;
                    margin: 0 auto;
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                ">
                    <!-- 左侧logo和导航链接 -->
                    <div style="display: flex; align-items: center; gap: 2rem;">
                        <a href="/" style="
                            color: white;
                            font-size: 1.5rem;
                            font-weight: bold;
                            text-decoration: none;
                            display: flex;
                            align-items: center;
                            gap: 0.5rem;
                        ">
                            <span style="font-size: 1.5rem;">🎬</span>
                            NSFW-Go
                        </a>
                        
                        <!-- 导航链接 -->
                        <div style="display: flex; gap: 1rem;">
                            <a href="/search.html" style="color: #9ca3af; text-decoration: none; padding: 0.5rem;">搜索</a>
                            <a href="/local-movies.html" style="color: #9ca3af; text-decoration: none; padding: 0.5rem;">本地库</a>
                            <a href="/rankings.html" style="color: #9ca3af; text-decoration: none; padding: 0.5rem;">排行榜</a>
                            <a href="/downloads.html" style="color: #9ca3af; text-decoration: none; padding: 0.5rem;">下载</a>
                        </div>
                    </div>
                    
                    <!-- 右侧工作模式按钮 -->
                    <div style="display: flex; gap: 1rem; align-items: center;">
                        <!-- 工作模式切换按钮 -->
                        <button 
                            id="workModeToggleBtn"
                            onclick="window.toggleWorkMode()"
                            style="
                                background: ${workMode ? 'linear-gradient(90deg, #10b981 0%, #059669 100%)' : 'linear-gradient(90deg, #6b7280 0%, #4b5563 100%)'};
                                color: white;
                                padding: 0.5rem 1.5rem;
                                border-radius: 0.5rem;
                                border: none;
                                cursor: pointer;
                                font-size: 1rem;
                                display: flex;
                                align-items: center;
                                gap: 0.5rem;
                                transition: all 0.3s;
                                box-shadow: 0 2px 4px rgba(0,0,0,0.2);
                            "
                            onmouseover="this.style.transform='scale(1.05)'"
                            onmouseout="this.style.transform='scale(1)'"
                        >
                            <span style="font-size: 1.2rem;">${workMode ? '🙈' : '👁️'}</span>
                            <span>${workMode ? '工作模式：开' : '工作模式：关'}</span>
                        </button>
                        
                        <!-- 设置按钮 -->
                        <a href="/config.html" style="
                            background: linear-gradient(90deg, #6b7280 0%, #4b5563 100%);
                            color: white;
                            padding: 0.5rem 1rem;
                            border-radius: 0.5rem;
                            text-decoration: none;
                            display: flex;
                            align-items: center;
                            gap: 0.5rem;
                            box-shadow: 0 2px 4px rgba(0,0,0,0.2);
                        ">
                            <span>⚙️</span>
                            <span>设置</span>
                        </a>
                    </div>
                </div>
            </nav>
        `;
        
        // 插入导航栏到页面顶部
        document.body.insertAdjacentHTML('afterbegin', navHTML);
        
        // 应用工作模式样式
        applyWorkModeStyles(workMode);
        
        // 显示工作模式状态提示
        if (workMode) {
            showWorkModeIndicator();
        }
        
        console.log('工作模式导航栏初始化完成！');
    }
    
    // 切换工作模式
    function toggleWorkMode() {
        const currentMode = localStorage.getItem('workMode') === 'true';
        const newMode = !currentMode;
        
        // 保存状态
        localStorage.setItem('workMode', String(newMode));
        
        // 更新按钮
        const btn = document.getElementById('workModeToggleBtn');
        if (btn) {
            btn.style.background = newMode 
                ? 'linear-gradient(90deg, #10b981 0%, #059669 100%)' 
                : 'linear-gradient(90deg, #6b7280 0%, #4b5563 100%)';
            btn.innerHTML = `
                <span style="font-size: 1.2rem;">${newMode ? '🙈' : '👁️'}</span>
                <span>${newMode ? '工作模式：开' : '工作模式：关'}</span>
            `;
        }
        
        // 应用样式
        applyWorkModeStyles(newMode);
        
        // 显示/隐藏状态提示
        if (newMode) {
            showWorkModeIndicator();
        } else {
            hideWorkModeIndicator();
        }
        
        // 触发自定义事件
        window.dispatchEvent(new CustomEvent('workModeChanged', { 
            detail: { enabled: newMode } 
        }));
    }
    
    // 应用工作模式样式
    function applyWorkModeStyles(enabled) {
        let styleEl = document.getElementById('work-mode-styles');
        
        if (enabled) {
            if (!styleEl) {
                styleEl = document.createElement('style');
                styleEl.id = 'work-mode-styles';
                document.head.appendChild(styleEl);
            }
            
            styleEl.textContent = `
                /* 工作模式 - 模糊所有图片 */
                .work-mode img,
                .work-mode .movie-cover,
                .work-mode .cover-image,
                .work-mode video {
                    filter: blur(25px) brightness(0.3) !important;
                    transition: filter 0.3s;
                }
                
                /* 鼠标悬停时稍微减少模糊 */
                .work-mode img:hover,
                .work-mode .movie-cover:hover,
                .work-mode .cover-image:hover {
                    filter: blur(15px) brightness(0.5) !important;
                }
                
                /* 导航栏logo不模糊 */
                #workModeNav img {
                    filter: none !important;
                }
            `;
            
            document.body.classList.add('work-mode');
        } else {
            if (styleEl) {
                styleEl.remove();
            }
            document.body.classList.remove('work-mode');
        }
    }
    
    // 显示工作模式指示器
    function showWorkModeIndicator() {
        // 移除旧的指示器
        hideWorkModeIndicator();
        
        // 创建新的指示器
        const indicator = document.createElement('div');
        indicator.id = 'work-mode-indicator';
        indicator.style.cssText = `
            position: fixed;
            bottom: 20px;
            left: 20px;
            background: linear-gradient(90deg, #10b981 0%, #059669 100%);
            color: white;
            padding: 10px 20px;
            border-radius: 50px;
            font-size: 14px;
            z-index: 10000;
            box-shadow: 0 4px 6px rgba(0,0,0,0.3);
            display: flex;
            align-items: center;
            gap: 8px;
            animation: pulse 2s infinite;
        `;
        indicator.innerHTML = '🙈 工作模式已开启';
        
        // 添加动画样式
        const style = document.createElement('style');
        style.textContent = `
            @keyframes pulse {
                0%, 100% { opacity: 1; transform: scale(1); }
                50% { opacity: 0.8; transform: scale(0.95); }
            }
        `;
        document.head.appendChild(style);
        
        document.body.appendChild(indicator);
    }
    
    // 隐藏工作模式指示器
    function hideWorkModeIndicator() {
        const indicator = document.getElementById('work-mode-indicator');
        if (indicator) {
            indicator.remove();
        }
    }
    
    // 导出全局函数
    window.toggleWorkMode = toggleWorkMode;
    
    // 初始化
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();