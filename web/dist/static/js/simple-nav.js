// 简化版导航栏 - 用于测试
document.addEventListener('DOMContentLoaded', function() {
    console.log('Simple nav loading...');
    
    const workMode = localStorage.getItem('workMode') === 'true';
    
    const navHTML = `
        <nav style="background: #1f2937; padding: 1rem; position: sticky; top: 0; z-index: 50;">
            <div style="max-width: 1280px; margin: 0 auto; display: flex; justify-content: space-between; align-items: center;">
                <div style="display: flex; align-items: center; gap: 1rem;">
                    <a href="/" style="color: white; font-size: 1.5rem; font-weight: bold; text-decoration: none;">
                        NSFW-Go
                    </a>
                </div>
                
                <div style="display: flex; gap: 1rem; align-items: center;">
                    <button 
                        id="workModeBtn"
                        onclick="toggleWorkMode()"
                        style="background: ${workMode ? '#10b981' : '#6b7280'}; color: white; padding: 0.5rem 1rem; border-radius: 0.5rem; border: none; cursor: pointer;"
                    >
                        ${workMode ? '🙈 工作模式开启' : '👁️ 工作模式关闭'}
                    </button>
                    
                    <a href="/config.html" style="background: #6b7280; color: white; padding: 0.5rem 1rem; border-radius: 0.5rem; text-decoration: none;">
                        ⚙️ 设置
                    </a>
                </div>
            </div>
        </nav>
    `;
    
    document.body.insertAdjacentHTML('afterbegin', navHTML);
    
    // 应用工作模式样式
    if (workMode) {
        applyWorkMode();
    }
    
    console.log('Simple nav loaded!');
});

function toggleWorkMode() {
    const currentMode = localStorage.getItem('workMode') === 'true';
    const newMode = !currentMode;
    localStorage.setItem('workMode', String(newMode));
    
    const btn = document.getElementById('workModeBtn');
    if (btn) {
        btn.style.background = newMode ? '#10b981' : '#6b7280';
        btn.textContent = newMode ? '🙈 工作模式开启' : '👁️ 工作模式关闭';
    }
    
    if (newMode) {
        applyWorkMode();
    } else {
        removeWorkMode();
    }
}

function applyWorkMode() {
    let styleEl = document.getElementById('work-mode-styles');
    if (!styleEl) {
        styleEl = document.createElement('style');
        styleEl.id = 'work-mode-styles';
        document.head.appendChild(styleEl);
    }
    styleEl.textContent = `
        body img {
            filter: blur(20px) brightness(0.5) !important;
        }
        body img:hover {
            filter: blur(10px) brightness(0.7) !important;
        }
    `;
    document.body.classList.add('work-mode');
}

function removeWorkMode() {
    const styleEl = document.getElementById('work-mode-styles');
    if (styleEl) {
        styleEl.remove();
    }
    document.body.classList.remove('work-mode');
}

window.toggleWorkMode = toggleWorkMode;