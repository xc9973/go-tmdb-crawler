/**
 * Modal Management
 * 处理添加和编辑剧集的模态框逻辑
 */

class ShowModal {
    constructor() {
        this.currentShow = null;
        this.isEditing = false;
        this.init();
    }

    init() {
        this.bindEvents();
    }

    bindEvents() {
        // TMDB搜索按钮
        document.getElementById('searchTmdbBtn').addEventListener('click', () => {
            this.searchTMDB();
        });

        // TMDB ID输入框回车事件
        document.getElementById('tmdbId').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                this.searchTMDB();
            }
        });

        // 保存按钮
        document.getElementById('saveShowBtn').addEventListener('click', () => {
            this.saveShow();
        });

        // 模态框关闭时重置表单
        const modal = document.getElementById('addShowModal');
        modal.addEventListener('hidden.bs.modal', () => {
            this.resetForm();
        });
    }

    /**
     * 打开添加剧集模态框
     */
    openAddModal() {
        this.isEditing = false;
        this.currentShow = null;
        
        document.querySelector('#addShowModal .modal-title').innerHTML = 
            '<i class="bi bi-plus-circle"></i> 添加剧集';
        document.getElementById('saveShowBtn').innerHTML = 
            '<i class="bi bi-save"></i> 保存';
        
        const modal = new bootstrap.Modal(document.getElementById('addShowModal'));
        modal.show();
    }

    /**
     * 打开编辑剧集模态框
     */
    async openEditModal(showId) {
        this.isEditing = true;
        
        try {
            const response = await api.getShow(showId);
            if (response.code === 0) {
                this.currentShow = response.data;
                this.fillForm(this.currentShow);
                
                document.querySelector('#addShowModal .modal-title').innerHTML = 
                    '<i class="bi bi-pencil"></i> 编辑剧集';
                document.getElementById('saveShowBtn').innerHTML = 
                    '<i class="bi bi-save"></i> 更新';
                
                const modal = new bootstrap.Modal(document.getElementById('addShowModal'));
                modal.show();
            } else {
                showsPage.showError('加载剧集信息失败: ' + response.message);
            }
        } catch (error) {
            showsPage.showError('加载剧集信息失败: ' + error.message);
        }
    }

    /**
     * 填充表单
     */
    fillForm(show) {
        document.getElementById('tmdbId').value = show.tmdb_id || '';
        document.getElementById('showName').value = show.name || '';
        document.getElementById('originalName').value = show.original_name || '';
        document.getElementById('showStatus').value = show.status || '';
        document.getElementById('showOverview').value = show.overview || '';
    }

    /**
     * 从TMDB搜索并填充表单
     */
    async searchTMDB() {
        const tmdbId = document.getElementById('tmdbId').value.trim();
        
        if (!tmdbId) {
            showsPage.showError('请输入TMDB ID');
            return;
        }

        const searchBtn = document.getElementById('searchTmdbBtn');
        const originalHtml = searchBtn.innerHTML;
        searchBtn.disabled = true;
        searchBtn.innerHTML = '<span class="spinner-border spinner-border-sm"></span> 搜索中...';

        try {
            // 调用爬虫API搜索TMDB
            const response = await api.crawlShow(parseInt(tmdbId));
            
            if (response.code === 0) {
                const show = response.data;
                
                // 自动填充表单
                document.getElementById('showName').value = show.name || '';
                document.getElementById('originalName').value = show.original_name || '';
                document.getElementById('showStatus').value = show.status || '';
                document.getElementById('showOverview').value = show.overview || '';
                
                showsPage.showSuccess('TMDB搜索成功,已自动填充信息');
            } else {
                showsPage.showError('TMDB搜索失败: ' + response.message);
            }
        } catch (error) {
            showsPage.showError('TMDB搜索失败: ' + error.message);
        } finally {
            searchBtn.disabled = false;
            searchBtn.innerHTML = originalHtml;
        }
    }

    /**
     * 保存剧集
     */
    async saveShow() {
        const tmdbId = document.getElementById('tmdbId').value.trim();
        const name = document.getElementById('showName').value.trim();

        // 验证必填字段
        if (!tmdbId) {
            showsPage.showError('请输入TMDB ID');
            document.getElementById('tmdbId').focus();
            return;
        }

        if (!name) {
            showsPage.showError('请输入剧集名称');
            document.getElementById('showName').focus();
            return;
        }

        const data = {
            tmdb_id: parseInt(tmdbId),
            name: name,
            original_name: document.getElementById('originalName').value.trim(),
            status: document.getElementById('showStatus').value,
            overview: document.getElementById('showOverview').value.trim()
        };

        const saveBtn = document.getElementById('saveShowBtn');
        const originalHtml = saveBtn.innerHTML;
        saveBtn.disabled = true;
        saveBtn.innerHTML = '<span class="spinner-border spinner-border-sm"></span> 保存中...';

        try {
            let response;
            
            if (this.isEditing && this.currentShow) {
                // 更新现有剧集
                response = await api.updateShow(this.currentShow.id, data);
            } else {
                // 添加新剧集
                response = await api.addShow(data);
            }

            if (response.code === 0) {
                showsPage.showSuccess(this.isEditing ? '更新成功' : '添加成功');
                
                // 关闭模态框
                const modal = bootstrap.Modal.getInstance(document.getElementById('addShowModal'));
                modal.hide();
                
                // 刷新列表
                showsPage.loadShows();
            } else {
                showsPage.showError((this.isEditing ? '更新' : '添加') + '失败: ' + response.message);
            }
        } catch (error) {
            showsPage.showError((this.isEditing ? '更新' : '添加') + '失败: ' + error.message);
        } finally {
            saveBtn.disabled = false;
            saveBtn.innerHTML = originalHtml;
        }
    }

    /**
     * 重置表单
     */
    resetForm() {
        document.getElementById('tmdbId').value = '';
        document.getElementById('showName').value = '';
        document.getElementById('originalName').value = '';
        document.getElementById('showStatus').value = '';
        document.getElementById('showOverview').value = '';
        
        this.currentShow = null;
        this.isEditing = false;
    }

    /**
     * 显示表单验证错误
     */
    showFieldError(fieldId, message) {
        const field = document.getElementById(fieldId);
        const feedback = field.nextElementSibling;
        
        if (feedback && feedback.classList.contains('invalid-feedback')) {
            feedback.textContent = message;
            feedback.style.display = 'block';
        }
        
        field.classList.add('is-invalid');
    }

    /**
     * 清除表单验证错误
     */
    clearFieldError(fieldId) {
        const field = document.getElementById(fieldId);
        const feedback = field.nextElementSibling;
        
        if (feedback && feedback.classList.contains('invalid-feedback')) {
            feedback.style.display = 'none';
        }
        
        field.classList.remove('is-invalid');
    }

    /**
     * 清除所有表单验证错误
     */
    clearAllErrors() {
        const fields = ['tmdbId', 'showName', 'originalName', 'showStatus', 'showOverview'];
        fields.forEach(fieldId => this.clearFieldError(fieldId));
    }
}

// 创建全局实例
let showModal;
document.addEventListener('DOMContentLoaded', () => {
    showModal = new ShowModal();
    window.showModal = showModal;
});
