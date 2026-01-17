/**
 * Common.js - 合并公共JS文件
 * 包含: auth-check.js, api.js, feedback.js, auth-ui.js
 * 版本: 2.2
 */

// ==================== Feedback Module ====================
/**
 * Feedback Module
 * 提供进度条、错误处理、确认对话框和重试逻辑
 */

class ProgressBar {
    constructor(containerId = 'feedbackProgressContainer') {
        this.containerId = containerId;
        this.total = 0;
        this.current = 0;
        this.success = 0;
        this.failed = 0;
        this.startTime = null;
        // DOM 缓存
        this.elements = {};
    }

    /**
     * 初始化并显示进度条
     */
    start(total, message = '正在处理...') {
        this.total = total;
        this.current = 0;
        this.success = 0;
        this.failed = 0;
        this.startTime = Date.now();

        let container = document.getElementById(this.containerId);
        if (!container) {
            container = document.createElement('div');
            container.id = this.containerId;
            container.className = 'progress-container my-3 d-none';
            document.body.appendChild(container);
        }

        container.innerHTML = `
            <div class="card shadow-sm">
                <div class="card-body">
                    <h6 class="card-title d-flex justify-content-between">
                        <span><i class="bi bi-cpu me-2"></i>${message}</span>
                        <span class="progress-percent">0%</span>
                    </h6>
                    <div class="progress mb-2" style="height: 10px;">
                        <div class="progress-bar progress-bar-striped progress-bar-animated"
                             role="progressbar" style="width: 0%"></div>
                    </div>
                    <div class="d-flex justify-content-between small text-muted">
                        <span class="progress-status">准备中...</span>
                        <span class="progress-counts">
                            成功: <span class="text-success success-count">0</span> |
                            失败: <span class="text-danger fail-count">0</span> |
                            总计: ${total}
                        </span>
                    </div>
                </div>
            </div>
        `;
        container.classList.remove('d-none');

        // 缓存 DOM 元素引用以提升性能
        this.elements.progressBar = container.querySelector('.progress-bar');
        this.elements.percentText = container.querySelector('.progress-percent');
        this.elements.successText = container.querySelector('.success-count');
        this.elements.failText = container.querySelector('.fail-count');
        this.elements.statusText = container.querySelector('.progress-status');
    }

    /**
     * 更新进度
     */
    update(current, success = true) {
        this.current = current;
        if (success) {
            this.success++;
        } else {
            this.failed++;
        }

        const percent = this.total > 0 ? Math.round((this.current / this.total) * 100) : 0;

        // 使用缓存的 DOM 引用
        if (this.elements.progressBar) this.elements.progressBar.style.width = `${percent}%`;
        if (this.elements.percentText) this.elements.percentText.textContent = `${percent}%`;
        if (this.elements.successText) this.elements.successText.textContent = this.success;
        if (this.elements.failText) this.elements.failText.textContent = this.failed;

        if (this.elements.statusText) {
            this.elements.statusText.textContent = `已完成 ${this.current} / ${this.total}`;
        }
    }

    /**
     * 完成并显示总结
     */
    complete(summary = '') {
        if (this.elements.progressBar) {
            this.elements.progressBar.classList.remove('progress-bar-animated', 'progress-bar-striped');
            this.elements.progressBar.classList.add('bg-success');
            this.elements.progressBar.style.width = '100%';
        }

        if (this.elements.statusText) {
            const duration = ((Date.now() - this.startTime) / 1000).toFixed(1);
            this.elements.statusText.textContent = summary || `处理完成！耗时 ${duration}秒`;
        }

        // 3秒后自动隐藏（可选，或者由调用者控制）
        // setTimeout(() => container.classList.add('d-none'), 3000);
    }

    hide() {
        const container = document.getElementById(this.containerId);
        if (container) container.classList.add('d-none');
    }
}

