/**
 * 认证UI组件
 * 统一管理所有页面的登录/退出按钮状态
 */

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

// 导出到全局
window.updateAuthUI = updateAuthUI;
window.handleAuthClick = handleAuthClick;
window.initAuthUI = initAuthUI;

// 页面加载时自动初始化
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initAuthUI);
} else {
    initAuthUI();
}
