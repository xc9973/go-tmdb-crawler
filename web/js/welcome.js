// Welcome Page JavaScript

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    // 检查是否已登录
    checkAuthAndRedirect();
    
    // 初始化登录按钮
    initLoginButton();
    
    // 初始化登录表单
    initLoginForm();
    
    // 初始化密码显示/隐藏切换
    initPasswordToggle();
});

// 检查认证状态并重定向
async function checkAuthAndRedirect() {
    try {
        const response = await api.checkAuth();
        if (response && response.code === 200) {
            // 已登录，重定向到主页
            window.location.href = '/';
        }
    } catch (error) {
        // 未登录，保持在欢迎页面
        console.log('未登录，显示欢迎页面');
    }
}

// 初始化登录按钮
function initLoginButton() {
    const loginBtn = document.getElementById('loginBtn');
    if (loginBtn) {
        loginBtn.addEventListener('click', function() {
            showLoginModal();
        });
    }
}

// 显示登录模态框
function showLoginModal() {
    const modal = new bootstrap.Modal(document.getElementById('loginModal'));
    modal.show();
    
    // 清空之前的错误信息
    hideLoginError();
    
    // 聚焦到输入框
    setTimeout(() => {
        document.getElementById('apiKey').focus();
    }, 500);
}

// 初始化登录表单
function initLoginForm() {
    const submitBtn = document.getElementById('submitLogin');
    const loginForm = document.getElementById('loginForm');
    
    if (submitBtn && loginForm) {
        submitBtn.addEventListener('click', handleLogin);
        
        // 支持回车键提交
        loginForm.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                e.preventDefault();
                handleLogin();
            }
        });
    }
}

// 处理登录
async function handleLogin() {
    const apiKeyInput = document.getElementById('apiKey');
    const submitBtn = document.getElementById('submitLogin');
    const apiKey = apiKeyInput.value.trim();
    
    // 验证输入
    if (!apiKey) {
        showLoginError('请输入API密钥');
        apiKeyInput.focus();
        return;
    }
    
    // 禁用提交按钮
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>登录中...';
    
    try {
        // 调用登录API
        const response = await api.login(apiKey);
        
        if (response && response.code === 200) {
            // 登录成功
            hideLoginError();
            showToast('登录成功，正在跳转...', 'success');
            
            // 延迟跳转到主页
            setTimeout(() => {
                window.location.href = '/';
            }, 1000);
        } else {
            // 登录失败
            showLoginError(response?.message || '登录失败，请检查API密钥');
            submitBtn.disabled = false;
            submitBtn.innerHTML = '<i class="bi bi-box-arrow-in-right me-2"></i>登录';
        }
    } catch (error) {
        // 登录异常
        console.error('登录错误:', error);
        showLoginError('登录失败，请稍后重试');
        submitBtn.disabled = false;
        submitBtn.innerHTML = '<i class="bi bi-box-arrow-in-right me-2"></i>登录';
    }
}

// 显示登录错误
function showLoginError(message) {
    const errorDiv = document.getElementById('loginError');
    const errorText = document.getElementById('loginErrorText');
    
    if (errorDiv && errorText) {
        errorText.textContent = message;
        errorDiv.classList.remove('d-none');
        
        // 添加抖动动画
        errorDiv.style.animation = 'shake 0.5s';
        setTimeout(() => {
            errorDiv.style.animation = '';
        }, 500);
    }
}

// 隐藏登录错误
function hideLoginError() {
    const errorDiv = document.getElementById('loginError');
    if (errorDiv) {
        errorDiv.classList.add('d-none');
    }
}

// 初始化密码显示/隐藏切换
function initPasswordToggle() {
    const toggleBtn = document.getElementById('togglePassword');
    const apiKeyInput = document.getElementById('apiKey');
    
    if (toggleBtn && apiKeyInput) {
        toggleBtn.addEventListener('click', function() {
            const type = apiKeyInput.getAttribute('type') === 'password' ? 'text' : 'password';
            apiKeyInput.setAttribute('type', type);
            
            // 切换图标
            const icon = toggleBtn.querySelector('i');
            if (type === 'text') {
                icon.classList.remove('bi-eye');
                icon.classList.add('bi-eye-slash');
            } else {
                icon.classList.remove('bi-eye-slash');
                icon.classList.add('bi-eye');
            }
        });
    }
}

// 显示提示消息
function showToast(message, type = 'info') {
    const toastContainer = document.getElementById('toastContainer');
    if (!toastContainer) return;
    
    const toastId = 'toast-' + Date.now();
    const toastClass = type === 'success' ? 'toast-success' : 
                      type === 'error' ? 'toast-error' : 'toast-info';
    
    const icon = type === 'success' ? 'bi-check-circle' : 
                 type === 'error' ? 'bi-exclamation-circle' : 'bi-info-circle';
    
    const toastHTML = `
        <div id="${toastId}" class="toast ${toastClass}" role="alert" aria-live="assertive" aria-atomic="true">
            <div class="toast-header">
                <i class="bi ${icon} me-2"></i>
                <strong class="me-auto">${type === 'success' ? '成功' : type === 'error' ? '错误' : '提示'}</strong>
                <button type="button" class="btn-close" data-bs-dismiss="toast" aria-label="Close"></button>
            </div>
            <div class="toast-body">
                ${message}
            </div>
        </div>
    `;
    
    toastContainer.insertAdjacentHTML('beforeend', toastHTML);
    
    const toastElement = document.getElementById(toastId);
    const toast = new bootstrap.Toast(toastElement, {
        delay: 3000
    });
    
    toast.show();
    
    // 自动移除
    toastElement.addEventListener('hidden.bs.toast', function() {
        toastElement.remove();
    });
}

// 添加抖动动画样式
const style = document.createElement('style');
style.textContent = `
    @keyframes shake {
        0%, 100% { transform: translateX(0); }
        10%, 30%, 50%, 70%, 90% { transform: translateX(-5px); }
        20%, 40%, 60%, 80% { transform: translateX(5px); }
    }
`;
document.head.appendChild(style);
