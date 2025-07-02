// 搜索页面主要功能
class SearchPage {
    constructor() {
        this.searchInput = document.getElementById('search-input');
        this.searchBtn = document.getElementById('search-btn');
        this.suggestionsDropdown = document.getElementById('suggestions-dropdown');
        this.searchResults = document.getElementById('search-results');
        this.emptyState = document.getElementById('empty-state');
        this.noResults = document.getElementById('no-results');
        this.loadingState = document.getElementById('loading-state');
        
        this.searchLocalCheckbox = document.getElementById('search-local');
        this.searchRankingsCheckbox = document.getElementById('search-rankings');
        this.searchJavdbCheckbox = document.getElementById('search-javdb');
        this.javdbOptions = document.getElementById('javdb-options');
        
        this.localResultsSection = document.getElementById('local-results-section');
        this.rankingsResultsSection = document.getElementById('rankings-results-section');
        this.javdbResultsSection = document.getElementById('javdb-results-section');
        this.localMoviesGrid = document.getElementById('local-movies-grid');
        this.rankingsGrid = document.getElementById('rankings-grid');
        this.javdbResultsContainer = document.getElementById('javdb-results-container');
        
        this.resultsTitle = document.getElementById('results-title');
        this.resultsStats = document.getElementById('results-stats');
        this.localCount = document.getElementById('local-count');
        this.rankingsCount = document.getElementById('rankings-count');
        this.javdbCount = document.getElementById('javdb-count');
        
        this.searchTimeout = null;
        this.currentQuery = '';
        
        this.init();
    }
    
    init() {
        this.bindEvents();
        this.checkUrlParams();
    }
    
