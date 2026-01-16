/**
 * Today Page Logic
 * 今日更新页面的核心逻辑
 */

class TodayPage {
    constructor() {
        this.selectedDate = new Date();
        this.shows = [];
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadTodayUpdates();
    }

    bindEvents() {
        // 日期选择
        document.getElementById('dateInput').addEventListener('change', (e) => {
            this.selectedDate = new Date(e.target.value);
            this.loadDateUpdates(this.selectedDate);
        });

        // 快捷选择按钮
        document.getElementById('todayBtn').addEventListener('click', () => {
            this.selectDate('today');
        });

        document.getElementById('yesterdayBtn').addEventListener('click', () => {
            this.selectDate('yesterday');
        });

        document.getElementById('weekBtn').addEventListener('click', () => {
            this.loadWeekUpdates();
        });

        document.getElementById('monthBtn').addEventListener('click', () => {
            this.loadMonthUpdates();
        });

        // 操作按钮
        document.getElementById('refreshBtn').addEventListener('click', () => {
            this.loadDateUpdates(this.selectedDate);
        });

        document.getElementById('publishTelegraphBtn').addEventListener('click', () => {
            this.publishToTelegraph();
        });

        document.getElementById('exportMarkdownBtn').addEventListener('click', () => {
            this.exportMarkdown();
        });

        // 移动端按钮事件（与桌面端共享处理函数）
        const publishBtnMobile = document.getElementById('publishTelegraphBtnMobile');
        const exportBtnMobile = document.getElementById('exportMarkdownBtnMobile');

        if (publishBtnMobile) {
            publishBtnMobile.addEventListener('click', () => {
                this.publishToTelegraph();
            });
        }
        if (exportBtnMobile) {
            exportBtnMobile.addEventListener('click', () => {
                this.exportMarkdown();
            });
        }

        // 初始化日期选择器为今天
        document.getElementById('dateInput').value = this.formatDateForInput(new Date());
    }

