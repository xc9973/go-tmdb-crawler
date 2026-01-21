/**
 * TMDB Crawler - Minimal Bundle
 * 精简合并版 - 包含核心功能
 */

// ==================== API Client ====================
class APIClient {
    constructor(baseURL = '/api/v1') {
        this.baseURL = baseURL;
    }

    async request(url, options = {}) {
        const headers = { 'Content-Type': 'application/json' };
        const opts = { credentials: 'include', headers, ...options };

        const response = await fetch(`${this.baseURL}${url}`, opts);
        if (response.status === 401) {
            window.location.href = '/login.html?redirect=' + encodeURIComponent(window.location.href);
            throw new Error('Unauthorized');
        }

        const data = await response.json();
        if (!response.ok) throw new Error(data.message || '请求失败');
        return data;
    }

    get(url, params = {}) {
        const qs = new URLSearchParams(params).toString();
        return this.request(qs ? `${url}?${qs}` : url, { method: 'GET' });
    }

    post(url, data) {
        return this.request(url, { method: 'POST', body: JSON.stringify(data) });
    }

    delete(url) {
        return this.request(url, { method: 'DELETE' });
    }

    // Shows API
    getShows(page, pageSize, search, status) {
        return this.get('/shows', { page, page_size: pageSize, search, status });
    }

    getShow(id) { return this.get(`/shows/${id}`); }
    addShow(data) { return this.post('/shows', data); }
    deleteShow(id) { return this.delete(`/shows/${id}`); }
    refreshShow(id) { return this.post(`/shows/${id}/refresh`); }
    refreshAll() { return this.post('/crawler/refresh-all'); }
    crawlShow(tmdbId) { return this.post(`/crawler/show/${tmdbId}`); }
}

const api = new APIClient();
window.api = api;

