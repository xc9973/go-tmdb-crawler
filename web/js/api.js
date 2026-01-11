/**
 * API Client for TMDB Crawler
 * 处理所有与后端API的通信
 */

class APIClient {
    constructor(baseURL = '/api/v1') {
        this.baseURL = baseURL;
    }

    /**
     * 通用请求方法
     */
    async request(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
            },
        };

        const finalOptions = { ...defaultOptions, ...options };

        try {
            const response = await fetch(`${this.baseURL}${url}`, finalOptions);
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
        const response = await fetch(`${this.baseURL}/publish/markdown/today`);
        return response.text();
    }

    /**
     * 获取剧集详情Markdown
     */
    async getShowMarkdown(id) {
        const response = await fetch(`${this.baseURL}/publish/markdown/show/${id}`);
        return response.text();
    }

    /**
     * 获取本周更新Markdown
     */
    async getWeeklyMarkdown() {
        const response = await fetch(`${this.baseURL}/publish/markdown/weekly`);
        return response.text();
    }
}

// 创建全局API客户端实例
const api = new APIClient();

// 导出到全局
window.api = api;
