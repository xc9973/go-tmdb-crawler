/**
 * 登录页面逻辑
 * 处理用户认证、会话管理和页面跳转
 */

// 页面加载完成后执行
document.addEventListener('DOMContentLoaded', async function() {
    // 检查是否已经登录
    const isAuthenticated = await checkAuthentication();
    
    if (isAuthenticated) {
        // 已登录,重定向到首页
        redirectToHome();
        return;
    }

    // 初始化登录表单
    initLoginForm();
    
    // 初始化密码显示/隐藏功能
    initPasswordToggle();
    
    // 从URL获取错误消息(如果有的话)
    const urlParams = new URLSearchParams(window.location.search);
    const errorMsg = urlParams.get('error');
    if (errorMsg) {
        showError(decodeURIComponent(errorMsg));
    }
});

/**
 * 检查认证状态
 */
async function checkAuthentication() {
    try {
        const response = await fetch('/api/v1/auth/session', {
            method: 'GET',
            credentials: 'include',
        });
        
        if (response.ok) {
            const data = await response.json();
            return data.code === 200 && data.data && data.data.authenticated === true;
        }
        return false;
    } catch (error) {
        console.error('检查认证状态失败:', error);
        return false;
    }
}

/**
 * 初始化登录表单
 */
function initLoginForm() {
    const loginForm = document.getElementById('loginForm');
    const apiKeyInput = document.getElementById('apiKeyInput');
    
    // 表单提交处理
    loginForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const apiKey = apiKeyInput.value.trim();
        const rememberMe = document.getElementById('rememberMeCheck').checked;
        
        if (!apiKey) {
            showError('请输入API密钥');
            return;
        }
        
        await handleLogin(apiKey, rememberMe);
    });
    
    // 自动聚焦到输入框
    setTimeout(() => {
        apiKeyInput.focus();
    }, 100);
}

/**
 * 处理登录
 */
async function handleLogin(apiKey, rememberMe) {
    const loginBtn = document.getElementById('loginBtn');
    const loginBtnText = document.getElementById('loginBtnText');
    const originalBtnText = loginBtnText.textContent;
    
    // 禁用按钮,显示加载状态
    loginBtn.disabled = true;
    loginBtnText.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>登录中...';
    
    // 清除之前的消息
    clearMessages();
    
    try {
        const response = await fetch('/api/v1/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({
                api_key: apiKey,
                remember_me: rememberMe
            }),
        });
        
        const data = await response.json();
        
        if (response.ok && data.code === 200) {
            // 登录成功
            showSuccess('登录成功!正在跳转...');
            
            // 延迟跳转,让用户看到成功消息
            setTimeout(() => {
                // 获取重定向URL(如果有)
                const urlParams = new URLSearchParams(window.location.search);
                const redirect = urlParams.get('redirect') || '/';
                
                window.location.href = redirect;
            }, 500);
        } else {
            // 登录失败
            const errorMsg = data.message || 'API密钥无效,请检查后重试';
            showError(errorMsg);
            
            // 重新启用按钮
            loginBtn.disabled = false;
            loginBtnText.textContent = originalBtnText;
        }
    } catch (error) {
        console.error('登录失败:', error);
        showError('网络错误,请检查连接后重试');
        
        // 重新启用按钮
        loginBtn.disabled = false;
        loginBtnText.textContent = originalBtnText;
    }
}

/**
 * 初始化密码显示/隐藏功能
 */
function initPasswordToggle() {
    const toggleBtn = document.getElementById('togglePasswordBtn');
    const toggleIcon = document.getElementById('togglePasswordIcon');
    const apiKeyInput = document.getElementById('apiKeyInput');
    
    toggleBtn.addEventListener('click', function() {
        const type = apiKeyInput.getAttribute('type') === 'password' ? 'text' : 'password';
        apiKeyInput.setAttribute('type', type);
        
        // 切换图标
        if (type === 'text') {
            toggleIcon.classList.remove('bi-eye');
            toggleIcon.classList.add('bi-eye-slash');
            toggleBtn.setAttribute('title', '隐藏密钥');
        } else {
            toggleIcon.classList.remove('bi-eye-slash');
            toggleIcon.classList.add('bi-eye');
            toggleBtn.setAttribute('title', '显示密钥');
        }
    });
}

/**
 * 显示错误消息
 */
function showError(message) {
    const errorAlert = document.getElementById('errorAlert');
    const errorText = document.getElementById('errorText');
    
    errorText.textContent = message;
    errorAlert.classList.remove('d-none');
    
    // 隐藏成功消息
    document.getElementById('successAlert').classList.add('d-none');
    
    // 3秒后自动隐藏
    setTimeout(() => {
        errorAlert.classList.add('d-none');
    }, 5000);
}

/**
 * 显示成功消息
 */
function showSuccess(message) {
    const successAlert = document.getElementById('successAlert');
    const successText = document.getElementById('successText');
    
    successText.textContent = message;
    successAlert.classList.remove('d-none');
    
    // 隐藏错误消息
    document.getElementById('errorAlert').classList.add('d-none');
}

/**
 * 清除所有消息
 */
function clearMessages() {
    document.getElementById('errorAlert').classList.add('d-none');
    document.getElementById('successAlert').classList.add('d-none');
}

/**
 * 重定向到首页
 */
function redirectToHome() {
    // 获取原始请求的URL(如果有)
    const urlParams = new URLSearchParams(window.location.search);
    const redirect = urlParams.get('redirect') || '/';
    
    window.location.href = redirect;
}

/**
 * 键盘快捷键支持
 */
document.addEventListener('keydown', function(e) {
    // ESC键清空输入
    if (e.key === 'Escape') {
        const apiKeyInput = document.getElementById('apiKeyInput');
        if (apiKeyInput === document.activeElement) {
            apiKeyInput.value = '';
            clearMessages();
        }
    }
});