    async loadTodayUpdates() {
        this.showLoading(true);

        try {
            // 使用今日更新API获取集数级别的更新
            const response = await api.getTodayUpdates();
            
            if (response.code === 0) {
                const updates = response.data || [];
                
                // 按剧集分组
                const showMap = new Map();
                
                updates.forEach(update => {
                    const showId = update.show_id;
                    
                    if (!showMap.has(showId)) {
                        showMap.set(showId, {
                            id: showId,
                            name: update.show_name,
                            poster_path: update.still_path,
                            status: 'Returning Series',
                            vote_average: update.vote_average,
                            first_air_date: update.air_date,
                            episode_count: 0,
                            episodes: []
                        });
                    }
                    
                    const show = showMap.get(showId);
                    show.episodes.push(update);
                    show.episode_count++;
                });
                
                this.shows = Array.from(showMap.values());
                this.renderShows();
                this.updateStats();
            } else {
                this.showError('加载失败: ' + response.message);
            }
        } catch (error) {
            this.showError('加载失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    renderShows() {
        const container = document.getElementById('showsContainer');
        const content = document.getElementById('showsContent');
        const empty = document.getElementById('emptyState');

        container.innerHTML = '';

        if (this.shows.length === 0) {
            content.style.display = 'none';
            empty.style.display = 'block';
            return;
        }

        content.style.display = 'block';
        empty.style.display = 'none';

        this.shows.forEach(show => {
            const row = document.createElement('div');
            row.className = 'episode-list-row mb-3';

            // 构建集数列表HTML
            let episodesHTML = '';
            if (show.episodes && show.episodes.length > 0) {
                episodesHTML = '<div class="episodes-list">';
                show.episodes.forEach(ep => {
                    const episodeCode = `S${ep.season_number}E${ep.episode_number}`;
                    const isUploaded = ep.uploaded || false;
                    const checkBtnClass = isUploaded ? 'uploaded' : '';
                    const btnTitle = isUploaded ? '已上传 - 点击取消' : '标记已上传';

                    episodesHTML += `
                        <div class="episode-row">
                            <span class="episode-code">${episodeCode}</span>
                            <span class="episode-name">${this.escapeHtml(ep.name)}</span>
                            <button class="upload-check-btn ${checkBtnClass}"
                                    data-episode-id="${ep.id}"
                                    onclick="todayPage.toggleUploaded(${ep.id}, event)"
                                    title="${btnTitle}">
                            </button>
                        </div>`;
                });
                episodesHTML += '</div>';
            }

            row.innerHTML = `
                <div class="card">
                    <div class="card-body py-2">
                        <div class="show-header">
                            <h5 class="show-name">${this.escapeHtml(show.name)}</h5>
                            <span class="show-date text-muted">${this.formatDate(show.first_air_date)}</span>
                        </div>
                        ${episodesHTML}
                    </div>
                </div>
            `;
            container.appendChild(row);
        });
    }

    updateStats() {
        const totalShows = this.shows.length;
        const totalEpisodes = this.shows.reduce((sum, show) => sum + (show.episode_count || 0), 0);
        
        document.getElementById('totalCount').textContent = totalShows;
        document.getElementById('newCount').textContent = totalEpisodes;
        document.getElementById('returningCount').textContent = totalShows;
        document.getElementById('endedCount').textContent = totalEpisodes;
    }

    selectDate(type) {
        const today = new Date();

        switch(type) {
            case 'today':
                this.selectedDate = today;
                break;
            case 'yesterday':
                this.selectedDate = new Date(today);
                this.selectedDate.setDate(today.getDate() - 1);
                break;
        }

        document.getElementById('dateInput').value = this.formatDateForInput(this.selectedDate);
        this.loadDateUpdates(this.selectedDate);
    }

    async loadDateUpdates(date) {
        this.showLoading(true);

        try {
            const dateStr = this.formatDateForInput(date);
            const response = await api.getDateRangeUpdates(dateStr, dateStr);

            if (response.code === 0) {
                const updates = response.data || [];

                // 按剧集分组
                const showMap = new Map();

                updates.forEach(update => {
                    const showId = update.show_id;

                    if (!showMap.has(showId)) {
                        showMap.set(showId, {
                            id: showId,
                            name: update.show_name,
                            poster_path: update.still_path,
                            status: 'Returning Series',
                            vote_average: update.vote_average,
                            first_air_date: update.air_date,
                            episode_count: 0,
                            episodes: []
                        });
                    }

                    const show = showMap.get(showId);
                    show.episodes.push(update);
                    show.episode_count++;
                });

                this.shows = Array.from(showMap.values());
                this.renderShows();
                this.updateStats();
            } else {
                this.showError('加载失败: ' + response.message);
            }
        } catch (error) {
            this.showError('加载失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    async loadWeekUpdates() {
        this.showLoading(true);

        try {
            const today = new Date();
            const dayOfWeek = today.getDay();
            const startOfWeek = new Date(today);
            startOfWeek.setDate(today.getDate() - dayOfWeek);
            const endOfWeek = new Date(today);
            endOfWeek.setDate(today.getDate() + (6 - dayOfWeek));

            const startDate = this.formatDateForInput(startOfWeek);
            const endDate = this.formatDateForInput(endOfWeek);

            const response = await api.getDateRangeUpdates(startDate, endDate);

            if (response.code === 0) {
                const updates = response.data || [];

                const showMap = new Map();
                updates.forEach(update => {
                    const showId = update.show_id;
                    if (!showMap.has(showId)) {
                        showMap.set(showId, {
                            id: showId,
                            name: update.show_name,
                            poster_path: update.still_path,
                            status: 'Returning Series',
                            vote_average: update.vote_average,
                            first_air_date: update.air_date,
                            episode_count: 0,
                            episodes: []
                        });
                    }
                    const show = showMap.get(showId);
                    show.episodes.push(update);
                    show.episode_count++;
                });

                this.shows = Array.from(showMap.values());
                this.renderShows();
                this.updateStats();
                this.selectedDate = startOfWeek;
                document.getElementById('dateInput').value = startDate;
            } else {
                this.showError('加载失败: ' + response.message);
            }
        } catch (error) {
            this.showError('加载失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    async loadMonthUpdates() {
        this.showLoading(true);

        try {
            const today = new Date();
            const startOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
            const endOfMonth = new Date(today.getFullYear(), today.getMonth() + 1, 0);

            const startDate = this.formatDateForInput(startOfMonth);
            const endDate = this.formatDateForInput(endOfMonth);

            const response = await api.getDateRangeUpdates(startDate, endDate);

            if (response.code === 0) {
                const updates = response.data || [];

                const showMap = new Map();
                updates.forEach(update => {
                    const showId = update.show_id;
                    if (!showMap.has(showId)) {
                        showMap.set(showId, {
                            id: showId,
                            name: update.show_name,
                            poster_path: update.still_path,
                            status: 'Returning Series',
                            vote_average: update.vote_average,
                            first_air_date: update.air_date,
                            episode_count: 0,
                            episodes: []
                        });
                    }
                    const show = showMap.get(showId);
                    show.episodes.push(update);
                    show.episode_count++;
                });

                this.shows = Array.from(showMap.values());
                this.renderShows();
                this.updateStats();
                this.selectedDate = startOfMonth;
                document.getElementById('dateInput').value = startDate;
            } else {
                this.showError('加载失败: ' + response.message);
            }
        } catch (error) {
            this.showError('加载失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    async refreshShow(id) {
        try {
            const response = await api.refreshShow(id);
            if (response.code === 0) {
                this.showSuccess('刷新成功');
                this.loadDateUpdates(this.selectedDate);
            } else {
                this.showError('刷新失败: ' + response.message);
            }
        } catch (error) {
            this.showError('刷新失败: ' + error.message);
        }
    }

    /**
     * 切换剧集的上传状态
     * @param {number} episodeId - 剧集ID
     * @param {Event} event - 点击事件
     */
    async toggleUploaded(episodeId, event) {
        event.stopPropagation(); // 防止触发父元素点击

        const btn = event.target.closest('.upload-check-btn');
        if (!btn) return;

        const isCurrentlyUploaded = btn.classList.contains('uploaded');

        // 检查认证状态
        if (!api.isAuthenticated) {
            this.showError('请先登录后再操作');
            // 触发登录流程
            if (typeof showLoginModal === 'function') {
                showLoginModal('标记上传状态需要管理员权限');
            }
            return;
        }

        // 乐观更新 UI
        btn.classList.add('loading');
        if (isCurrentlyUploaded) {
            btn.classList.remove('uploaded');
        } else {
            btn.classList.add('uploaded');
        }

        try {
            // 调用 API
            if (isCurrentlyUploaded) {
                await api.unmarkEpisodeUploaded(episodeId);
                btn.title = '标记已上传';
            } else {
                await api.markEpisodeUploaded(episodeId);
                btn.title = '已上传 - 点击取消';
            }

            // 成功后更新本地数据
            this.updateLocalEpisodeStatus(episodeId, !isCurrentlyUploaded);
            this.showSuccess(isCurrentlyUploaded ? '已取消标记' : '已标记为上传');
        } catch (error) {
            // 失败回滚 UI
            if (isCurrentlyUploaded) {
                btn.classList.add('uploaded');
                btn.title = '已上传 - 点击取消';
            } else {
                btn.classList.remove('uploaded');
                btn.title = '标记已上传';
            }

            // 检查是否是认证错误
            if (error.message && error.message.includes('Unauthorized')) {
                this.showError('登录已过期，请重新登录');
                if (typeof showLoginModal === 'function') {
                    showLoginModal('登录已过期，请重新登录以继续操作');
                }
            } else {
                this.showError('操作失败: ' + error.message);
            }
        } finally {
            btn.classList.remove('loading');
        }
    }

    /**
     * 更新本地剧集的上传状态
     * @param {number} episodeId - 剧集ID
     * @param {boolean} uploaded - 上传状态
     */
    updateLocalEpisodeStatus(episodeId, uploaded) {
        // 更新 this.shows 中对应剧集的 uploaded 状态
        for (const show of this.shows) {
            if (show.episodes) {
                const episode = show.episodes.find(ep => ep.id === episodeId);
                if (episode) {
                    episode.uploaded = uploaded;
                    break;
                }
            }
        }
    }

    async publishToTelegraph() {
        // 检查认证状态
        if (!api.isAuthenticated) {
            this.showError('请先登录后再操作');
            if (typeof showLoginModal === 'function') {
                showLoginModal('发布到 Telegraph 需要管理员权限');
            }
            return;
        }

        if (!confirm('确定要发布今日更新到Telegraph吗?')) return;

        try {
            this.showLoading(true);
            console.log('[publishToTelegraph] 开始发布...');

            const response = await api.publishToday();
            console.log('[publishToTelegraph] 响应:', response);

            if (response.code === 0 && response.data && response.data.success) {
                this.showSuccess('发布成功!');

                if (response.data.url) {
                    this.showPublishModal(response.data.url);
                }
            } else {
                const errorMsg = response.message || response.data?.message || '未知错误';
                this.showError('发布失败: ' + errorMsg);
            }
        } catch (error) {
            console.error('[publishToTelegraph] 错误:', error);

            // 检查是否是认证错误
            if (error.message && error.message.includes('Unauthorized')) {
                this.showError('登录已过期，请重新登录');
                if (typeof showLoginModal === 'function') {
                    showLoginModal('登录已过期，请重新登录以继续操作');
                }
            } else {
                this.showError('发布失败: ' + error.message);
            }
        } finally {
            this.showLoading(false);
        }
    }

    async exportMarkdown() {
        try {
            const markdown = await api.getTodayMarkdown();
            
            const blob = new Blob([markdown], { type: 'text/markdown' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `今日更新_${this.formatDateForFile(new Date())}.md`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

            this.showSuccess('Markdown导出成功');
        } catch (error) {
            this.showError('导出失败: ' + error.message);
        }
    }

    showPublishModal(url) {
        const modal = document.createElement('div');
        modal.className = 'modal fade';
        modal.innerHTML = `
            <div class="modal-dialog">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">发布成功</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                    </div>
                    <div class="modal-body">
                        <p>今日更新已成功发布到Telegraph!</p>
                        <div class="input-group">
                            <input type="text" class="form-control" value="${url}" readonly>
                            <button class="btn btn-outline-secondary" onclick="window.open('${url}', '_blank')">
                                <i class="bi bi-box-arrow-up-right"></i> 打开
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `;
        document.body.appendChild(modal);
        const bsModal = new bootstrap.Modal(modal);
        bsModal.show();
        modal.addEventListener('hidden.bs.toast', () => {
            modal.remove();
        });
    }

    showLoading(show) {
        const spinner = document.getElementById('loadingSpinner');
        const content = document.getElementById('showsContent');
        
        if (show) {
            spinner.style.display = 'block';
            content.style.display = 'none';
        } else {
            spinner.style.display = 'none';
        }
    }

    showSuccess(message) {
        this.showToast(message, 'success');
    }

    showError(message) {
        this.showToast(message, 'danger');
    }

    showToast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast align-items-center text-white bg-${type} border-0`;
        toast.innerHTML = `
            <div class="d-flex">
                <div class="toast-body">${message}</div>
                <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
            </div>
        `;
        container.appendChild(toast);

        const bsToast = new bootstrap.Toast(toast);
        bsToast.show();

        toast.addEventListener('hidden.bs.toast', () => {
            toast.remove();
        });
    }

    renderStatusBadge(status) {
        const badges = {
            'Returning Series': '<span class="badge badge-returning">连载中</span>',
            'Ended': '<span class="badge badge-ended">已完结</span>',
            'Canceled': '<span class="badge badge-canceled">已取消</span>'
        };
        return badges[status] || `<span class="badge bg-secondary">${status || '未知'}</span>`;
    }

    getPosterUrl(path) {
        if (!path) {
            return 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="80" height="120"%3E%3Crect fill="%23ddd" width="80" height="120"/%3E%3Ctext fill="%23999" x="50%25" y="50%25" text-anchor="middle" dy=".3em" font-size="12"%3E无海报%3C/text%3E%3C/svg%3E';
        }
        return `https://image.tmdb.org/t/p/w200${path}`;
    }

    formatDate(dateStr) {
        if (!dateStr) return '-';
        return dateStr.split('T')[0];
    }

    formatDateForInput(date) {
        return date.toISOString().split('T')[0];
    }

    formatDateForFile(date) {
        return date.toISOString().split('T')[0];
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// 初始化页面
let todayPage;
document.addEventListener('DOMContentLoaded', async () => {
    // 先初始化认证状态，确保 api.isAuthenticated 正确设置
    if (typeof initAuthUI === 'function') {
        await initAuthUI();
    }
    todayPage = new TodayPage();
});