class ErrorHandler {
    /**
     * 将 API 错误映射为友好中文
     */
    static getFriendlyMessage(rawMessage) {
        if (!rawMessage) return '未知错误';

        const errorMap = {
            'Unauthorized': '未登录或登录已过期',
            'Internal Server Error': '服务器内部错误',
            'Network Error': '网络连接失败',
            'Failed to fetch': '网络请求失败，请检查连接',
            'tmdb_id already exists': '该 TMDB ID 已存在',
            'invalid api key': 'API 密钥无效',
            'context deadline exceeded': '请求超时',
            'Resource not found': '资源不存在 (404)',
            'rate limit': '请求过于频繁，请稍后再试',
            'Service Unavailable': '服务暂时不可用',
            'connection refused': '无法连接到服务器',
        };

        for (const [key, value] of Object.entries(errorMap)) {
            if (rawMessage.includes(key)) return value;
        }

        return rawMessage;
    }
}

class ConfirmDialog {
    /**
     * 显示 Bootstrap 模态对话框，返回 Promise
     */
    static show(options = {}) {
        const {
            title = '确认',
            message = '确定要执行此操作吗？',
            confirmText = '确定',
            cancelText = '取消',
            confirmClass = 'btn-primary',
            cancelClass = 'btn-secondary'
        } = options;

        return new Promise((resolve) => {
            let modalEl = document.getElementById('confirmModal');
            if (!modalEl) {
                modalEl = document.createElement('div');
                modalEl.id = 'confirmModal';
                modalEl.className = 'modal fade';
                modalEl.setAttribute('tabindex', '-1');
                document.body.appendChild(modalEl);
            }

            modalEl.innerHTML = `
                <div class="modal-dialog modal-dialog-centered">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">${title}</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                        </div>
                        <div class="modal-body">
                            <p>${message}</p>
                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn ${cancelClass}" data-bs-dismiss="modal">${cancelText}</button>
                            <button type="button" class="btn ${confirmClass}" id="confirmModalBtn">${confirmText}</button>
                        </div>
                    </div>
                </div>
            `;

            const modal = bootstrap.Modal.getOrCreateInstance(modalEl);
            const confirmBtn = modalEl.querySelector('#confirmModalBtn');

            confirmBtn.onclick = () => {
                modal.hide();
                resolve(true);
            };

            modalEl.addEventListener('hidden.bs.modal', () => {
                resolve(false);
                modal.dispose();
            }, { once: true });

            modal.show();
        });
    }
}

class RetryHandler {
    /**
     * 执行带重试逻辑的函数
     */
    static async execute(fn, options = {}) {
        const {
            retries = 3,
            delay = 1000,
            onRetry = null,
            shouldRetry = () => true
        } = options;

        let lastError;
        for (let i = 0; i < retries; i++) {
            try {
                return await fn();
            } catch (error) {
                lastError = error;
                if (i < retries - 1 && shouldRetry(error)) {
                    if (onRetry) onRetry(i + 1, error);
                    const backoff = Math.min(30000, delay * Math.pow(2, i));
                    await new Promise(r => setTimeout(r, backoff)); // 指数退避
                } else {
                    break;
                }
            }
        }
        throw lastError;
    }
}

// ==================== API Client ====================
/**
 * API Client for TMDB Crawler
 * 处理所有与后端API的通信
 */

class APIClient {
    constructor(baseURL = '/api/v1') {
        this.baseURL = baseURL;
        this.isAuthenticated = false;
    }

    /**
     * 检查是否已认证
     * 通过调用session接口验证
     */
    async checkAuth() {
        try {
            const response = await fetch(`${this.baseURL}/auth/session`, {
                method: 'GET',
                credentials: 'include', // 包含cookie
            });

            if (response.ok) {
                const data = await response.json();
                this.isAuthenticated = data.code === 200;
                return this.isAuthenticated;
            }
            return false;
        } catch (error) {
            console.error('检查认证状态失败:', error);
            return false;
        }
    }