    bindEvents() {
        // 搜索按钮点击
        this.searchBtn.addEventListener('click', () => {
            this.performSearch();
        });
        
        // 回车搜索
        this.searchInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.performSearch();
            }
        });
        
        // 实时搜索建议
        this.searchInput.addEventListener('input', (e) => {
            clearTimeout(this.searchTimeout);
            const query = e.target.value.trim();
            
            if (query.length >= 2) {
                this.searchTimeout = setTimeout(() => {
                    this.getSuggestions(query);
                }, 300);
            } else {
                this.hideSuggestions();
            }
        });
        
        // 点击外部隐藏建议
        document.addEventListener('click', (e) => {
            if (!this.searchInput.contains(e.target) && !this.suggestionsDropdown.contains(e.target)) {
                this.hideSuggestions();
            }
        });
        
        // 搜索选项变化
        this.searchLocalCheckbox.addEventListener('change', () => {
            if (this.currentQuery) {
                this.performSearch();
            }
        });
        
        this.searchRankingsCheckbox.addEventListener('change', () => {
            if (this.currentQuery) {
                this.performSearch();
            }
        });
        
        // JAVDb搜索选项变化
        this.searchJavdbCheckbox.addEventListener('change', () => {
            if (this.searchJavdbCheckbox.checked) {
                this.javdbOptions.classList.remove('hidden');
            } else {
                this.javdbOptions.classList.add('hidden');
            }
            
            if (this.currentQuery) {
                this.performSearch();
            }
        });
    }
    
    // 检查URL参数
    checkUrlParams() {
        const urlParams = new URLSearchParams(window.location.search);
        const query = urlParams.get('q');
        if (query) {
            this.searchInput.value = query;
            this.performSearch();
        }
    }
    
    // 执行搜索
    async performSearch() {
        const query = this.searchInput.value.trim();
        if (!query) {
            showNotification('请输入搜索关键词', 'warning');
            return;
        }
        
        this.currentQuery = query;
        this.showLoading();
        this.hideSuggestions();
        
        try {
            const searchLocal = this.searchLocalCheckbox.checked;
            const searchRankings = this.searchRankingsCheckbox.checked;
            const searchJavdb = this.searchJavdbCheckbox.checked;
            
            if (!searchLocal && !searchRankings && !searchJavdb) {
                showNotification('请至少选择一个搜索范围', 'warning');
                this.hideLoading();
                return;
            }
            
            // 存储所有搜索结果
            let localResults = { local_movies: [], rankings: [] };
            let javdbResults = null;
            
            // 执行本地和排行榜搜索
            if (searchLocal || searchRankings) {
                // 确定搜索类型
                let searchType = 'all';
                if (searchLocal && !searchRankings) {
                    searchType = 'local';
                } else if (!searchLocal && searchRankings) {
                    searchType = 'ranking';
                }
                
                const params = new URLSearchParams({
                    q: query,
                    type: searchType,
                    page: 1,
                    limit: 20
                });
                
                const response = await fetch(`/api/v1/search/?${params}`);
                const data = await response.json();
                
                if (!response.ok) {
                    throw new Error(data.message || '本地搜索失败');
                }
                
                if (data.code === 'SUCCESS' && data.data) {
                    localResults = data.data;
                }
            }
            
            // 执行JAVDb搜索
            if (searchJavdb) {
                try {
                    javdbResults = await this.performJAVDbSearch(query);
                } catch (javdbError) {
                    console.error('JAVDb搜索失败:', javdbError);
                    // JAVDb搜索失败不影响整体搜索
                    showNotification(`JAVDb搜索失败: ${javdbError.message}`, 'warning');
                }
            }
            
            // 显示所有搜索结果
            this.displayAllResults(localResults, javdbResults);
            
        } catch (error) {
            console.error('搜索错误:', error);
            showNotification(`搜索失败: ${error.message}`, 'error');
            this.hideLoading();
        }
    }
    
    // 执行JAVDb搜索
    async performJAVDbSearch(query) {
        // 获取搜索类型
        const javdbTypeRadios = document.querySelectorAll('input[name="javdb-type"]');
        let searchType = 'auto';
        for (const radio of javdbTypeRadios) {
            if (radio.checked) {
                searchType = radio.value;
                break;
            }
        }
        
        const params = new URLSearchParams({
            q: query
        });
        
        // 如果不是自动识别，指定搜索类型
        if (searchType !== 'auto') {
            params.append('type', searchType);
        }
        
        const response = await fetch(`/api/v1/search/javdb?${params}`);
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.message || 'JAVDb搜索失败');
        }
        
        if (data.code === 'SUCCESS' && data.data) {
            return data.data;
        } else {
            throw new Error(data.message || 'JAVDb搜索返回数据格式错误');
        }
    }
    
    // 获取搜索建议
    async getSuggestions(query) {
        try {
            const response = await fetch(`/api/v1/search/suggestions?q=${encodeURIComponent(query)}`);
            const data = await response.json();
            
            if (response.ok && data.code === 'SUCCESS' && data.data && data.data.suggestions && data.data.suggestions.length > 0) {
                this.showSuggestions(data.data.suggestions);
            } else {
                this.hideSuggestions();
            }
        } catch (error) {
            console.error('获取搜索建议失败:', error);
            this.hideSuggestions();
        }
    }
    
    // 显示搜索建议
    showSuggestions(suggestions) {
        this.suggestionsDropdown.innerHTML = '';
        
        suggestions.forEach(suggestion => {
            const item = document.createElement('div');
            item.className = 'search-suggestion px-4 py-3 text-white';
            item.innerHTML = `
                <div class="flex items-center">
                    <i class="fas fa-search text-gray-400 mr-2"></i>
                    <span>${this.highlightMatch(suggestion, this.searchInput.value)}</span>
                </div>
            `;
            
            item.addEventListener('click', () => {
                this.searchInput.value = suggestion;
                this.performSearch();
                this.hideSuggestions();
            });
            
            this.suggestionsDropdown.appendChild(item);
        });
        
        this.suggestionsDropdown.classList.remove('hidden');
    }
    
    // 隐藏搜索建议
    hideSuggestions() {
        this.suggestionsDropdown.classList.add('hidden');
    }
    
    // 高亮匹配文本
    highlightMatch(text, query) {
        const regex = new RegExp(`(${query})`, 'gi');
        return text.replace(regex, '<strong class="text-blue-400">$1</strong>');
    }
    
    // 显示所有搜索结果
    displayAllResults(localData, javdbData) {
        this.hideLoading();
        
        const localMovies = localData.local_movies || [];
        const rankings = localData.rankings || [];
        const hasJavdbResults = javdbData !== null;
        
        const totalResults = localMovies.length + rankings.length + (hasJavdbResults ? 1 : 0);
        
        if (totalResults === 0) {
            this.showNoResults();
            return;
        }
        
        // 更新统计信息
        this.resultsTitle.textContent = `"${this.currentQuery}" 的搜索结果`;
        this.resultsStats.textContent = `找到 ${totalResults} 个结果`;
        this.localCount.textContent = localMovies.length;
        this.rankingsCount.textContent = rankings.length;
        this.javdbCount.textContent = hasJavdbResults ? 1 : 0;
        
        // 渲染本地影片结果
        this.renderLocalMovies(localMovies);
        
        // 渲染排行榜结果
        this.renderRankings(rankings);
        
        // 渲染JAVDb搜索结果
        this.renderJAVDbResults(javdbData);
        
        // 显示/隐藏相应部分
        this.localResultsSection.style.display = localMovies.length > 0 ? 'block' : 'none';
        this.rankingsResultsSection.style.display = rankings.length > 0 ? 'block' : 'none';
        this.javdbResultsSection.style.display = hasJavdbResults ? 'block' : 'none';
        
        this.showResults();
    }
    
    // 渲染本地影片
    renderLocalMovies(movies) {
        this.localMoviesGrid.innerHTML = '';
        
        movies.forEach(movie => {
            const movieCard = this.createLocalMovieCard(movie);
            this.localMoviesGrid.appendChild(movieCard);
        });
    }
    
    // 渲染排行榜影片
    renderRankings(rankings) {
        this.rankingsGrid.innerHTML = '';
        
        rankings.forEach(ranking => {
            const rankingCard = this.createRankingCard(ranking);
            this.rankingsGrid.appendChild(rankingCard);
        });
    }
    
    // 创建本地影片卡片
    createLocalMovieCard(movie) {
        const card = document.createElement('div');
        card.className = 'movie-card rounded-xl overflow-hidden';
        
        const imageUrl = movie.has_fanart && movie.fanart_url ? movie.fanart_url : 'static/images/placeholder.svg';
        const fileSize = this.formatFileSize(movie.size);
        
        card.innerHTML = `
            <div class="relative">
                <img src="${imageUrl}" alt="${movie.title}" class="w-full h-48 object-cover" 
                     onerror="this.src='static/images/placeholder.svg'">
                <div class="absolute top-3 left-3">
                    <span class="count-badge px-3 py-1 rounded-full text-xs font-semibold">
                        <i class="fas fa-hdd mr-1"></i>本地
                    </span>
                </div>
            </div>
            <div class="p-4">
                <h3 class="font-semibold text-white mb-3 line-clamp-2" title="${movie.title}">
                    ${movie.title}
                </h3>
                <div class="text-sm text-gray-400 space-y-2">
                    ${movie.code ? `<p><i class="fas fa-tag mr-2 text-blue-400"></i>番号: ${movie.code}</p>` : ''}
                    ${movie.actress ? `<p><i class="fas fa-user mr-2 text-purple-400"></i>演员: ${movie.actress}</p>` : ''}
                    ${fileSize ? `<p><i class="fas fa-file mr-2 text-green-400"></i>大小: ${fileSize}</p>` : ''}
                    ${movie.format ? `<p><i class="fas fa-video mr-2 text-yellow-400"></i>格式: ${movie.format}</p>` : ''}
                    <p><i class="fas fa-folder mr-2 text-gray-500"></i>路径: ${movie.path.split('/').pop()}</p>
                </div>
            </div>
        `;
        
        return card;
    }
    
    // 创建排行榜影片卡片
    createRankingCard(ranking) {
        const card = document.createElement('div');
        card.className = 'movie-card rounded-xl overflow-hidden';
        
        const imageUrl = ranking.cover_url || 'static/images/placeholder.svg';
        
        card.innerHTML = `
            <div class="relative">
                <img src="${imageUrl}" alt="${ranking.title}" class="w-full h-48 object-cover" 
                     onerror="this.src='static/images/placeholder.svg'">
                <div class="absolute top-3 left-3">
                    <span class="count-badge yellow px-3 py-1 rounded-full text-xs font-semibold">
                        <i class="fas fa-trophy mr-1"></i>${ranking.rank_type}
                    </span>
                </div>
                <div class="absolute top-3 right-3">
                    <span class="rating-badge px-3 py-1 rounded-full text-xs font-semibold">
                        #${ranking.position}
                    </span>
                </div>
            </div>
            <div class="p-4">
                <h3 class="font-semibold text-white mb-3 line-clamp-2" title="${ranking.title}">
                    ${ranking.title}
                </h3>
                <div class="text-sm text-gray-400 space-y-2">
                    ${ranking.code ? `<p><i class="fas fa-tag mr-2 text-blue-400"></i>番号: ${ranking.code}</p>` : ''}
                    <p><i class="fas fa-chart-line mr-2 text-yellow-400"></i>排名: #${ranking.position}</p>
                    <p><i class="fas fa-list mr-2 text-purple-400"></i>类型: ${this.getRankTypeText(ranking.rank_type)}</p>
                </div>
                ${ranking.local_exists ? `
                <div class="mt-3">
                    <span class="count-badge green inline-flex items-center px-3 py-1 text-xs rounded-full font-medium">
                        <i class="fas fa-check mr-1"></i>本地已有
                    </span>
                </div>
                ` : ''}
            </div>
        `;
        
        return card;
    }
    
    // 获取排行榜类型文本
    getRankTypeText(rankType) {
        const typeMap = {
            'daily': '日榜',
            'weekly': '周榜', 
            'monthly': '月榜'
        };
        return typeMap[rankType] || rankType;
    }
    
    // 工具函数：格式化文件大小
    formatFileSize(bytes) {
        if (!bytes) return '';
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
    }
    
    // 工具函数：格式化时长
    formatDuration(seconds) {
        if (!seconds) return '';
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        
        if (hours > 0) {
            return `${hours}:${minutes.toString().padStart(2, '0')}:${(seconds % 60).toString().padStart(2, '0')}`;
        } else {
            return `${minutes}:${(seconds % 60).toString().padStart(2, '0')}`;
        }
    }
    
    // 显示加载状态
    showLoading() {
        this.hideAllStates();
        this.loadingState.classList.remove('hidden');
    }
    
    // 隐藏加载状态
    hideLoading() {
        this.loadingState.classList.add('hidden');
    }
    
    // 显示搜索结果
    showResults() {
        this.hideAllStates();
        this.searchResults.classList.remove('hidden');
    }
    
    // 显示无结果状态
    showNoResults() {
        this.hideAllStates();
        this.noResults.classList.remove('hidden');
    }
    
    // 隐藏所有状态
    hideAllStates() {
        this.searchResults.classList.add('hidden');
        this.emptyState.classList.add('hidden');
        this.noResults.classList.add('hidden');
        this.loadingState.classList.add('hidden');
    }
    
    // 渲染JAVDb搜索结果
    renderJAVDbResults(javdbData) {
        this.javdbResultsContainer.innerHTML = '';
        
        if (!javdbData) {
            return;
        }
        
        // 检查是影片搜索结果还是演员搜索结果
        if (javdbData.code) {
            // 影片搜索结果 - 使用网格布局
            const gridContainer = document.createElement('div');
            gridContainer.className = 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6';
            
            const movieCard = this.createJAVDbMovieCard(javdbData);
            gridContainer.appendChild(movieCard);
            this.javdbResultsContainer.appendChild(gridContainer);
        } else if (javdbData.name) {
            // 演员搜索结果 - 保持原有布局
            const actressCard = this.createJAVDbActressCard(javdbData);
            this.javdbResultsContainer.appendChild(actressCard);
        }
    }
    
    // 创建JAVDb影片卡片
    createJAVDbMovieCard(movie) {
        const card = document.createElement('div');
        card.className = 'movie-card rounded-xl overflow-hidden';
        
        // 确保评分是数字类型且大于0
        const rating = parseFloat(movie.rating);
        const hasRating = !isNaN(rating) && rating > 0;
        

        
        card.innerHTML = `
            <div class="relative">
                ${movie.cover_url ? `<img src="${movie.cover_url}" alt="${movie.title}" class="w-full h-48 object-cover">` : `
                    <div class="w-full h-48 flex items-center justify-center bg-gray-700">
                        <i class="fas fa-film text-4xl text-gray-400"></i>
                    </div>
                `}
                <div class="absolute top-3 left-3">
                    <span class="count-badge green px-3 py-1 rounded-full text-xs font-semibold">
                        <i class="fas fa-globe mr-1"></i>JAVDb
                    </span>
                </div>
                ${hasRating ? `
                    <div class="absolute top-3 right-3 z-10">
                        <span class="rating-badge flex items-center px-3 py-1 rounded-full text-xs font-semibold shadow-lg">
                            <i class="fas fa-star mr-1"></i>
                            ${rating.toFixed(1)}
                        </span>
                    </div>
                ` : ''}
            </div>
            <div class="p-4">
                <h3 class="font-semibold text-white mb-3 line-clamp-2" title="${movie.title}">
                    ${movie.title}
                </h3>
                <div class="text-sm text-gray-400 space-y-2">
                    ${movie.code ? `<p><i class="fas fa-tag mr-2 text-blue-400"></i>番号: ${movie.code}</p>` : ''}
                    ${hasRating ? `<p><i class="fas fa-star mr-2 text-yellow-400"></i>评分: ${rating.toFixed(1)}</p>` : ''}
                    ${movie.release_date ? `<p><i class="fas fa-calendar mr-2 text-green-400"></i>发行: ${movie.release_date}</p>` : ''}
                    <p>
                        <a href="${movie.detail_url}" target="_blank" class="text-blue-400 hover:text-blue-300 transition-colors">
                            <i class="fas fa-external-link-alt mr-2"></i>在JAVDb查看详情
                        </a>
                    </p>
                </div>
            </div>
        `;
        
        return card;
    }
    
    // 创建JAVDb演员卡片
    createJAVDbActressCard(actress) {
        const card = document.createElement('div');
        card.className = 'movie-card rounded-xl';
        
        card.innerHTML = `
            <div class="p-6">
                <div class="flex items-center space-x-4">
                    <div class="flex-shrink-0">
                        ${actress.avatar_url ? `
                            <img src="${actress.avatar_url}" alt="${actress.name}" class="w-16 h-16 rounded-full object-cover border-2 border-white/20">
                        ` : `
                            <div class="w-16 h-16 rounded-full bg-white/10 flex items-center justify-center border-2 border-white/20">
                                <i class="fas fa-user text-2xl text-gray-400"></i>
                            </div>
                        `}
                    </div>
                    <div class="flex-1">
                        <div class="flex items-center gap-3 mb-2">
                            <h3 class="text-lg font-semibold text-white">${actress.name}</h3>
                            <span class="count-badge green px-3 py-1 rounded-full text-xs font-semibold">
                                <i class="fas fa-user mr-1"></i>演员
                            </span>
                        </div>
                        ${actress.movie_count > 0 ? `
                            <p class="text-gray-400 text-sm mb-3">共 ${actress.movie_count} 部作品</p>
                        ` : ''}
                        <div>
                            <a href="${actress.detail_url}" target="_blank" class="text-blue-400 hover:text-blue-300 text-sm font-medium transition-colors">
                                <i class="fas fa-external-link-alt mr-2"></i>在JAVDb查看详情
                            </a>
                        </div>
                    </div>
                </div>
                
                ${actress.movies && actress.movies.length > 0 ? `
                    <div class="mt-6">
                        <h4 class="text-sm font-semibold text-white mb-4">最新作品</h4>
                        <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                            ${actress.movies.slice(0, 4).map(movie => `
                                <div class="bg-white/5 rounded-lg p-3 border border-white/10">
                                    ${movie.cover_url ? `
                                        <img src="${movie.cover_url}" alt="${movie.title}" class="w-full h-24 object-cover rounded mb-2">
                                    ` : `
                                        <div class="w-full h-24 bg-gray-700 rounded mb-2 flex items-center justify-center">
                                            <i class="fas fa-film text-gray-400"></i>
                                        </div>
                                    `}
                                    <div class="text-xs">
                                        <div class="font-semibold text-blue-400 mb-1">${movie.code || '无番号'}</div>
                                        <div class="text-gray-300 line-clamp-1">${movie.title}</div>
                                        ${movie.release_date ? `<div class="text-gray-500 mt-1">${movie.release_date}</div>` : ''}
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                        ${actress.movies.length > 4 ? `
                            <div class="mt-4 text-center">
                                <a href="${actress.detail_url}" target="_blank" class="text-blue-400 hover:text-blue-300 text-sm transition-colors">
                                    查看全部 ${actress.movies.length} 部作品 →
                                </a>
                            </div>
                        ` : ''}
                    </div>
                ` : ''}
            </div>
        `;
        
        return card;
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    new SearchPage();
}); 