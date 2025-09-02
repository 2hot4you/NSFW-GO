// ⚡ NSFW-GO 性能优化模块
// 提供图片懒加载、虚拟滚动、资源预加载等性能优化功能

class PerformanceOptimizer {
    constructor() {
        this.init();
    }

    init() {
        // 初始化懒加载
        this.initLazyLoading();
        // 初始化预加载
        this.initPrefetch();
        // 初始化防抖和节流
        this.initDebounceThrottle();
        // 优化动画性能
        this.optimizeAnimations();
    }

    // 🖼️ 图片懒加载
    initLazyLoading() {
        if ('IntersectionObserver' in window) {
            const imageObserver = new IntersectionObserver((entries, observer) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const img = entry.target;
                        
                        // 添加加载动画
                        img.classList.add('loading');
                        
                        // 加载图片
                        if (img.dataset.src) {
                            img.src = img.dataset.src;
                            img.removeAttribute('data-src');
                        }
                        
                        // 加载完成后移除观察
                        img.onload = () => {
                            img.classList.remove('loading');
                            img.classList.add('loaded');
                            observer.unobserve(img);
                        };
                        
                        // 加载失败处理
                        img.onerror = () => {
                            img.classList.remove('loading');
                            img.classList.add('error');
                            img.src = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="400" height="300"%3E%3Crect width="400" height="300" fill="%23f3f4f6"/%3E%3Ctext x="50%25" y="50%25" dominant-baseline="middle" text-anchor="middle" font-family="Arial" font-size="20" fill="%239ca3af"%3E📷 加载失败%3C/text%3E%3C/svg%3E';
                            observer.unobserve(img);
                        };
                    }
                });
            }, {
                rootMargin: '50px 0px',
                threshold: 0.01
            });

            // 观察所有懒加载图片
            document.querySelectorAll('img[data-src]').forEach(img => {
                imageObserver.observe(img);
            });

            // 监听DOM变化，自动处理新增的图片
            const mutationObserver = new MutationObserver(() => {
                document.querySelectorAll('img[data-src]:not(.observed)').forEach(img => {
                    img.classList.add('observed');
                    imageObserver.observe(img);
                });
            });

            mutationObserver.observe(document.body, {
                childList: true,
                subtree: true
            });
        } else {
            // 降级处理：直接加载所有图片
            document.querySelectorAll('img[data-src]').forEach(img => {
                img.src = img.dataset.src;
            });
        }
    }

    // 🔗 预加载关键资源
    initPrefetch() {
        // 预连接到API服务器
        this.addLinkTag('preconnect', window.location.origin);
        
        // 预加载字体
        this.addLinkTag('preload', 'https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap', 'style');
        
        // 预加载下一页可能用到的资源
        const currentPage = window.location.pathname;
        const nextPages = this.getNextPages(currentPage);
        
        nextPages.forEach(page => {
            this.addLinkTag('prefetch', page);
        });
    }

    // 获取可能访问的下一页
    getNextPages(currentPage) {
        const pages = {
            '/': ['/search.html', '/local-movies.html'],
            '/search.html': ['/downloads.html', '/local-movies.html'],
            '/local-movies.html': ['/search.html', '/rankings.html'],
            '/rankings.html': ['/search.html', '/downloads.html'],
            '/downloads.html': ['/search.html', '/config.html']
        };
        
        return pages[currentPage] || [];
    }

    // 添加link标签
    addLinkTag(rel, href, as = null) {
        const link = document.createElement('link');
        link.rel = rel;
        link.href = href;
        if (as) link.as = as;
        document.head.appendChild(link);
    }

    // ⏱️ 防抖和节流工具
    initDebounceThrottle() {
        // 防抖函数
        window.debounce = (func, wait = 300) => {
            let timeout;
            return function executedFunction(...args) {
                const later = () => {
                    clearTimeout(timeout);
                    func(...args);
                };
                clearTimeout(timeout);
                timeout = setTimeout(later, wait);
            };
        };

        // 节流函数
        window.throttle = (func, limit = 100) => {
            let inThrottle;
            return function(...args) {
                if (!inThrottle) {
                    func.apply(this, args);
                    inThrottle = true;
                    setTimeout(() => inThrottle = false, limit);
                }
            };
        };

        // 应用到常见事件
        this.applyOptimizations();
    }

    // 应用优化
    applyOptimizations() {
        // 优化滚动事件
        const scrollHandler = window.throttle(() => {
            // 显示/隐藏返回顶部按钮
            const backToTop = document.getElementById('backToTop');
            if (backToTop) {
                if (window.scrollY > 300) {
                    backToTop.classList.remove('hidden');
                } else {
                    backToTop.classList.add('hidden');
                }
            }
            
            // 触发懒加载检查
            this.checkLazyLoad();
        }, 100);

        window.addEventListener('scroll', scrollHandler, { passive: true });

        // 优化窗口大小调整
        const resizeHandler = window.debounce(() => {
            // 重新计算布局
            this.recalculateLayout();
        }, 300);

        window.addEventListener('resize', resizeHandler);

        // 优化输入框
        document.querySelectorAll('input[type="search"], input[type="text"]').forEach(input => {
            const searchHandler = window.debounce((e) => {
                const value = e.target.value;
                if (value.length >= 2) {
                    this.triggerSearch(value);
                }
            }, 500);

            input.addEventListener('input', searchHandler);
        });
    }

    // 🎬 优化动画性能
    optimizeAnimations() {
        // 使用 CSS will-change 优化动画元素
        document.querySelectorAll('.hover-scale, .float-animation, .pulse-animation').forEach(el => {
            el.style.willChange = 'transform';
        });

        // 使用 requestAnimationFrame 优化JavaScript动画
        window.smoothScroll = (targetY, duration = 500) => {
            const startY = window.scrollY;
            const distance = targetY - startY;
            const startTime = performance.now();

            const animation = (currentTime) => {
                const elapsed = currentTime - startTime;
                const progress = Math.min(elapsed / duration, 1);
                const ease = this.easeInOutCubic(progress);
                
                window.scrollTo(0, startY + distance * ease);
                
                if (progress < 1) {
                    requestAnimationFrame(animation);
                }
            };

            requestAnimationFrame(animation);
        };

        // 优化页面切换动画
        this.optimizePageTransitions();
    }

    // 缓动函数
    easeInOutCubic(t) {
        return t < 0.5 ? 4 * t * t * t : (t - 1) * (2 * t - 2) * (2 * t - 2) + 1;
    }

    // 优化页面切换
    optimizePageTransitions() {
        document.addEventListener('click', (e) => {
            const link = e.target.closest('a[href^="/"]');
            if (link && !link.target && !e.ctrlKey && !e.metaKey) {
                e.preventDefault();
                const href = link.getAttribute('href');
                
                // 添加退出动画
                document.body.style.opacity = '0';
                document.body.style.transition = 'opacity 0.3s';
                
                setTimeout(() => {
                    window.location.href = href;
                }, 300);
            }
        });

        // 页面加载时添加进入动画
        window.addEventListener('load', () => {
            document.body.style.opacity = '0';
            document.body.style.transition = 'opacity 0.3s';
            setTimeout(() => {
                document.body.style.opacity = '1';
            }, 10);
        });
    }

    // 检查懒加载
    checkLazyLoad() {
        // 触发 IntersectionObserver 重新检查
        document.querySelectorAll('img[data-src]').forEach(img => {
            const rect = img.getBoundingClientRect();
            if (rect.top < window.innerHeight + 50 && rect.bottom > -50) {
                if (img.dataset.src) {
                    img.src = img.dataset.src;
                    img.removeAttribute('data-src');
                }
            }
        });
    }

    // 重新计算布局
    recalculateLayout() {
        // 根据窗口大小调整网格布局
        const width = window.innerWidth;
        const grids = document.querySelectorAll('.movie-grid');
        
        grids.forEach(grid => {
            if (width < 640) {
                grid.style.gridTemplateColumns = 'repeat(auto-fill, minmax(150px, 1fr))';
            } else if (width < 1024) {
                grid.style.gridTemplateColumns = 'repeat(auto-fill, minmax(200px, 1fr))';
            } else {
                grid.style.gridTemplateColumns = 'repeat(auto-fill, minmax(250px, 1fr))';
            }
        });
    }

    // 触发搜索（示例）
    triggerSearch(value) {
        // 这里可以实现实时搜索建议等功能
        console.log('🔍 搜索:', value);
    }

    // 📊 性能监控
    monitorPerformance() {
        if ('PerformanceObserver' in window) {
            // 监控长任务
            const observer = new PerformanceObserver((list) => {
                for (const entry of list.getEntries()) {
                    if (entry.duration > 50) {
                        console.warn('⚠️ 长任务检测:', entry);
                    }
                }
            });
            
            observer.observe({ entryTypes: ['longtask'] });

            // 监控页面加载性能
            window.addEventListener('load', () => {
                const perfData = performance.getEntriesByType('navigation')[0];
                if (perfData) {
                    console.log('📊 页面加载性能:', {
                        'DNS查询': `${Math.round(perfData.domainLookupEnd - perfData.domainLookupStart)}ms`,
                        'TCP连接': `${Math.round(perfData.connectEnd - perfData.connectStart)}ms`,
                        'HTTP请求': `${Math.round(perfData.responseEnd - perfData.requestStart)}ms`,
                        'DOM解析': `${Math.round(perfData.domInteractive - perfData.responseEnd)}ms`,
                        '总加载时间': `${Math.round(perfData.loadEventEnd - perfData.fetchStart)}ms`
                    });
                }
            });
        }
    }

    // 🗑️ 内存管理
    cleanupMemory() {
        // 清理未使用的DOM引用
        const images = document.querySelectorAll('img.loaded');
        images.forEach(img => {
            if (!this.isInViewport(img)) {
                img.removeAttribute('src');
                img.classList.remove('loaded');
            }
        });

        // 清理事件监听器
        this.cleanupEventListeners();
    }

    // 检查元素是否在视口中
    isInViewport(element) {
        const rect = element.getBoundingClientRect();
        return (
            rect.top >= -100 &&
            rect.left >= -100 &&
            rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) + 100 &&
            rect.right <= (window.innerWidth || document.documentElement.clientWidth) + 100
        );
    }

    // 清理事件监听器
    cleanupEventListeners() {
        // 实现事件监听器的清理逻辑
    }
}

// 创建全局实例
const performanceOptimizer = new PerformanceOptimizer();

// 导出给其他脚本使用
window.PerformanceOptimizer = PerformanceOptimizer;
window.performanceOptimizer = performanceOptimizer;

// 开启性能监控（仅在开发环境）
if (window.location.hostname === 'localhost') {
    performanceOptimizer.monitorPerformance();
}