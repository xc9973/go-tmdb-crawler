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

// 导出到全局
window.feedback = {
    progress: new ProgressBar(),
    error: ErrorHandler,
    confirm: ConfirmDialog,
    retry: RetryHandler
};
