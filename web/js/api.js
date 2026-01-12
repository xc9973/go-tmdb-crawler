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
                return { success: true, data: data.data };
            }
            
            return { success: false, message: data.message };
        } catch (error) {
            console.error('登录失败:', error);
            return { success: false, message: error.message };
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
            const response = await fetch(`${this.baseURL}${url}`, finalOptions);
            
            // 处理401未认证
            if (response.status === 401) {
                this.isAuthenticated = false;
                // 触发认证失败事件
                window.dispatchEvent(new CustomEvent('auth-required'));
                throw new Error('需要登录认证');
            }
            
            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || '请求失败');
            }

            return data;
        } catch (error) {
            console.error('API请求错误:', error);
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
}

// 创建全局API客户端实例
const api = new APIClient();

// 导出到全局
window.api = api;

// ========== 认证UI组件 ==========

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

// 监听认证需求事件
window.addEventListener('auth-required', () => {
    showLoginModal('需要登录认证才能继续操作');
});

// 导出认证函数
window.showLoginModal = showLoginModal;
window.checkAuth = checkAuth;
