/**
 * Backup page functionality
 */

// Load backup status on page load
document.addEventListener('DOMContentLoaded', async () => {
    await loadBackupStatus();
    setupEventListeners();
});

/**
 * Load and display backup status
 */
async function loadBackupStatus() {
    try {
        const response = await api.get('/backup/status');
        if (response.code === 0) {
            const stats = response.data.stats;
            document.getElementById('showsCount').textContent = stats.shows || 0;
            document.getElementById('episodesCount').textContent = stats.episodes || 0;
            document.getElementById('crawlLogsCount').textContent = stats.crawl_logs || 0;
            document.getElementById('telegraphPostsCount').textContent = stats.telegraph_posts || 0;

            if (response.data.last_backup) {
                const lastBackup = new Date(response.data.last_backup);
                document.getElementById('lastBackupInfo').innerHTML =
                    `<small class="text-muted">上次备份: ${lastBackup.toLocaleString('zh-CN')}</small>`;
            }
        }
    } catch (error) {
        console.error('Failed to load backup status:', error);
        feedback.error.show('加载备份状态失败: ' + error.message);
    }
}

/**
 * Setup event listeners
 */
function setupEventListeners() {
    // Export button
    document.getElementById('exportBtn').addEventListener('click', handleExport);

    // Import form
    document.getElementById('importForm').addEventListener('submit', handleImport);

    // File input change - show warning
    document.getElementById('backupFile').addEventListener('change', (e) => {
        const warning = document.getElementById('importWarning');
        if (e.target.files.length > 0) {
            warning.classList.remove('d-none');
        } else {
            warning.classList.add('d-none');
        }
    });
}

/**
 * Handle export button click
 */
async function handleExport() {
    const btn = document.getElementById('exportBtn');
    const originalText = btn.innerHTML;

    try {
        // Show loading state
        btn.disabled = true;
        btn.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>导出中...';

        // Call export API
        const response = await fetch('/api/v1/backup/export', {
            method: 'GET',
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('导出失败');
        }

        // Get filename from Content-Disposition header
        const contentDisposition = response.headers.get('Content-Disposition');
        let filename = 'tmdb-backup.json';
        if (contentDisposition) {
            const match = contentDisposition.match(/filename="(.+)"/);
            if (match) {
                filename = match[1];
            }
        }

        // Download file
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);

        feedback.success.show('备份导出成功');
        await loadBackupStatus();
    } catch (error) {
        console.error('Export failed:', error);
        feedback.error.show('导出失败: ' + error.message);
    } finally {
        btn.disabled = false;
        btn.innerHTML = originalText;
    }
}

/**
 * Handle import form submission
 */
async function handleImport(e) {
    e.preventDefault();

    const fileInput = document.getElementById('backupFile');
    const mode = document.getElementById('importMode').value;
    const btn = document.getElementById('importBtn');

    if (!fileInput.files.length) {
        feedback.error.show('请选择备份文件');
        return;
    }

    const file = fileInput.files[0];

    // Validate file size (50MB)
    if (file.size > 50 * 1024 * 1024) {
        feedback.error.show('文件过大 (最大 50MB)');
        return;
    }

    // Validate file type
    if (!file.name.endsWith('.json')) {
        feedback.error.show('仅支持 JSON 格式备份文件');
        return;
    }

    // Confirm for replace mode
    if (mode === 'replace') {
        if (!confirm('警告: 替换模式将清空所有现有数据!\n\n确定要继续吗?')) {
            return;
        }
    }

    const originalText = btn.innerHTML;

    try {
        // Show loading state
        btn.disabled = true;
        btn.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>导入中...';

        // Create FormData
        const formData = new FormData();
        formData.append('file', file);
        formData.append('mode', mode);

        // Call import API
        const response = await fetch('/api/v1/backup/import', {
            method: 'POST',
            credentials: 'include',
            body: formData
        });

        const data = await response.json();

        if (!response.ok || data.code !== 0) {
            throw new Error(data.message || '导入失败');
        }

        // Show result modal
        showImportResult(true, data.data);

        // Refresh status
        await loadBackupStatus();

        // Reset form
        document.getElementById('importForm').reset();
        document.getElementById('importWarning').classList.add('d-none');

    } catch (error) {
        console.error('Import failed:', error);
        showImportResult(false, null, error.message);
    } finally {
        btn.disabled = false;
        btn.innerHTML = originalText;
    }
}

/**
 * Show import result modal
 */
function showImportResult(success, result, errorMessage = '') {
    const modal = document.getElementById('importResultModal');
    const header = document.getElementById('importResultHeader');
    const body = document.getElementById('importResultBody');

    if (success) {
        header.className = 'modal-header bg-success text-white';
        body.innerHTML = `
            <div class="text-center">
                <i class="bi bi-check-circle" style="font-size: 3rem;"></i>
                <h5 class="mt-3">导入成功!</h5>
                <hr>
                <div class="row text-start">
                    <div class="col-6"><strong>剧集:</strong></div>
                    <div class="col-6">${result.shows_imported}</div>
                    <div class="col-6"><strong>集数:</strong></div>
                    <div class="col-6">${result.episodes_imported}</div>
                    <div class="col-6"><strong>爬取日志:</strong></div>
                    <div class="col-6">${result.crawl_logs_imported}</div>
                    <div class="col-6"><strong>Telegraph文章:</strong></div>
                    <div class="col-6">${result.telegraph_posts_imported}</div>
                    ${result.conflicts_skipped > 0 ? `
                        <div class="col-12 mt-2">
                            <div class="alert alert-warning mb-0">
                                <i class="bi bi-exclamation-triangle"></i>
                                跳过 ${result.conflicts_skipped} 条冲突记录 (ID已存在)
                            </div>
                        </div>
                    ` : ''}
                </div>
            </div>
        `;
    } else {
        header.className = 'modal-header bg-danger text-white';
        body.innerHTML = `
            <div class="text-center">
                <i class="bi bi-x-circle" style="font-size: 3rem;"></i>
                <h5 class="mt-3">导入失败</h5>
                <p class="text-danger">${errorMessage}</p>
            </div>
        `;
    }

    new bootstrap.Modal(modal).show();
}
