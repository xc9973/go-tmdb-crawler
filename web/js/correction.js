/**
 * Correction Page Logic
 * Handles stale show detection and correction UI
 */

class CorrectionPage {
    constructor() {
        this.staleShows = [];
        this.init();
    }

    async init() {
        // Bind events
        document.getElementById('runCorrectionBtn').addEventListener('click', () => this.runDetection());
        document.getElementById('runDetectionBtn').addEventListener('click', () => this.runDetection());

        // Initial load
        await this.loadStatus();
    }

    async loadStatus() {
        try {
            const data = await api.getCorrectionStatus();
            if (data.code === 200) {
                this.updateHealthCard(data.data);
                this.staleShows = data.data.stale_shows || [];
                this.renderStaleTable();
            }
        } catch (error) {
            console.error('Failed to load status:', error);
        }
    }

    updateHealthCard(status) {
        document.getElementById('totalCount').textContent = status.total_shows || 0;
        document.getElementById('staleCount').textContent = status.stale_count || 0;
        const normalCount = (status.total_shows || 0) - (status.stale_count || 0);
        document.getElementById('normalCount').textContent = normalCount;

        // Update last check time
        const now = new Date();
        document.getElementById('lastCheck').textContent = `上次检测: ${now.toLocaleTimeString()}`;
    }

    async runDetection() {
        const btn = document.getElementById('runCorrectionBtn');
        const spinner = document.getElementById('loadingSpinner');

        btn.disabled = true;
        spinner.style.display = 'block';

        try {
            const data = await api.runCorrectionNow();
            if (data.code === 200 || data.code === 202) {
                this.staleShows = data.data.stale_shows || [];
                this.updateHealthCard(data.data);
                this.renderStaleTable();
                this.showToast(`检测完成：发现 ${this.staleShows.length} 个过期剧集`, 'success');
            }
        } catch (error) {
            this.showToast('检测失败: ' + error.message, 'error');
        } finally {
            btn.disabled = false;
            spinner.style.display = 'none';
        }
    }

    renderStaleTable() {
        const tbody = document.getElementById('staleTableBody');

        if (this.staleShows.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="text-center text-muted">暂无过期剧集</td></tr>';
            return;
        }

        tbody.innerHTML = this.staleShows.map(show => `
            <tr>
                <td><strong>${show.show_name}</strong></td>
                <td><img src="${show.poster_path ? 'https://image.tmdb.org/t/p/w92' + show.poster_path : '/css/placeholder.png'}" width="46" style="border-radius: 4px;"></td>
                <td>${show.normal_interval} 天</td>
                <td>${new Date(show.latest_episode_date).toLocaleDateString()}</td>
                <td class="warning"><strong>${show.days_overdue}</strong> 天</td>
                <td><span class="badge bg-warning">过期</span></td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="correctionPage.refreshShow(${show.show_id}, ${show.tmdb_id})">
                        <i class="bi bi-arrow-clockwise"></i> 刷新
                    </button>
                    <button class="btn btn-sm" onclick="correctionPage.clearStale(${show.show_id})">
                        <i class="bi bi-x"></i> 忽略
                    </button>
                </td>
            </tr>
        `).join('');
    }

    async refreshShow(showId, tmdbId) {
        try {
            await api.refreshStaleShow(showId);
            this.showToast('刷新任务已创建', 'success');
            await this.loadStatus();
        } catch (error) {
            this.showToast('刷新失败: ' + error.message, 'error');
        }
    }

    async clearStale(showId) {
        if (!confirm('确定要清除过期标记吗？')) return;

        try {
            await api.clearStaleFlag(showId);
            this.showToast('过期标记已清除', 'success');
            await this.loadStatus();
        } catch (error) {
            this.showToast('操作失败: ' + error.message, 'error');
        }
    }

    showToast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.textContent = message;
        toast.style.opacity = '1';
        container.appendChild(toast);

        setTimeout(() => {
            toast.style.opacity = '0';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }
}

// Initialize page
const correctionPage = new CorrectionPage();
