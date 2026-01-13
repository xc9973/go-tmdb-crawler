# 认证功能文档

## 📋 概述

本文档描述了TMDB剧集管理系统的认证功能实现。系统采用**强制认证**机制,确保所有管理页面都需要用户登录后才能访问。

## 🔒 安全特性

### 1. 强制认证
- **首次访问**: 用户首次访问系统时,自动重定向到登录页面
- **会话验证**: 所有页面加载前都会验证用户认证状态
- **自动跳转**: 未认证用户无法访问任何管理页面

### 2. Cookie安全
- **HttpOnly**: 防止XSS攻击窃取Cookie
- **SameSite=Strict**: 防止CSRF攻击
- **会话管理**: 支持"记住我"功能,最长30天有效期

### 3. 认证流程
```
用户访问页面
    ↓
检查Cookie中的session_token
    ↓
Token不存在或无效 → 重定向到 /login.html
    ↓
用户输入API Key
    ↓
验证API Key
    ↓
创建Session并设置HttpOnly Cookie
    ↓
重定向到原页面
    ↓
用户正常访问
```

## 📁 文件结构

### 前端文件
```
web/
├── login.html              # 登录页面
├── css/
│   └── login.css          # 登录页面样式
└── js/
    ├── login.js           # 登录逻辑
    └── auth-check.js      # 通用认证检查脚本
```

### 后端文件
```
api/
└── auth.go                # 认证API处理器

middleware/
└── auth.go                # 认证中间件

services/
└── auth.go                # 认证服务
```

## 🚀 使用指南

### 用户登录流程

1. **访问系统**
   - 打开浏览器,访问系统地址(如 `http://localhost:8080`)
   - 系统自动检测到未登录,重定向到登录页面

2. **输入API密钥**
   - 在登录页面输入管理员API密钥
   - 可选择"记住登录状态"(默认勾选,30天有效)
   - 点击"登录"按钮

3. **登录成功**
   - 系统验证API密钥
   - 创建安全的会话Cookie
   - 自动跳转到原访问页面

4. **后续访问**
   - 在Cookie有效期内,无需重新登录
   - 关闭浏览器后重新打开,如果勾选了"记住我",仍然保持登录状态

### 退出登录

点击页面右上角的"退出"按钮,确认后即可退出登录:
- 清除服务器端的会话
- 清除浏览器中的Cookie
- 重定向到登录页面

## 🔧 配置说明

### 环境变量

在 `.env` 文件中配置:

```bash
# 管理员API密钥(必需)
ADMIN_API_KEY=your-secret-key-here

# 是否允许远程访问(可选)
ALLOW_REMOTE_ADMIN=false

# 会话有效期(可选,默认30天)
SESSION_DURATION=720h
```

### 安全建议

1. **生产环境配置**
   ```bash
   # 使用强密码作为API密钥
   ADMIN_API_KEY=$(openssl rand -base64 32)
   
   # 启用HTTPS后,设置secure标志
   COOKIE_SECURE=true
   ```

2. **API密钥管理**
   - 定期更换API密钥
   - 不要将密钥提交到版本控制系统
   - 使用环境变量或密钥管理服务

3. **会话安全**
   - 默认会话有效期为30天
   - 可根据安全需求调整
   - 建议启用"记住我"功能,提升用户体验

## 📊 API接口

### 1. 登录接口

**请求**
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "api_key": "your-api-key",
  "remember_me": true
}
```

**响应**
```json
{
  "code": 200,
  "message": "登录成功",
  "success": true,
  "data": {
    "token": "jwt-token-here",
    "expires_at": "2024-02-13T02:12:44Z",
    "session_id": "abc123",
    "is_first_login": true
  }
}
```

### 2. 获取会话信息

**请求**
```http
GET /api/v1/auth/session
Cookie: session_token=jwt-token-here
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "authenticated": true,
    "session_id": "abc123",
    "created_at": "2024-01-13T02:12:44Z",
    "expires_at": "2024-02-13T02:12:44Z",
    "last_active": "2024-01-13T02:12:44Z"
  }
}
```

### 3. 登出接口

**请求**
```http
POST /api/v1/auth/logout
Cookie: session_token=jwt-token-here
```

**响应**
```json
{
  "code": 200,
  "message": "登出成功"
}
```

## 🛡️ 安全机制

### 1. 防止XSS攻击
- 使用HttpOnly Cookie,JavaScript无法访问
- Content Security Policy (CSP) 头部保护
- 输入验证和输出转义

### 2. 防止CSRF攻击
- SameSite=Strict Cookie属性
- 验证请求来源
- 使用JWT token验证

### 3. 防止暴力破解
- 登录失败次数限制(5次)
- IP封禁机制(30分钟)
- 失败尝试记录

### 4. 会话管理
- JWT token签名验证
- 会话过期自动失效
- 支持主动登出

## 📱 用户体验

### 登录页面特性

1. **简洁设计**
   - 渐变背景,现代化UI
   - 响应式布局,支持移动端
   - 深色模式自动适配

2. **交互优化**
   - 密码显示/隐藏切换
   - 键盘快捷键支持(Enter登录,ESC清空)
   - 加载状态提示

3. **错误处理**
   - 清晰的错误提示
   - 自动聚焦输入框
   - 5秒后自动隐藏错误消息

### 认证检查

1. **自动检查**
   - 页面加载时自动验证认证状态
   - 未认证时自动重定向到登录页
   - 登录成功后返回原页面

2. **透明体验**
   - 已登录用户无感知
   - Cookie过期时友好提示
   - 支持多设备同时登录

## 🔍 故障排查

### 常见问题

1. **无法登录**
   - 检查API密钥是否正确
   - 确认后端服务正常运行
   - 查看浏览器控制台错误信息

2. **频繁要求重新登录**
   - 检查Cookie设置
   - 确认"记住我"选项已勾选
   - 检查会话有效期配置

3. **登录后无法访问页面**
   - 清除浏览器Cookie
   - 检查认证中间件配置
   - 查看后端日志

### 调试模式

在 `web/js/auth-check.js` 中启用调试模式:

```javascript
const CONFIG = {
    debug: true,  // 启用调试日志
    // ...
};
```

## 📚 参考资源

- [借鉴代码: Cli-Proxy-API-Management-Center](../借鉴代码/Cli-Proxy-API-Management-Center-main/)
- [JWT.io](https://jwt.io/) - JWT token介绍
- [OWASP认证备忘单](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)

## 📝 更新日志

### 2024-01-13
- ✅ 实现强制登录功能
- ✅ 创建独立登录页面
- ✅ 添加"记住我"功能
- ✅ 优化Cookie安全设置
- ✅ 完善认证检查机制

---

**文档版本**: 1.0  
**最后更新**: 2024-01-13  
**维护者**: System Administrator