// ==================== Toast Notification ====================
function showToast(message, type = 'info') {
    const container = document.getElementById('toastContainer') || createToastContainer();
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = message;
    container.appendChild(toast);

    setTimeout(() => {
        toast.style.opacity = '1';
    }, 10);

    setTimeout(() => {
        toast.style.opacity = '0';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

function createToastContainer() {
    const container = document.createElement('div');
    container.id = 'toastContainer';
    container.className = 'toast-container';
    document.body.appendChild(container);
    return container;
}

// ==================== Shows Page ====================
class ShowsPage {
    constructor() {
        this.currentPage = 1;
        this.pageSize = 25;
        this.search = '';
        this.status = '';
        this.sort = 'id';
        this.order = 'asc';
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadShows();
    }

    bindEvents() {
        // 搜索
        document.getElementById('searchInput').addEventListener('input',
            this.debounce((e) => {
                this.search = e.target.value;
                this.currentPage = 1;
                this.loadShows();
            }, 500)
        );

        // 状态筛选
        document.getElementById('statusFilter').addEventListener('change', (e) => {
            this.status = e.target.value;
            this.currentPage = 1;
            this.loadShows();
        });

        // 每页大小
        document.getElementById('pageSizeSelect').addEventListener('change', (e) => {
            this.pageSize = parseInt(e.target.value);
            this.currentPage = 1;
            this.loadShows();
        });

        // 刷新全部
        document.getElementById('refreshAllBtn').addEventListener('click', () => {
            if (confirm('确定要刷新所有剧集吗？')) this.refreshAll();
        });

        // 排序
        document.querySelectorAll('th[data-sort]').forEach(th => {
            th.addEventListener('click', () => {
                const sortBy = th.dataset.sort;
                this.order = (this.sort === sortBy) ? (this.order === 'asc' ? 'desc' : 'asc') : 'asc';
                this.sort = sortBy;
                this.loadShows();
            });
        });

        // 添加剧集
        document.getElementById('saveShowBtn').addEventListener('click', () => this.saveShow());
        document.getElementById('searchTmdbBtn').addEventListener('click', () => this.searchTMDB());
    }

    async loadShows() {
        document.getElementById('loadingSpinner').style.display = 'block';
        document.getElementById('showsTable').style.opacity = '0.5';

        try {
            const response = await api.getShows(this.currentPage, this.pageSize, this.search, this.status);

            if (response.code === 0) {
                this.shows = response.data.items;
                this.totalCount = response.data.total;
                this.totalPages = Math.ceil(this.totalCount / this.pageSize);
                this.renderTable();
                this.renderPagination();
                this.updateStats();
            }
        } catch (error) {
            showToast('加载数据失败: ' + error.message, 'error');
        } finally {
            document.getElementById('loadingSpinner').style.display = 'none';
            document.getElementById('showsTable').style.opacity = '1';
        }
    }

    renderTable() {
        const tbody = document.getElementById('showsTableBody');
        tbody.innerHTML = '';

        if (this.shows.length === 0) {
            tbody.innerHTML = '<tr><td colspan="8" class="text-center">暂无数据</td></tr>';
            return;
        }

        this.shows.forEach(show => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td><input type="checkbox" class="form-check-input" data-id="${show.id}"></td>
                <td>${show.id}</td>
                <td><a href="show_detail.html?id=${show.id}" class="show-link">${this.escapeHtml(show.name)}</a></td>
                <td>${this.escapeHtml(show.original_name || '-')}</td>
                <td>${this.renderStatusBadge(show.status)}</td>
                <td>${show.first_air_date?.split('T')[0] || '-'}</td>
                <td>${show.vote_average?.toFixed(1) || '-'}</td>
                <td>
                    <button class="btn btn-sm" onclick="showsPage.refreshShow(${show.id})">
                        <i class="bi bi-arrow-clockwise"></i>
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="showsPage.deleteShow(${show.id})">
                        <i class="bi bi-trash"></i>
                    </button>
                </td>
            `;
            tbody.appendChild(tr);
        });
    }

    renderPagination() {
        const pagination = document.getElementById('pagination');
        pagination.innerHTML = '';

        const prevDisabled = this.currentPage === 1 ? 'disabled' : '';
        pagination.innerHTML += `<a href="#" class="page-link ${prevDisabled}" onclick="event.preventDefault();if(!${this.currentPage===1}){showsPage.currentPage--;showsPage.loadShows()}">&laquo;</a>`;

        const startPage = Math.max(1, this.currentPage - 2);
        const endPage = Math.min(this.totalPages, this.currentPage + 2);

        for (let i = startPage; i <= endPage; i++) {
            const active = i === this.currentPage ? 'active' : '';
            pagination.innerHTML += `<a href="#" class="page-link ${active}" onclick="event.preventDefault();showsPage.currentPage=${i};showsPage.loadShows()">${i}</a>`;
        }

        const nextDisabled = this.currentPage === this.totalPages ? 'disabled' : '';
        pagination.innerHTML += `<a href="#" class="page-link ${nextDisabled}" onclick="event.preventDefault();if(${this.currentPage<this.totalPages}){showsPage.currentPage++;showsPage.loadShows()}">&raquo;</a>`;
    }

    updateStats() {
        document.getElementById('totalCount').textContent = this.totalCount;
        document.getElementById('currentPageInfo').textContent = `${this.currentPage}/${this.totalPages}`;
        document.getElementById('returningCount').textContent = this.shows.filter(s => s.status === 'Returning Series').length;
        document.getElementById('endedCount').textContent = this.shows.filter(s => s.status === 'Ended').length;
    }

    async refreshShow(id) {
        try {
            await api.refreshShow(id);
            showToast('刷新成功', 'success');
            this.loadShows();
        } catch (error) {
            showToast('刷新失败: ' + error.message, 'error');
        }
    }

    async deleteShow(id) {
        if (!confirm('确定要删除该剧集吗？')) return;

        try {
            await api.deleteShow(id);
            showToast('删除成功', 'success');
            this.loadShows();
        } catch (error) {
            showToast('删除失败: ' + error.message, 'error');
        }
    }

    async refreshAll() {
        try {
            const response = await api.refreshAll();
            showToast(`刷新完成！处理 ${response.data.count} 个剧集`, 'success');
            this.loadShows();
        } catch (error) {
            showToast('刷新失败: ' + error.message, 'error');
        }
    }

    async saveShow() {
        const tmdbId = document.getElementById('tmdbId').value.trim();
        const name = document.getElementById('showName').value.trim();

        if (!tmdbId || !name) {
            showToast('请填写TMDB ID和名称', 'error');
            return;
        }

        try {
            const response = await api.addShow({
                tmdb_id: parseInt(tmdbId),
                name,
                original_name: document.getElementById('originalName').value,
                status: document.getElementById('showStatus').value,
                overview: document.getElementById('showOverview').value
            });

            if (response.code === 0) {
                showToast('添加成功', 'success');
                this.closeModal();
                this.loadShows();
            }
        } catch (error) {
            showToast('添加失败: ' + error.message, 'error');
        }
    }

    async searchTMDB() {
        const tmdbId = document.getElementById('tmdbId').value.trim();
        if (!tmdbId) {
            showToast('请输入TMDB ID', 'error');
            return;
        }

        const btn = document.getElementById('searchTmdbBtn');
        btn.disabled = true;
        btn.innerHTML = '<i class="bi bi-hourglass-split"></i> 搜索中...';

        try {
            const response = await api.crawlShow(parseInt(tmdbId));

            if (response.code === 0 && response.data) {
                const show = response.data;
                document.getElementById('showName').value = show.name || '';
                document.getElementById('originalName').value = show.original_name || '';
                document.getElementById('showStatus').value = show.status || '';
                document.getElementById('showOverview').value = show.overview || '';
                showToast('TMDB搜索成功', 'success');
            }
        } catch (error) {
            showToast('TMDB搜索失败: ' + error.message, 'error');
        } finally {
            btn.disabled = false;
            btn.innerHTML = '<i class="bi bi-search"></i> 搜索';
        }
    }

    closeModal() {
        document.getElementById('addShowModal').classList.remove('show');
        document.getElementById('tmdbId').value = '';
        document.getElementById('showName').value = '';
        document.getElementById('originalName').value = '';
        document.getElementById('showStatus').value = '';
        document.getElementById('showOverview').value = '';
    }

    renderStatusBadge(status) {
        const badges = {
            'Returning Series': '<span class="badge badge-returning">连载中</span>',
            'Ended': '<span class="badge badge-ended">已完结</span>',
            'Canceled': '<span class="badge badge-canceled">已取消</span>'
        };
        return badges[status] || `<span class="badge">${status || '未知'}</span>`;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    debounce(func, wait) {
        let timeout;
        return function(...args) {
            clearTimeout(timeout);
            timeout = setTimeout(() => func.apply(this, args), wait);
        };
    }
}

// ==================== Theme Toggle ====================
function initTheme() {
    const theme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', theme);

    document.getElementById('themeToggle').addEventListener('click', () => {
        const current = document.documentElement.getAttribute('data-theme');
        const newTheme = current === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
    });
}

// ==================== Modal ====================
function initModal() {
    document.addEventListener('click', (e) => {
        const trigger = e.target.closest('[data-modal-target]');
        const close = e.target.closest('[data-modal-close]');

        if (trigger) {
            const modal = document.querySelector(trigger.getAttribute('data-modal-target'));
            if (modal) modal.classList.add('show');
        }

        if (close) {
            const modal = close.closest('.modal');
            if (modal) modal.classList.remove('show');
        }
    });
}

// ==================== Mobile Nav ====================
function initMobileNav() {
    const toggle = document.getElementById('navbarToggle');
    const nav = document.getElementById('navbarNav');

    toggle?.addEventListener('click', () => {
        nav?.classList.toggle('show');
    });
}

// ==================== Initialize ====================
document.addEventListener('DOMContentLoaded', () => {
    window.showsPage = new ShowsPage();
    initTheme();
    initModal();
    initMobileNav();
});
