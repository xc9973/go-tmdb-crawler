/**
 * Shows Page Logic
 * 剧集列表页面的核心逻辑
 */

class ShowsPage {
    constructor() {
        this.currentPage = 1;
        this.pageSize = 25;
        this.totalPages = 1;
        this.totalCount = 0;
        this.search = '';
        this.status = '';
        this.sort = 'id';
        this.order = 'asc';
        this.selectedShows = new Set();
        this.shows = [];

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

        // 状态过滤
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
            this.refreshAll();
        });

        // 全选
        document.getElementById('selectAll').addEventListener('change', (e) => {
            this.toggleSelectAll(e.target.checked);
        });

        // 表头排序
        document.querySelectorAll('th[data-sort]').forEach(th => {
            th.addEventListener('click', () => {
                const sortBy = th.dataset.sort;
                if (this.sort === sortBy) {
                    this.order = this.order === 'asc' ? 'desc' : 'asc';
                } else {
                    this.sort = sortBy;
                    this.order = 'asc';
                }
                this.loadShows();
            });
        });

        // 添加剧集
        document.getElementById('saveShowBtn').addEventListener('click', () => {
            this.saveShow();
        });

        // TMDB搜索
        document.getElementById('searchTmdbBtn').addEventListener('click', () => {
            this.searchTMDB();
        });

        // 批量操作
        document.getElementById('batchRefreshBtn')?.addEventListener('click', () => {
            this.batchRefresh();
        });

        document.getElementById('batchDeleteBtn')?.addEventListener('click', () => {
            this.batchDelete();
        });

        document.getElementById('clearSelectionBtn')?.addEventListener('click', () => {
            this.clearSelection();
        });
    }

    async loadShows() {
        this.showLoading(true);

        try {
            const response = await api.getShows(
                this.currentPage,
                this.pageSize,
                this.search,
                this.status
            );

            if (response.code === 0) {
                this.shows = response.data.items;
                this.totalCount = response.data.total;
                this.totalPages = Math.ceil(this.totalCount / this.pageSize);

                this.renderTable();
                this.renderPagination();
                this.updateStats();
            } else {
                this.showError('加载数据失败: ' + response.message);
            }
        } catch (error) {
            this.showError('加载数据失败: ' + error.message);
        } finally {
            this.showLoading(false);
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
                <td>
                    <input type="checkbox" class="form-check-input show-checkbox" 
                           data-id="${show.id}" ${this.selectedShows.has(show.id) ? 'checked' : ''}>
                </td>
                <td>${show.id}</td>
                <td>
                    <a href="#" class="text-decoration-none" data-show-id="${show.id}">
                        <strong>${this.escapeHtml(show.name)}</strong>
                    </a>
                </td>
                <td>${this.escapeHtml(show.original_name || '-')}</td>
                <td>${this.renderStatusBadge(show.status)}</td>
                <td>${this.formatDate(show.first_air_date)}</td>
                <td>${this.renderRating(show.vote_average)}</td>
                <td>
                    <button class="btn btn-sm btn-outline-primary btn-action" onclick="showsPage.refreshShow(${show.id})">
                        <i class="bi bi-arrow-clockwise"></i>
                    </button>
                    <button class="btn btn-sm btn-outline-danger btn-action" onclick="showsPage.deleteShow(${show.id})">
                        <i class="bi bi-trash"></i>
                    </button>
                </td>
            `;

            // 绑定checkbox事件
            const checkbox = tr.querySelector('.show-checkbox');
            checkbox.addEventListener('change', () => {
                if (checkbox.checked) {
                    this.selectedShows.add(show.id);
                } else {
                    this.selectedShows.delete(show.id);
                }
                this.updateBatchActions();
            });

            tbody.appendChild(tr);
        });

        // 更新全选状态
        const allChecked = this.shows.length > 0 && 
                          this.shows.every(show => this.selectedShows.has(show.id));
        document.getElementById('selectAll').checked = allChecked;
    }

    renderPagination() {
        const pagination = document.getElementById('pagination');
        pagination.innerHTML = '';

        // 上一页
        const prevLi = document.createElement('li');
        prevLi.className = `page-item ${this.currentPage === 1 ? 'disabled' : ''}`;
        prevLi.innerHTML = '<a class="page-link" href="#">&laquo;</a>';
        prevLi.addEventListener('click', (e) => {
            e.preventDefault();
            if (this.currentPage > 1) {
                this.currentPage--;
                this.loadShows();
            }
        });
        pagination.appendChild(prevLi);

        // 页码
        const startPage = Math.max(1, this.currentPage - 2);
        const endPage = Math.min(this.totalPages, this.currentPage + 2);

        for (let i = startPage; i <= endPage; i++) {
            const li = document.createElement('li');
            li.className = `page-item ${i === this.currentPage ? 'active' : ''}`;
            li.innerHTML = `<a class="page-link" href="#">${i}</a>`;
            li.addEventListener('click', (e) => {
                e.preventDefault();
                this.currentPage = i;
                this.loadShows();
            });
            pagination.appendChild(li);
        }

        // 下一页
        const nextLi = document.createElement('li');
        nextLi.className = `page-item ${this.currentPage === this.totalPages ? 'disabled' : ''}`;
        nextLi.innerHTML = '<a class="page-link" href="#">&raquo;</a>';
        nextLi.addEventListener('click', (e) => {
            e.preventDefault();
            if (this.currentPage < this.totalPages) {
                this.currentPage++;
                this.loadShows();
            }
        });
        pagination.appendChild(nextLi);
    }

    updateStats() {
        document.getElementById('totalCount').textContent = this.totalCount;
        document.getElementById('currentPageInfo').textContent = 
            `${this.currentPage}/${this.totalPages}`;

        // 统计状态
        const returningCount = this.shows.filter(s => s.status === 'Returning Series').length;
        const endedCount = this.shows.filter(s => s.status === 'Ended').length;

        document.getElementById('returningCount').textContent = returningCount;
        document.getElementById('endedCount').textContent = endedCount;
    }

    async refreshShow(id) {
        if (!confirm('确定要刷新该剧集数据吗?')) return;

        try {
            this.showLoading(true);
            const response = await api.refreshShow(id);

            if (response.code === 0) {
                this.showSuccess('刷新成功');
                this.loadShows();
            } else {
                this.showError('刷新失败: ' + response.message);
            }
        } catch (error) {
            this.showError('刷新失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    async deleteShow(id) {
        if (!confirm('确定要删除该剧集吗?此操作不可恢复!')) return;

        try {
            this.showLoading(true);
            await api.deleteShow(id);

            this.showSuccess('删除成功');
            this.selectedShows.delete(id);
            this.loadShows();
        } catch (error) {
            this.showError('删除失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    async refreshAll() {
        if (!confirm('确定要刷新所有剧集吗?这可能需要一些时间...')) return;

        try {
            this.showLoading(true);
            const response = await api.refreshAll();

            if (response.code === 0) {
                this.showSuccess(`刷新完成! 共处理 ${response.data.count} 个剧集`);
                this.loadShows();
            } else {
                this.showError('刷新失败: ' + response.message);
            }
        } catch (error) {
            this.showError('刷新失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    async saveShow() {
        const tmdbId = document.getElementById('tmdbId').value;
        const name = document.getElementById('showName').value;

        if (!tmdbId || !name) {
            this.showError('请填写TMDB ID和名称');
            return;
        }

        const data = {
            tmdb_id: parseInt(tmdbId),
            name: name,
            original_name: document.getElementById('originalName').value,
            status: document.getElementById('showStatus').value,
            overview: document.getElementById('showOverview').value
        };

        try {
            const response = await api.addShow(data);

            if (response.code === 0) {
                this.showSuccess('添加成功');
                bootstrap.Modal.getInstance(document.getElementById('addShowModal')).hide();
                this.loadShows();
            } else {
                this.showError('添加失败: ' + response.message);
            }
        } catch (error) {
            this.showError('添加失败: ' + error.message);
        }
    }

    async searchTMDB() {
        const tmdbId = document.getElementById('tmdbId').value;
        if (!tmdbId) {
            this.showError('请输入TMDB ID');
            return;
        }

        try {
            this.showLoading(true);
            const response = await api.searchTMDB(tmdbId);

            if (response.code === 0 && response.data && response.data.results && Array.isArray(response.data.results) && response.data.results.length > 0) {
                const show = response.data.results[0];
                
                // 检查 show 对象是否存在必要的字段
                if (!show || (!show.name && !show.original_name)) {
                    this.showError('TMDB搜索成功，但未返回剧集信息，请稍后重试或手动填写');
                    return;
                }
                
                // 填充表单
                document.getElementById('showName').value = show.name || '';
                document.getElementById('originalName').value = show.original_name || '';
                document.getElementById('showOverview').value = show.overview || '';
                document.getElementById('showStatus').value = 'Returning Series';
                
                this.showSuccess('TMDB搜索成功');
            } else {
                this.showError('TMDB搜索失败: ' + (response.message || '未找到相关剧集'));
            }
        } catch (error) {
            this.showError('TMDB搜索失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    toggleSelectAll(checked) {
        this.shows.forEach(show => {
            if (checked) {
                this.selectedShows.add(show.id);
            } else {
                this.selectedShows.delete(show.id);
            }
        });

        document.querySelectorAll('.show-checkbox').forEach(checkbox => {
            checkbox.checked = checked;
        });

        this.updateBatchActions();
    }

    updateBatchActions() {
        const batchActions = document.getElementById('batchActions');
        const selectedCount = document.getElementById('selectedCount');

        if (this.selectedShows.size > 0) {
            batchActions.style.display = 'block';
            selectedCount.textContent = this.selectedShows.size;
        } else {
            batchActions.style.display = 'none';
        }
    }

    async batchRefresh() {
        if (!confirm(`确定要刷新选中的 ${this.selectedShows.size} 个剧集吗?`)) return;

        try {
            this.showLoading(true);
            // 批量刷新逻辑
            for (const id of this.selectedShows) {
                await api.refreshShow(id);
            }

            this.showSuccess('批量刷新完成');
            this.clearSelection();
            this.loadShows();
        } catch (error) {
            this.showError('批量刷新失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    async batchDelete() {
        if (!confirm(`确定要删除选中的 ${this.selectedShows.size} 个剧集吗?此操作不可恢复!`)) return;

        try {
            this.showLoading(true);
            for (const id of this.selectedShows) {
                await api.deleteShow(id);
            }

            this.showSuccess('批量删除完成');
            this.clearSelection();
            this.loadShows();
        } catch (error) {
            this.showError('批量删除失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    clearSelection() {
        this.selectedShows.clear();
        document.getElementById('selectAll').checked = false;
        document.querySelectorAll('.show-checkbox').forEach(cb => cb.checked = false);
        this.updateBatchActions();
    }

    showLoading(show) {
        const spinner = document.getElementById('loadingSpinner');
        const table = document.getElementById('showsTable');
        
        if (show) {
            spinner.style.display = 'block';
            table.style.opacity = '0.5';
        } else {
            spinner.style.display = 'none';
            table.style.opacity = '1';
        }
    }

    showSuccess(message) {
        this.showToast(message, 'success');
    }

    showError(message) {
        this.showToast(message, 'danger');
    }

    showInfo(message) {
        this.showToast(message, 'info');
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

    renderRating(rating) {
        if (!rating) return '-';
        const stars = Math.round(rating / 2);
        return `${rating.toFixed(1)} <i class="bi bi-star-fill rating-stars"></i>`;
    }

    formatDate(dateStr) {
        if (!dateStr) return '-';
        return dateStr.split('T')[0];
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
}

// 初始化页面
let showsPage;
document.addEventListener('DOMContentLoaded', () => {
    showsPage = new ShowsPage();
});
