/**
 * Show Detail Page Logic
 * 剧集详情页面的核心逻辑
 */

class ShowDetailPage {
    constructor() {
        this.showId = null;
        this.show = null;
        this.episodes = [];
        this.init();
    }

    init() {
        // 从URL获取show ID
        const urlParams = new URLSearchParams(window.location.search);
        this.showId = urlParams.get('id');

        if (!this.showId) {
            this.showError('缺少剧集ID参数');
            return;
        }

        this.bindEvents();
        this.loadShowDetail();
    }

    bindEvents() {
        // 刷新按钮
        document.getElementById('refreshShowBtn').addEventListener('click', () => {
            this.refreshShow();
        });

        // 导出Markdown
        document.getElementById('exportMarkdownBtn').addEventListener('click', () => {
            this.exportMarkdown();
        });

        // 发布到Telegraph
        document.getElementById('publishTelegraphBtn').addEventListener('click', () => {
            this.publishToTelegraph();
        });
    }

    async loadShowDetail() {
        this.showLoading(true);

        try {
            // 加载剧集详情
            const showResponse = await api.getShow(this.showId);
            
            if (showResponse.code !== 0) {
                throw new Error(showResponse.message || '加载失败');
            }

            this.show = showResponse.data;
            
            // 渲染剧集信息
            this.renderShowInfo();
            
            // 加载集数列表
            await this.loadEpisodes();
            
            // 加载爬取历史
            this.loadCrawlHistory();
            
            // 显示内容
            document.getElementById('showDetailContent').style.display = 'block';
        } catch (error) {
            this.showError('加载剧集详情失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    async loadEpisodes() {
        try {
            const response = await api.getShowEpisodes(this.showId);
            
            if (response.code === 0) {
                this.episodes = response.data.seasons || [];
                this.renderEpisodes();
            } else {
                console.error('加载集数失败:', response.message);
                this.renderEpisodes(); // 渲染空状态
            }
        } catch (error) {
            console.error('加载集数失败:', error);
            this.renderEpisodes(); // 渲染空状态
        }
    }

    renderShowInfo() {
        // 基本信息
        document.getElementById('showName').textContent = this.show.name;
        document.getElementById('showOriginalName').textContent = this.show.original_name || '';
        document.getElementById('showFirstAirDate').textContent = this.formatDate(this.show.first_air_date);
        document.getElementById('showGenres').textContent = this.show.genres || '-';
        document.getElementById('showTmdbId').textContent = this.show.tmdb_id;
        document.getElementById('showOverview').textContent = this.show.overview || '暂无简介';

        // 状态徽章
        const statusBadge = document.getElementById('showStatus');
        statusBadge.innerHTML = this.renderStatusBadge(this.show.status);

        // 评分
        const ratingBadge = document.getElementById('showRating');
        if (this.show.vote_average) {
            ratingBadge.innerHTML = `评分: ${this.show.vote_average.toFixed(1)} <i class="bi bi-star-fill"></i>`;
        } else {
            ratingBadge.style.display = 'none';
        }

        // 海报
        const poster = document.getElementById('showPoster');
        if (this.show.poster_path) {
            poster.src = `https://image.tmdb.org/t/p/w500${this.show.poster_path}`;
        } else {
            poster.src = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="300" height="450"%3E%3Crect fill="%23ddd" width="300" height="450"/%3E%3Ctext fill="%23999" x="50%25" y="50%25" text-anchor="middle" dy=".3em"%3E无海报%3C/text%3E%3C/svg%3E';
        }
    }

    renderEpisodes() {
        const seasonTabs = document.getElementById('seasonTabs');
        const episodesContent = document.getElementById('episodesContent');
        
        seasonTabs.innerHTML = '';
        episodesContent.innerHTML = '';

        if (!this.episodes || this.episodes.length === 0) {
            // 显示空状态
            const emptyDiv = document.createElement('div');
            emptyDiv.className = 'text-center text-muted py-5';
            emptyDiv.innerHTML = '<i class="bi bi-inbox fs-1"></i><p class="mt-3">暂无集数数据</p>';
            episodesContent.appendChild(emptyDiv);
            return;
        }

        // 按季度编号排序
        this.episodes.sort((a, b) => a.season_number - b.season_number);

        this.episodes.forEach((season, index) => {
            // 季度标签
            const tabItem = document.createElement('li');
            tabItem.className = 'nav-item';
            tabItem.innerHTML = `
                <button class="nav-link ${index === 0 ? 'active' : ''}"
                        data-bs-toggle="tab"
                        data-bs-target="#season-${season.season_number}"
                        type="button">
                    第${season.season_number}季 <span class="badge bg-secondary">${season.episode_count}</span>
                </button>
            `;
            seasonTabs.appendChild(tabItem);

            // 剧集内容
            const contentDiv = document.createElement('div');
            contentDiv.className = `tab-pane fade ${index === 0 ? 'show active' : ''}`;
            contentDiv.id = `season-${season.season_number}`;
            
            let tableHTML = `
                <div class="table-responsive">
                    <table class="table table-sm table-hover">
                        <thead>
                            <tr>
                                <th width="12%">集数</th>
                                <th width="38%">名称</th>
                                <th width="18%">播出日期</th>
                                <th width="17%">评分</th>
                                <th width="15%">更新时间</th>
                            </tr>
                        </thead>
                        <tbody>
            `;

            if (season.episodes && season.episodes.length > 0) {
                season.episodes.forEach(ep => {
                    const episodeCode = `S${season.season_number}E${ep.episode_number}`;
                    tableHTML += `
                        <tr>
                            <td><strong>${episodeCode}</strong></td>
                            <td>
                                ${this.escapeHtml(ep.name)}
                                ${ep.overview ? `<small class="text-muted d-block">${this.escapeHtml(ep.overview.substring(0, 100))}${ep.overview.length > 100 ? '...' : ''}</small>` : ''}
                            </td>
                            <td>${this.formatDate(ep.air_date)}</td>
                            <td>
                                ${ep.vote_average ? `
                                    <span class="badge bg-warning text-dark">
                                        ${ep.vote_average.toFixed(1)} <i class="bi bi-star-fill"></i>
                                    </span>
                                ` : '-'}
                            </td>
                            <td><small class="text-muted">${this.formatDateTime(ep.updated_at)}</small></td>
                        </tr>
                    `;
                });
            } else {
                tableHTML += `
                    <tr>
                        <td colspan="5" class="text-center text-muted">
                            暂无数据
                        </td>
                    </tr>
                `;
            }

            tableHTML += `
                        </tbody>
                    </table>
                </div>
            `;

            contentDiv.innerHTML = tableHTML;
            episodesContent.appendChild(contentDiv);
        });
    }

    async loadCrawlHistory() {
        try {
            const response = await api.getCrawlLogs(1, 10);
            
            if (response.code === 0) {
                const logs = response.data.items || [];
                this.renderCrawlHistory(logs);
            }
        } catch (error) {
            console.error('加载爬取历史失败:', error);
        }
    }

    renderCrawlHistory(logs) {
        const tbody = document.getElementById('crawlHistoryBody');
        tbody.innerHTML = '';

        if (logs.length === 0) {
            tbody.innerHTML = '<tr><td colspan="4" class="text-center text-muted">暂无爬取记录</td></tr>';
            return;
        }

        logs.forEach(log => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${this.formatDateTime(log.created_at)}</td>
                <td>${log.operation || '-'}</td>
                <td>${this.renderLogStatus(log.status)}</td>
                <td>${log.message || '-'}</td>
            `;
            tbody.appendChild(tr);
        });
    }

    async refreshShow() {
        if (!confirm('确定要刷新该剧集数据吗?')) return;

        try {
            this.showLoading(true);
            const response = await api.refreshShow(this.showId);

            if (response.code === 0) {
                this.showSuccess('刷新成功');
                await this.loadShowDetail();
            } else {
                this.showError('刷新失败: ' + response.message);
            }
        } catch (error) {
            this.showError('刷新失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    async exportMarkdown() {
        try {
            const markdown = await api.getShowMarkdown(this.showId);
            
            // 创建下载链接
            const blob = new Blob([markdown], { type: 'text/markdown' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${this.show.name}.md`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

            this.showSuccess('Markdown导出成功');
        } catch (error) {
            this.showError('导出失败: ' + error.message);
        }
    }

    async publishToTelegraph() {
        if (!confirm('确定要发布到Telegraph吗?')) return;

        try {
            this.showLoading(true);
            const response = await api.publishShow(this.showId);

            if (response.code === 0 && response.data.success) {
                this.showSuccess('发布成功!');
                
                // 显示Telegraph链接
                if (response.data.url) {
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
                                    <p>文章已成功发布到Telegraph!</p>
                                    <div class="input-group">
                                        <input type="text" class="form-control" value="${response.data.url}" readonly>
                                        <button class="btn btn-outline-secondary" onclick="window.open('${response.data.url}', '_blank')">
                                            打开
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    `;
                    document.body.appendChild(modal);
                    const bsModal = new bootstrap.Modal(modal);
                    bsModal.show();
                    modal.addEventListener('hidden.bs.modal', () => {
                        modal.remove();
                    });
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

    showLoading(show) {
        const spinner = document.getElementById('loadingSpinner');
        const content = document.getElementById('showDetailContent');
        
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

    renderLogStatus(status) {
        const badges = {
            'success': '<span class="badge bg-success">成功</span>',
            'failed': '<span class="badge bg-danger">失败</span>',
            'pending': '<span class="badge bg-warning text-dark">进行中</span>'
        };
        return badges[status] || `<span class="badge bg-secondary">${status}</span>`;
    }

    formatDate(dateStr) {
        if (!dateStr) return '-';
        return dateStr.split('T')[0];
    }

    formatDateTime(dateStr) {
        if (!dateStr) return '-';
        const date = new Date(dateStr);
        return date.toLocaleString('zh-CN');
    }
}

// 初始化页面
let showDetailPage;
document.addEventListener('DOMContentLoaded', () => {
    showDetailPage = new ShowDetailPage();
});