    /**
     * 登录
     */
    async login(apiKey) {
        try {
            const response = await fetch(`${this.baseURL}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({ api_key: apiKey }),
            });

            const data = await response.json();

            if (response.ok && data.code === 200) {
                this.isAuthenticated = true;
                return data; // 返回完整的响应数据
            }

            return data; // 返回完整的响应数据，包含错误信息
        } catch (error) {
            console.error('登录失败:', error);
            return { code: 500, message: error.message };
        }
    }

    /**
     * 登出
     */
    async logout() {
        try {
            await fetch(`${this.baseURL}/auth/logout`, {
                method: 'POST',
                credentials: 'include',
            });
        } catch (error) {
            console.error('登出失败:', error);
        } finally {
            this.isAuthenticated = false;
        }
    }

    /**
     * 通用请求方法
     */
    async request(url, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
        };

        const defaultOptions = {
            credentials: 'include', // 包含cookie
            headers,
        };

        const finalOptions = { ...defaultOptions, ...options };
        // 合并headers
        if (options.headers) {
            finalOptions.headers = { ...headers, ...options.headers };
        }

        try {
            // 判断是否为幂等请求（GET/PUT/DELETE 通常安全）
            const isIdempotent = ['GET', 'HEAD', 'PUT', 'DELETE'].includes((options.method || 'GET').toUpperCase());
            const autoRetry = options.retry !== undefined ? options.retry : isIdempotent;

            return await RetryHandler.execute(async () => {
                const response = await fetch(`${this.baseURL}${url}`, finalOptions);

                // 处理401未认证
                if (response.status === 401) {
                    this.isAuthenticated = false;
                    // 触发认证失败事件
                    window.dispatchEvent(new CustomEvent('auth-required'));
                    const error = new Error('Unauthorized');
                    error.status = 401;
                    throw error;
                }

                // 解析 JSON，增强健壮性
                let data;
                const text = await response.text();
                try {
                    data = JSON.parse(text);
                } catch (parseError) {
                    // 如果响应不是 JSON（例如服务器返回 HTML 错误页），使用文本消息
                    data = { message: text || '无法解析响应' };
                }

                if (!response.ok) {
                    const error = new Error(data.message || 'Internal Server Error');
                    error.status = response.status;
                    throw error;
                }

                return data;
            }, {
                shouldRetry: (error) => {
                    // 仅在 autoRetry 为 true 时才重试
                    if (!autoRetry) return false;
                    // 仅重试网络错误 (无 status) 或 5xx 错误
                    return !error.status || error.status >= 500;
                }
            });
        } catch (error) {
            // 转换友好消息
            error.message = ErrorHandler.getFriendlyMessage(error.message);
            console.error('API请求最终错误:', error);
            throw error;
        }
    }

    /**
     * GET请求
     */
    async get(url, params = {}) {
        const queryString = new URLSearchParams(params).toString();
        const fullUrl = queryString ? `${url}?${queryString}` : url;
        return this.request(fullUrl, { method: 'GET' });
    }

    /**
     * POST请求
     */
    async post(url, data = {}) {
        return this.request(url, {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    /**
     * PUT请求
     */
    async put(url, data = {}) {
        return this.request(url, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    /**
     * DELETE请求
     */
    async delete(url) {
        return this.request(url, { method: 'DELETE' });
    }

    // ========== 剧集管理 API ==========

    /**
     * 获取剧集列表
     */
    async getShows(page = 1, pageSize = 25, search = '', status = '') {
        const params = { page, page_size: pageSize };
        if (search) params.search = search;
        if (status) params.status = status;
        return this.get('/shows', params);
    }

    /**
     * 获取单个剧集详情
     */
    async getShow(id) {
        return this.get(`/shows/${id}`);
    }

    /**
     * 添加剧集
     */
    async addShow(data) {
        return this.post('/shows', data);
    }

    /**
     * 更新剧集
     */
    async updateShow(id, data) {
        return this.put(`/shows/${id}`, data);
    }

    /**
     * 删除剧集
     */
    async deleteShow(id) {
        return this.delete(`/shows/${id}`);
    }

    /**
     * 刷新剧集数据
     */
    async refreshShow(id) {
        return this.post(`/shows/${id}/refresh`);
    }

    /**
     * 获取在播剧集
     */
    async getReturningShows() {
        return this.get('/shows/returning');
    }

    /**
     * 获取剧集集数列表
     */
    async getShowEpisodes(id) {
        return this.get(`/shows/${id}/episodes`);
    }

    // ========== 爬虫控制 API ==========

    /**
     * 爬取单个剧集
     */
    async crawlShow(tmdbId) {
        return this.post(`/crawler/show/${tmdbId}`);
    }

    /**
     * 刷新所有剧集
     */
    async refreshAll() {
        return this.post('/crawler/refresh-all');
    }

    /**
     * 获取爬取日志
     */
    async getCrawlLogs(page = 1, pageSize = 25, status = '') {
        const params = { page, page_size: pageSize };
        if (status) params.status = status;
        return this.get('/crawler/logs', params);
    }

    /**
     * 获取爬取状态
     */
    async getCrawlerStatus() {
    	return this.get('/crawler/status');
    }

    /**
      * 搜索TMDB剧集
      */
    async searchTMDB(query, page = 1) {
    	return this.get('/crawler/search/tmdb', { query, page });
    }

    /**
     * 获取今日更新 (集数级别)
     */
    async getTodayUpdates() {
        return this.get('/calendar/today');
    }

    /**
     * 获取日期范围更新
     */
    async getDateRangeUpdates(startDate, endDate) {
        return this.get('/crawler/updates', {
            start_date: startDate,
            end_date: endDate
        });
    }

    // ========== Episode Upload Tracking ==========

    /**
     * 标记剧集已上传
     * POST /api/v1/episodes/:id/uploaded
     */
    async markEpisodeUploaded(episodeId) {
        return this.post(`/episodes/${episodeId}/uploaded`, {});
    }

    /**
     * 取消标记剧集已上传
     * DELETE /api/v1/episodes/:id/uploaded
     */
    async unmarkEpisodeUploaded(episodeId) {
        return this.delete(`/episodes/${episodeId}/uploaded`);
    }

    // ========== 发布 API ==========

    /**
     * 发布今日更新到Telegraph
     */
    async publishToday() {
        return this.post('/publish/today');
    }

    /**
     * 发布日期范围更新
     */
    async publishDateRange(startDate, endDate) {
        return this.post('/publish/range', { start_date: startDate, end_date: endDate });
    }

    /**
     * 发布单个剧集
     */
    async publishShow(id) {
        return this.post(`/publish/show/${id}`);
    }

    /**
     * 发布本周更新
     */
    async publishWeekly() {
        return this.post('/publish/weekly');
    }

    /**
     * 发布本月更新
     */
    async publishMonthly() {
        return this.post('/publish/monthly');
    }

    // ========== Markdown API ==========

    /**
     * 获取今日更新Markdown
     */
    async getTodayMarkdown() {
        const response = await fetch(`${this.baseURL}/publish/markdown/today`, {
            credentials: 'include',
        });
        return response.text();
    }

    /**
     * 获取剧集详情Markdown
     */
    async getShowMarkdown(id) {
        const response = await fetch(`${this.baseURL}/publish/markdown/show/${id}`, {
            credentials: 'include',
        });
        return response.text();
    }

    /**
     * 获取本周更新Markdown
     */
    async getWeeklyMarkdown() {
        const response = await fetch(`${this.baseURL}/publish/markdown/weekly`, {
            credentials: 'include',
        });
        return response.text();
    }

    // ========== Backup API ==========

    /**
     * 导出备份数据
     */
    async exportBackup() {
        const response = await fetch(`${this.baseURL}/backup/export`, {
            method: 'GET',
            credentials: 'include'
        });

        if (!response.ok) {
            const error = new Error('导出失败');
            error.status = response.status;
            throw error;
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

        return { success: true, filename };
    }

    /**
     * 获取备份状态
     */
    async getBackupStatus() {
        return this.get('/backup/status');
    }
}

// 创建全局API客户端实例
const api = new APIClient();

// 导出到全局
window.api = api;
window.feedback = {
    progress: new ProgressBar(),
    error: ErrorHandler,
    confirm: ConfirmDialog,
    retry: RetryHandler
};

// ==================== 认证UI组件 ====================

/**
 * 显示登录模态框
 */
function showLoginModal(message = '') {
    // 检查是否已存在登录模态框
    let modal = document.getElementById('loginModal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'loginModal';
        modal.className = 'modal fade';
        modal.innerHTML = `
            <div class="modal-dialog modal-dialog-centered">
                <div class="modal-content">
                    <div class="modal-header bg-primary text-white">
                        <h5 class="modal-title">
                            <i class="bi bi-shield-lock me-2"></i>管理员登录
                        </h5>
                    </div>
                    <div class="modal-body">
                        <div id="loginError" class="alert alert-danger d-none"></div>
                        <div id="loginMessage" class="alert alert-info d-none"></div>
                        <form id="loginForm">
                            <div class="mb-3">
                                <label for="apiKeyInput" class="form-label">API 密钥</label>
                                <input type="password" class="form-control" id="apiKeyInput"
                                       placeholder="请输入管理员API密钥" required>
                                <div class="form-text">
                                    联系管理员获取API密钥
                                </div>
                            </div>
                            <div class="alert alert-info mb-0">
                                <small>
                                    <i class="bi bi-info-circle me-1"></i>
                                    登录状态将在浏览器关闭后自动清除，更安全
                                </small>
                            </div>
                        </form>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-primary" id="loginBtn">
                            <i class="bi bi-box-arrow-in-right me-2"></i>登录
                        </button>
                    </div>
                </div>
            </div>
        `;
        document.body.appendChild(modal);

        // 绑定登录事件
        document.getElementById('loginBtn').addEventListener('click', handleLogin);
        document.getElementById('apiKeyInput').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                handleLogin();
            }
        });
    }

    // 显示消息
    const msgEl = document.getElementById('loginMessage');
    if (message) {
        msgEl.textContent = message;
        msgEl.classList.remove('d-none');
    } else {
        msgEl.classList.add('d-none');
    }

    // 隐藏错误
    document.getElementById('loginError').classList.add('d-none');

    // 显示模态框
    const bsModal = new bootstrap.Modal(modal, {
        backdrop: 'static',
        keyboard: false
    });
    bsModal.show();
}

/**
 * 处理登录
 */
async function handleLogin() {
    const apiKeyInput = document.getElementById('apiKeyInput');
    const loginBtn = document.getElementById('loginBtn');
    const errorEl = document.getElementById('loginError');

    const apiKey = apiKeyInput.value.trim();

    if (!apiKey) {
        errorEl.textContent = '请输入API密钥';
        errorEl.classList.remove('d-none');
        return;
    }

    // 禁用按钮
    loginBtn.disabled = true;
    loginBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>验证中...';

    try {
        // 使用新的登录API
        const result = await api.login(apiKey);

        if (result.success) {
            // 关闭模态框
            const modal = bootstrap.Modal.getInstance(document.getElementById('loginModal'));
            modal.hide();

            // 触发认证成功事件
            window.dispatchEvent(new CustomEvent('auth-success'));

            // 刷新页面
            location.reload();
        } else {
            errorEl.textContent = result.message || 'API密钥无效,请检查后重试';
            errorEl.classList.remove('d-none');
        }
    } catch (error) {
        errorEl.textContent = '验证失败: ' + error.message;
        errorEl.classList.remove('d-none');
    } finally {
        loginBtn.disabled = false;
        loginBtn.innerHTML = '<i class="bi bi-box-arrow-in-right me-2"></i>登录';
    }
}

/**
 * 检查认证状态,如果未认证则显示登录框
 */
async function checkAuth() {
    // 检查认证状态
    const authenticated = await api.checkAuth();

    if (!authenticated) {
        showLoginModal('请先登录以访问管理功能');
        return false;
    }

    return true;
}

// 更新登录按钮状态
function updateAuthUI() {
    const btn = document.getElementById('loginLogoutBtn');
    if (!btn) return; // 如果页面没有这个按钮,直接返回

    if (api.isAuthenticated) {
        btn.innerHTML = '<i class="bi bi-box-arrow-right"></i> 退出';
        btn.className = 'btn btn-outline-warning btn-sm';
    } else {
        btn.innerHTML = '<i class="bi bi-box-arrow-in-right"></i> 登录';
        btn.className = 'btn btn-outline-light btn-sm';
    }
}

// 处理登录/退出点击
function handleAuthClick() {
    if (api.isAuthenticated) {
        if (confirm('确定要退出登录吗?')) {
            api.logout().then(() => {
                updateAuthUI();
                window.location.href = '/login.html';
            });
        }
    } else {
        // 跳转到登录页
        window.location.href = '/login.html?redirect=' + encodeURIComponent(window.location.href);
    }
}

// 初始化认证UI
async function initAuthUI() {
    // 先检查认证状态
    await api.checkAuth();
    // 更新UI
    updateAuthUI();
}

// 监听认证需求事件
window.addEventListener('auth-required', () => {
    showLoginModal('需要登录认证才能继续操作');
});

// 导出到全局
window.showLoginModal = showLoginModal;
window.handleLogin = handleLogin;
window.checkAuth = checkAuth;
window.updateAuthUI = updateAuthUI;
window.handleAuthClick = handleAuthClick;
window.initAuthUI = initAuthUI;

// ==================== 异步认证检查 ====================
/**
 * 异步认证检查脚本
 * 不阻塞页面渲染,在后台检查认证状态
 */
(function() {
    'use strict';

    // 配置
    const CONFIG = {
        // 认证检查API端点
        authCheckEndpoint: '/api/v1/auth/session',
        // 登录页面路径
        loginPagePath: '/login.html',
        // 不需要认证的页面路径 (仅登录页和欢迎页公开)
        publicPages: [
            '/login.html',
            '/welcome.html'
        ],
        // 是否在控制台输出调试信息
        debug: false
    };

    // 当前页面路径
    const currentPath = window.location.pathname;

    // 检查当前页面是否为公开页面
    function isPublicPage() {
        return CONFIG.publicPages.some(page => currentPath.endsWith(page));
    }

    // 如果是公开页面,不需要检查认证
    if (isPublicPage()) {
        log('当前页面为公开页面,跳过认证检查');
        return;
    }

    // 异步检查认证状态(不阻塞页面渲染)
    async function checkAuthAsync() {
        try {
            const response = await fetch(CONFIG.authCheckEndpoint, {
                method: 'GET',
                credentials: 'include',
            });

            if (response.ok) {
                const data = await response.json();
                const isAuthenticated = data.code === 200 && data.data && data.data.authenticated === true;

                if (!isAuthenticated) {
                    log('未认证,重定向到登录页');
                    redirectToLogin();
                    return false;
                }

                log('已认证,允许访问');
                // 更新全局API客户端的认证状态
                if (window.api) {
                    window.api.isAuthenticated = true;
                }
                // 更新登录按钮UI
                if (typeof updateAuthUI === 'function') {
                    updateAuthUI();
                }
                return true;
            } else {
                log('认证检查失败,重定向到登录页');
                redirectToLogin();
                return false;
            }
        } catch (error) {
            log('认证检查出错:', error);
            // 网络错误时不立即重定向,等待用户操作
            return false;
        }
    }

    // 重定向到登录页
    function redirectToLogin() {
        // 保存当前页面URL,登录后可以返回
        const currentUrl = window.location.href;
        const loginUrl = CONFIG.loginPagePath + '?redirect=' + encodeURIComponent(currentUrl);

        log('重定向到:', loginUrl);
        window.location.href = loginUrl;
    }

    // 调试日志
    function log(...args) {
        if (CONFIG.debug && console.log) {
            console.log('[AuthCheck]', ...args);
        }
    }

    // 页面加载后异步检查认证(不阻塞渲染)
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            // 使用 setTimeout 确保不阻塞页面渲染
            setTimeout(checkAuthAsync, 0);
        });
    } else {
        setTimeout(checkAuthAsync, 0);
    }

    // 导出认证检查函数(供其他脚本使用)
    window.AuthCheck = {
        checkAuth: checkAuthAsync,
        isPublicPage: isPublicPage,
        redirectToLogin: redirectToLogin
    };

})();
