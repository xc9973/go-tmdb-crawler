/**
 * 通用认证检查脚本
 * 在所有需要认证的页面中引入此脚本
 * 
 * 使用方法:
 * 1. 在页面中引入此脚本: <script src="js/auth-check.js"></script>
 * 2. 脚本会自动检查认证状态,未认证时重定向到登录页
 * 3. 登录成功后会自动返回原页面
 */

(function() {
    'use strict';
    
    // 配置
    const CONFIG = {
        // 认证检查API端点
        authCheckEndpoint: '/api/v1/auth/session',
        // 登录页面路径
        loginPagePath: '/login.html',
        // 不需要认证的页面路径
        publicPages: [
            '/login.html',
            '/welcome.html',
            '/index.html',
            '/show_detail.html',
            '/today.html'
        ],
        // 认证检查间隔(毫秒),0表示只检查一次
        checkInterval: 0,
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
    
    // 检查认证状态
    async function checkAuth() {
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
                return true;
            } else {
                log('认证检查失败,重定向到登录页');
                redirectToLogin();
                return false;
            }
        } catch (error) {
            log('认证检查出错:', error);
            // 网络错误时,为了安全起见,重定向到登录页
            redirectToLogin();
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
    
    // 页面加载时立即检查认证
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', checkAuth);
    } else {
        checkAuth();
    }
    
    // 如果需要定期检查认证状态
    if (CONFIG.checkInterval > 0) {
        setInterval(checkAuth, CONFIG.checkInterval);
    }
    
    // 导出认证检查函数(供其他脚本使用)
    window.AuthCheck = {
        checkAuth: checkAuth,
        isPublicPage: isPublicPage,
        redirectToLogin: redirectToLogin
    };
    
})();

/**
 * 使用示例:
 * 
 * 1. 在HTML中引入:
 *    <script src="js/auth-check.js"></script>
 * 
 * 2. 在其他脚本中使用:
 *    if (window.AuthCheck) {
 *        await window.AuthCheck.checkAuth();
 *    }
 * 
 * 3. 手动触发认证检查:
 *    window.AuthCheck.redirectToLogin();
 */
