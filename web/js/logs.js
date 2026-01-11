/**
 * Logs Page Logic
 * 爬取日志页面的核心逻辑
 */

class LogsPage {
    constructor() {
        this.currentPage = 1;
        this.pageSize = 50;
        this.totalPages = 1;
        this.totalCount = 0;
        this.status = '';
        this.operation = '';
        this.logs = [];
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadLogs();
    }

    bindEvents() {
        // 状态过滤
        document.getElementById('statusFilter').addEventListener('change', (e) => {
            this.status = e.target.value;
            this.currentPage = 1;
            this.loadLogs();
        });

        // 操作过滤
        document.getElementById('operationFilter').addEventListener('change', (e) => {
            this.operation = e.target.value;
            this.currentPage = 1;
            this.loadLogs();
        });

        // 每页大小
        document.getElementById('pageSizeSelect').addEventListener('change', (e) => {
            this.pageSize = parseInt(e.target.value);
            this.currentPage = 1;
            this.loadLogs();
        });

        // 刷新按钮
        document.getElementById('refreshLogsBtn').addEventListener('click', () => {
            this.loadLogs();
        });

        // 导出日志
        document.getElementById('exportLogsBtn').addEventListener('click', () => {
            this.exportLogs();
        });
    }

    async loadLogs() {
        this.showLoading(true);

        try {
            const response = await api.getCrawlLogs(
                this.currentPage,
                this.pageSize,
                this.status
            );

            if (response.code === 0) {
                this.logs = response.data.items || [];
                this.totalCount = response.data.total || 0;
                this.totalPages = Math.ceil(this.totalCount / this.pageSize);

                this.renderTable();
                this.renderPagination();
                this.updateStats();
            } else {
                this.showError('加载日志失败: ' + response.message);
            }
        } catch (error) {
            this.showError('加载日志失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    renderTable() {
        const tbody = document.getElementById('logsTableBody');
        tbody.innerHTML = '';

        if (this.logs.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="text-center">暂无日志记录</td></tr>';
            return;
        }

        this.logs.forEach(log => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${log.id || '-'}</td>
                <td>${this.formatDateTime(log.created_at)}</td>
                <td>${this.renderOperation(log.operation)}</td>
                <td>${log.show_id || '-'}</td>
                <td>${log.tmdb_id || '-'}</td>
                <td>${this.renderStatus(log.status)}</td>
                <td>${this.escapeHtml(log.message || '-')}</td>
            `;
            tbody.appendChild(tr);
        });
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
                this.loadLogs();
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
                this.loadLogs();
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
                this.loadLogs();
            }
        });
        pagination.appendChild(nextLi);
    }

    updateStats() {
        const totalCount = this.totalCount;
        const successCount = this.logs.filter(l => l.status === 'success').length;
        const failedCount = this.logs.filter(l => l.status === 'failed').length;
        const successRate = totalCount > 0 ? ((successCount / totalCount) * 100).toFixed(1) : 0;

        document.getElementById('totalCount').textContent = totalCount;
        document.getElementById('successCount').textContent = successCount;
        document.getElementById('failedCount').textContent = failedCount;
        document.getElementById('successRate').textContent = successRate + '%';
    }

    exportLogs() {
        if (this.logs.length === 0) {
            this.showError('没有可导出的日志');
            return;
        }

        // 创建CSV内容
        const headers = ['ID', '时间', '操作', '剧集ID', 'TMDB ID', '状态', '消息'];
        const rows = this.logs.map(log => [
            log.id || '',
            this.formatDateTime(log.created_at),
            log.operation || '',
            log.show_id || '',
            log.tmdb_id || '',
            log.status || '',
            log.message || ''
        ]);

        const csvContent = [
            headers.join(','),
            ...rows.map(row => row.map(cell => `"${cell}"`).join(','))
        ].join('\n');

        // 创建下载链接
        const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `crawl_logs_${new Date().toISOString().split('T')[0]}.csv`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        this.showSuccess('日志导出成功');
    }

    showLoading(show) {
        const spinner = document.getElementById('loadingSpinner');
        const table = document.getElementById('logsTable');
        
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

    renderStatus(status) {
        const badges = {
            'success': '<span class="badge bg-success">成功</span>',
            'failed': '<span class="badge bg-danger">失败</span>',
            'pending': '<span class="badge bg-warning text-dark">进行中</span>'
        };
        return badges[status] || `<span class="badge bg-secondary">${status || '未知'}</span>`;
    }

    renderOperation(operation) {
        const labels = {
            'crawl': '爬取',
            'refresh': '刷新',
            'batch': '批量',
            'init': '初始化'
        };
        return labels[operation] || operation || '-';
    }

    formatDateTime(dateStr) {
        if (!dateStr) return '-';
        const date = new Date(dateStr);
        return date.toLocaleString('zh-CN');
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// 初始化页面
let logsPage;
document.addEventListener('DOMContentLoaded', () => {
    logsPage = new LogsPage();
});
