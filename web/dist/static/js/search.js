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
        card.className = 'movie-card bg-gray-800 rounded-lg overflow-hidden relative group';
        
        card.innerHTML = `
            <div class="aspect-w-2 aspect-h-3 relative">
                ${ranking.cover_url ? `<img src="${ranking.cover_url}" alt="${ranking.title}" class="object-cover w-full h-full rounded-t-lg">` : `
                    <div class="w-full h-full flex items-center justify-center bg-gray-700 rounded-t-lg">
                        <i class="fas fa-film text-4xl text-gray-400"></i>
                    </div>
                `}
                <div class="absolute top-3 left-3">
                    <span class="count-badge yellow px-3 py-1 rounded-full text-xs font-semibold">
                        <i class="fas fa-trophy mr-1"></i>#${ranking.position}
                    </span>
                </div>
                ${ranking.rating > 0 ? `
                    <div class="absolute top-3 right-3" style="z-index: 100000;">
                        <span class="rating-badge flex items-center px-3 py-1 rounded-full text-xs font-semibold shadow-lg">
                            <i class="fas fa-star mr-1"></i>
                            ${ranking.rating.toFixed(1)}
                    </span>
                </div>
                ` : ''}
            </div>
            <div class="p-4">
                <h3 class="font-semibold text-white mb-3 line-clamp-2" title="${ranking.title}">
                    ${ranking.title}
                </h3>
                <div class="text-sm text-gray-400 space-y-2">
                    ${ranking.code ? `
                        <div class="flex items-center justify-between">
                            <div class="flex items-center">
                                <i class="fas fa-tag mr-2 text-blue-400"></i>番号: ${ranking.code}
                </div>
                            ${!ranking.local_exists ? `
                                <button onclick="searchPage.downloadMovie('${ranking.code}', '${ranking.title.replace(/'/g, "\\'")}', this)" 
                                        class="download-btn bg-blue-500 hover:bg-blue-600 text-white px-2 py-1 rounded text-xs font-medium flex items-center transition-colors duration-200">
                                    <i class="fas fa-download mr-1"></i>下载
                                </button>
                            ` : `
                                <span class="count-badge green inline-flex items-center px-2 py-1 text-xs rounded-full font-medium">
                                    <i class="fas fa-check mr-1"></i>已有
                    </span>
                            `}
                        </div>
                    ` : ''}
                    ${ranking.rating > 0 ? `<p><i class="fas fa-star mr-2 text-yellow-400"></i>评分: ${ranking.rating.toFixed(1)}</p>` : ''}
                    ${ranking.release_date ? `<p><i class="fas fa-calendar mr-2 text-green-400"></i>发行: ${ranking.release_date}</p>` : ''}
                </div>
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
    createJAVDbMovieCard(movieData) {
        const card = document.createElement('div');
        card.className = 'movie-card bg-gray-800 rounded-lg overflow-hidden relative group';
        
        // 添加封面图片
        const coverContainer = document.createElement('div');
        coverContainer.className = 'aspect-w-2 aspect-h-3 relative';
        
        const img = document.createElement('img');
        img.src = movieData.cover_url || '/static/images/placeholder.jpg';
        img.alt = movieData.title;
        img.className = 'object-cover w-full h-full rounded-t-lg';
        coverContainer.appendChild(img);
        
        // 添加评分徽章
        if (movieData.rating > 0) {
            const ratingBadge = document.createElement('div');
            ratingBadge.className = 'absolute top-2 right-2 bg-orange-500 text-white px-2 py-1 rounded-full text-sm font-medium flex items-center';
            ratingBadge.style.zIndex = '100000';
            ratingBadge.innerHTML = `<i class="fas fa-star mr-1"></i>${movieData.rating.toFixed(1)}`;
            coverContainer.appendChild(ratingBadge);
        }
        
        card.appendChild(coverContainer);
        
        // 添加影片信息
        const info = document.createElement('div');
        info.className = 'p-4';
        
        const title = document.createElement('h3');
        title.className = 'text-lg font-medium text-white mb-2';
        title.textContent = movieData.title;
        info.appendChild(title);
        
        const meta = document.createElement('div');
        meta.className = 'text-sm text-gray-400 space-y-2';
        
        // 番号和下载按钮的容器
        const codeContainer = document.createElement('div');
        codeContainer.className = 'flex items-center justify-between mb-2';
        
        const codeDiv = document.createElement('div');
        codeDiv.className = 'flex items-center';
        codeDiv.innerHTML = `<i class="fas fa-hashtag mr-1"></i>${movieData.code}`;
        
        // 添加下载按钮
        const downloadBtn = document.createElement('button');
        downloadBtn.className = 'download-btn bg-blue-500 hover:bg-blue-600 text-white px-2 py-1 rounded text-xs font-medium flex items-center transition-colors duration-200';
        downloadBtn.innerHTML = '<i class="fas fa-download mr-1"></i>下载';
        downloadBtn.onclick = async (e) => {
            e.preventDefault();
            e.stopPropagation();
            await this.searchAndDownloadTorrent(movieData.code);
        };
        
        codeContainer.appendChild(codeDiv);
        codeContainer.appendChild(downloadBtn);
        meta.appendChild(codeContainer);
        
        if (movieData.release_date) {
            const date = document.createElement('div');
            date.innerHTML = `<i class="far fa-calendar mr-1"></i>${movieData.release_date}`;
            meta.appendChild(date);
        }
        
        info.appendChild(meta);
        card.appendChild(info);
        
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
    
    // 下载电影
    async downloadMovie(code, title, buttonElement) {
        if (!code) {
            this.showNotification('番号不能为空', 'error');
            return;
        }
        
        const originalButton = buttonElement;
        const originalContent = originalButton.innerHTML;
        
        try {
            // 更新按钮状态
            originalButton.disabled = true;
            originalButton.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i>搜索中...';
            
            this.showNotification(`正在搜索 ${code} 的种子...`, 'info');
            
            // 使用专门的番号搜索API（会检查本地是否存在）
            const searchResponse = await fetch(`/api/v1/torrents/search/code?code=${encodeURIComponent(code)}`);
            const searchData = await searchResponse.json();
            
            if (searchResponse.status === 409) {
                // 番号已存在
                this.showNotification(searchData.message, 'warning');
                originalButton.innerHTML = '<i class="fas fa-check mr-2"></i>已存在';
                originalButton.className = 'count-badge green inline-flex items-center px-3 py-1 text-xs rounded-full font-medium';
                originalButton.disabled = true;
                return;
            }
            
            if (!searchResponse.ok) {
                throw new Error(searchData.message || '搜索种子失败');
            }
            
            const results = searchData.data.results;
            if (!results || results.length === 0) {
                throw new Error('未找到可用的种子');
            }
            
            // 显示种子选择界面
            this.showTorrentSelection(code, title, results, originalButton);
            
        } catch (error) {
            console.error('搜索种子失败:', error);
            this.showNotification(`搜索失败: ${error.message}`, 'error');
            
            // 恢复按钮状态
            originalButton.disabled = false;
            originalButton.innerHTML = originalContent;
        }
    }
    
    // 显示种子选择界面
    showTorrentSelection(code, title, torrents, originalButton) {
        // 创建弹窗
        const modal = document.createElement('div');
        
        // 完全使用内联样式，不依赖CSS类
        modal.style.cssText = `
            position: fixed !important;
            top: 0 !important;
            left: 0 !important;
            right: 0 !important;
            bottom: 0 !important;
            width: 100vw !important;
            height: 100vh !important;
            display: flex !important;
            align-items: center !important;
            justify-content: center !important;
            padding: 1rem !important;
            z-index: 99999 !important;
            background-color: rgba(0, 0, 0, 0.5) !important;
            backdrop-filter: blur(8px) !important;
        `;
        
        modal.innerHTML = `
            <div style="
                background: #111827 !important;
                border-radius: 1rem !important;
                border: 1px solid rgba(255, 255, 255, 0.1) !important;
                max-width: 64rem !important;
                width: 100% !important;
                max-height: 80vh !important;
                overflow: hidden !important;
                box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25) !important;
                margin: auto !important;
            ">
                <div style="padding: 1.5rem; border-bottom: 1px solid rgba(255, 255, 255, 0.1);">
                    <div style="display: flex; align-items: center; justify-content: space-between;">
                        <div>
                            <h3 style="font-size: 1.25rem; font-weight: bold; color: white; margin: 0;">${title}</h3>
                            <p style="color: #9CA3AF; margin: 0.25rem 0 0 0;">选择要下载的版本 - 按文件大小排序</p>
                        </div>
                        <button onclick="closeModal(this)" 
                                style="color: #9CA3AF; background: none; border: none; cursor: pointer; font-size: 1.25rem; padding: 0.5rem;">
                            ✕
                        </button>
                    </div>
                </div>
                
                <div style="padding: 1.5rem; max-height: 24rem; overflow-y: auto;">
                    <div style="display: flex; flex-direction: column; gap: 0.75rem;">
                        ${torrents.map((torrent, index) => `
                            <div style="
                                background: rgba(31, 41, 55, 0.5);
                                border-radius: 0.75rem;
                                padding: 1rem;
                                border: 1px solid rgba(75, 85, 99, 0.5);
                                transition: border-color 0.2s;
                            " onmouseover="this.style.borderColor='rgba(59, 130, 246, 0.5)'" onmouseout="this.style.borderColor='rgba(75, 85, 99, 0.5)'">
                                <div style="display: flex; align-items: center; justify-content: space-between;">
                                    <div style="flex: 1; min-width: 0;">
                                        <h4 style="color: white; font-weight: 500; margin: 0 0 0.5rem 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" title="${torrent.title}">
                                            ${torrent.title}
                                        </h4>
                                        <div style="display: flex; align-items: center; gap: 1rem; font-size: 0.875rem; color: #9CA3AF;">
                                            <span style="display: flex; align-items: center;">
                                                💾 ${torrent.sizeFormatted}
                                            </span>
                                            <span style="display: flex; align-items: center;">
                                                ⬆️ ${torrent.seeders} 做种
                                            </span>
                                            <span style="display: flex; align-items: center;">
                                                ⬇️ ${torrent.leechers} 下载
                                            </span>
                                            ${torrent.tracker ? `
                                            <span style="display: flex; align-items: center;">
                                                🖥️ ${torrent.tracker}
                                            </span>
                                            ` : ''}
                                        </div>
                                    </div>
                                    <button onclick="searchPage.downloadTorrent('${torrent.magnetUri || torrent.link}', '${code}', '${torrent.title.replace(/'/g, "\\'")}', ${torrent.size}, '${torrent.tracker || ''}', 'close')" 
                                            style="
                                                background: #2563EB;
                                                color: white;
                                                padding: 0.5rem 1rem;
                                                border-radius: 0.5rem;
                                                border: none;
                                                font-size: 0.875rem;
                                                font-weight: 500;
                                                cursor: pointer;
                                                transition: background-color 0.2s;
                                            " onmouseover="this.style.backgroundColor='#1D4ED8'" onmouseout="this.style.backgroundColor='#2563EB'">
                                        ⬇️ 下载
                                    </button>
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>
                
                <div style="padding: 1.5rem; border-top: 1px solid rgba(255, 255, 255, 0.1); background: rgba(31, 41, 55, 0.3);">
                    <div style="display: flex; align-items: center; justify-content: space-between;">
                        <p style="font-size: 0.875rem; color: #9CA3AF; margin: 0;">
                            ℹ️ 建议选择文件大小较大的版本，通常画质更好
                        </p>
                        <button onclick="closeModal(this)" 
                                style="
                                    padding: 0.5rem 1rem;
                                    background: #374151;
                                    color: white;
                                    border-radius: 0.5rem;
                                    border: none;
                                    font-size: 0.875rem;
                                    cursor: pointer;
                                    transition: background-color 0.2s;
                                " onmouseover="this.style.backgroundColor='#4B5563'" onmouseout="this.style.backgroundColor='#374151'">
                            取消
                        </button>
                    </div>
                </div>
            </div>
        `;
        
        // 添加点击背景关闭弹窗
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.remove();
            }
        });
        
        // 添加全局关闭函数
        window.closeModal = function(button) {
            const modal = button.closest('div');
            while (modal && !modal.style.position) {
                modal = modal.parentElement;
            }
            if (modal && modal.style.position === 'fixed') {
                modal.remove();
            }
        };
        
        document.body.appendChild(modal);
        
        // 强制重新计算样式
        modal.offsetHeight;
    }
    
    // 下载选定的种子
    async downloadTorrent(downloadUri, code, title, size, tracker, shouldClose) {
        try {
            this.showNotification('正在添加下载任务...', 'info');
            
            const response = await fetch('/api/v1/torrents/download', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    magnet_uri: downloadUri.startsWith('magnet:') ? downloadUri : '',
                    link: !downloadUri.startsWith('magnet:') ? downloadUri : '',
                    code: code,
                    title: title,
                    size: size,
                    tracker: tracker
                })
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.message || '添加下载任务失败');
            }
            
            this.showNotification('✅ 已添加到下载队列！', 'success');
            
            // 关闭弹窗
            if (shouldClose === 'close') {
                // 查找并关闭弹窗
                const modals = document.querySelectorAll('div[style*="position: fixed"]');
                modals.forEach(modal => {
                    if (modal.style.position === 'fixed' && modal.style.zIndex === '99999') {
                        modal.remove();
                    }
                });
            }
            
            // 可选：跳转到下载管理页面
            setTimeout(() => {
                if (confirm('下载任务已添加，是否前往下载管理页面？')) {
                    window.location.href = '/downloads.html';
                }
            }, 1000);
            
        } catch (error) {
            console.error('下载失败:', error);
            this.showNotification(`下载失败: ${error.message}`, 'error');
        }
    }
    
    // 显示通知
    showNotification(message, type = 'info') {
        // 移除现有通知
        const existingNotification = document.querySelector('.notification');
        if (existingNotification) {
            existingNotification.remove();
        }
        
        const notification = document.createElement('div');
        notification.className = `notification fixed top-4 right-4 z-50 px-6 py-3 rounded-lg text-white font-medium transition-all duration-300 transform`;
        
        let bgColor = 'bg-blue-600';
        let icon = 'fas fa-info-circle';
        
        switch (type) {
            case 'success':
                bgColor = 'bg-green-600';
                icon = 'fas fa-check-circle';
                break;
            case 'error':
                bgColor = 'bg-red-600';
                icon = 'fas fa-exclamation-circle';
                break;
            case 'warning':
                bgColor = 'bg-yellow-600';
                icon = 'fas fa-exclamation-triangle';
                break;
        }
        
        notification.className += ` ${bgColor}`;
        notification.innerHTML = `
            <div class="flex items-center">
                <i class="${icon} mr-2"></i>
                ${message}
            </div>
        `;
        
        document.body.appendChild(notification);
        
        // 自动移除
        setTimeout(() => {
            notification.style.transform = 'translateX(100%)';
            setTimeout(() => notification.remove(), 300);
        }, 4000);
    }

    // 搜索并下载种子
    async searchAndDownloadTorrent(code) {
        try {
            showNotification('正在搜索种子...', 'info');
            
            // 调用Jackett API搜索种子
            const response = await fetch(`/api/v1/torrents/search?q=${encodeURIComponent(code)}`);
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.message || '搜索种子失败');
            }
            
            if (!data.data || data.data.length === 0) {
                showNotification('未找到可用的种子', 'warning');
                return;
            }
            
            // 选择最佳种子（这里简单地选择第一个结果）
            const bestTorrent = data.data[0];
            
            // 添加到qBittorrent
            const downloadResponse = await fetch('/api/v1/torrents/download', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: `magnet_uri=${encodeURIComponent(bestTorrent.magnetUri)}`,
            });
            
            const downloadData = await downloadResponse.json();
            
            if (!downloadResponse.ok) {
                throw new Error(downloadData.message || '添加下载任务失败');
            }
            
            showNotification('已添加到下载队列', 'success');
            
        } catch (error) {
            console.error('下载错误:', error);
            showNotification(`下载失败: ${error.message}`, 'error');
        }
    }
}

// 全局变量，供HTML中的onclick使用
let searchPage;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    searchPage = new SearchPage();
}); 