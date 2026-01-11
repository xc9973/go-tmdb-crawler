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
            this.loadTodayUpdates();
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
            this.loadTodayUpdates();
        });

        document.getElementById('publishTelegraphBtn').addEventListener('click', () => {
            this.publishToTelegraph();
        });

        document.getElementById('exportMarkdownBtn').addEventListener('click', () => {
            this.exportMarkdown();
        });

        // 初始化日期选择器为今天
        document.getElementById('dateInput').value = this.formatDateForInput(new Date());
    }

    async loadTodayUpdates() {
        this.showLoading(true);

        try {
            // 获取今日更新的剧集
            const response = await api.getShows(1, 100, '', '');
            
            if (response.code === 0) {
                // 过滤出今日更新的剧集
                this.shows = this.filterTodayShows(response.data.items || []);
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

    filterTodayShows(shows) {
        // 这里应该根据实际API返回的更新时间过滤
        // 暂时返回所有剧集作为示例
        return shows;
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
            const col = document.createElement('div');
            col.className = 'col-md-6 col-lg-4 mb-3';
            col.innerHTML = `
                <div class="card h-100">
                    <div class="card-body">
                        <div class="d-flex align-items-start">
                            <img src="${this.getPosterUrl(show.poster_path)}" 
                                 alt="${show.name}" 
                                 class="rounded me-3" 
                                 style="width: 80px; height: 120px; object-fit: cover;">
                            <div class="flex-grow-1">
                                <h5 class="card-title">${this.escapeHtml(show.name)}</h5>
                                <p class="card-text text-muted small mb-2">
                                    ${this.escapeHtml(show.original_name || '')}
                                </p>
                                <div class="mb-2">
                                    ${this.renderStatusBadge(show.status)}
                                    ${show.vote_average ? `<span class="badge bg-warning text-dark">
                                        ${show.vote_average.toFixed(1)} <i class="bi bi-star-fill"></i>
                                    </span>` : ''}
                                </div>
                                <p class="card-text small text-muted">
                                    <i class="bi bi-calendar"></i> ${this.formatDate(show.first_air_date)}
                                </p>
                            </div>
                        </div>
                    </div>
                    <div class="card-footer bg-transparent">
                        <div class="btn-group w-100">
                            <a href="/show_detail.html?id=${show.id}" class="btn btn-sm btn-outline-primary">
                                <i class="bi bi-eye"></i> 详情
                            </a>
                            <button class="btn btn-sm btn-outline-secondary" onclick="todayPage.refreshShow(${show.id})">
                                <i class="bi bi-arrow-clockwise"></i> 刷新
                            </button>
                        </div>
                    </div>
                </div>
            `;
            container.appendChild(col);
        });
    }

    updateStats() {
        const total = this.shows.length;
        const newShows = this.shows.filter(s => s.status === 'Returning Series').length;
        const returning = this.shows.filter(s => s.status === 'Ended').length;
        const ended = this.shows.filter(s => s.status === 'Canceled').length;

        document.getElementById('totalCount').textContent = total;
        document.getElementById('newCount').textContent = newShows;
        document.getElementById('returningCount').textContent = returning;
        document.getElementById('endedCount').textContent = ended;
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
        this.loadTodayUpdates();
    }

    async loadWeekUpdates() {
        this.showLoading(true);
        try {
            const response = await api.getWeeklyMarkdown();
            // 解析markdown并显示
            this.showSuccess('本周更新加载成功');
        } catch (error) {
            this.showError('加载失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    async loadMonthUpdates() {
        this.showLoading(true);
        try {
            const response = await api.publishMonthly();
            if (response.code === 0) {
                this.shows = response.data.shows || [];
                this.renderShows();
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
                this.loadTodayUpdates();
            } else {
                this.showError('刷新失败: ' + response.message);
            }
        } catch (error) {
            this.showError('刷新失败: ' + error.message);
        }
    }

    async publishToTelegraph() {
        if (!confirm('确定要发布今日更新到Telegraph吗?')) return;

        try {
            this.showLoading(true);
            const response = await api.publishToday();

            if (response.code === 0 && response.data.success) {
                this.showSuccess('发布成功!');
                
                if (response.data.url) {
                    this.showPublishModal(response.data.url);
                }
            } else {
                this.showError('发布失败: ' + (response.message || '未知错误'));
            }
        } catch (error) {
            this.showError('发布失败: ' + error.message);
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
document.addEventListener('DOMContentLoaded', () => {
    todayPage = new TodayPage();
});
