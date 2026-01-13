# T-13: 首次登录强制认证功能实现计划

## 📋 需求概述

实现"第一次登录需要强制认证，第二次登录之后可以保存登录信息"的功能。

### 核心需求
1. **首次登录强制认证**：用户首次访问系统时，必须进行身份认证
2. **记住登录状态**：认证成功后，系统可以记住用户的登录状态（通过Cookie）
3. **会话持久化**：用户关闭浏览器后，再次访问时无需重新认证（在会话有效期内）

## 🎯 设计方案

### 1. 数据库设计

需要在 `sessions` 表中添加字段来跟踪首次登录：

```sql
-- 新增迁移文件
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS is_first_login BOOLEAN DEFAULT true;
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(255);
```

### 2. 认证流程设计

#### 首次登录流程
```
用户访问系统
    ↓
检查Cookie中的session_token
    ↓
Token不存在或无效 → 重定向到登录页面
    ↓
用户输入API Key
    ↓
验证API Key
    ↓
创建Session (is_first_login = true)
    ↓
设置HttpOnly Cookie
    ↓
重定向到首页
```

#### 后续访问流程
```
用户访问系统
    ↓
检查Cookie中的session_token
    ↓
验证Token有效性
    ↓
Token有效 → 直接访问（无需重新认证）
    ↓
更新LastActive时间
```

### 3. 前端实现

#### 登录页面 (`web/login.html`)
- 简洁的登录表单
- API Key输入框
- 记住我选项（默认勾选）
- 错误提示显示

#### 认证检查逻辑
- 在所有页面加载前检查认证状态
- 未认证时自动跳转到登录页
- 已认证时正常显示内容

### 4. 后端实现

#### API端点
- `POST /api/v1/auth/login` - 登录接口（已存在）
- `GET /api/v1/auth/session` - 获取会话信息（已存在）
- `POST /api/v1/auth/logout` - 登出接口（已存在）

#### 中间件增强
- 修改 `middleware/auth.go` 中的认证逻辑
- 支持可选认证模式（公开页面）
- 支持强制认证模式（管理页面）

## 📁 文件修改清单

### 需要修改的文件

1. **数据库迁移**
   - ✅ 新建 `migrations/003_add_first_login_fields.sql`
   - 添加 `is_first_login` 和 `device_fingerprint` 字段

2. **模型层**
   - ✏️ 修改 `models/session.go`
   - 添加 `IsFirstLogin` 和 `DeviceFingerprint` 字段

3. **服务层**
   - ✏️ 修改 `services/auth.go`
   - 在 `Login` 方法中设置 `is_first_login = true`
   - 在 `ValidateToken` 中检查首次登录状态

4. **API层**
   - ✏️ 修改 `api/auth.go`
   - 在登录响应中返回 `is_first_login` 标识
   - 添加首次登录提示信息

5. **中间件**
   - ✏️ 修改 `middleware/auth.go`
   - 添加可选认证中间件（已存在 `OptionalAdminAuth`）
   - 优化认证失败时的重定向逻辑

6. **前端**
   - ✅ 新建 `web/login.html` - 登录页面
   - ✅ 新建 `web/js/login.js` - 登录逻辑
   - ✏️ 修改 `web/js/api.js` - 添加认证检查
   - ✏️ 修改现有页面 - 添加认证状态检查

7. **配置**
   - ✏️ 修改 `config/config.go`
   - 添加会话持久化配置选项

8. **文档**
   - ✅ 新建 `docs/FIRST_LOGIN_AUTH.md` - 功能说明文档
   - ✏️ 更新 `docs/API.md` - 添加认证相关API文档

## 🔧 实现步骤

### Phase 1: 数据库和模型层
1. 创建数据库迁移文件
2. 更新 Session 模型
3. 测试数据库变更

### Phase 2: 服务层和API层
1. 修改 AuthService 的 Login 方法
2. 修改 AuthHandler 的登录响应
3. 添加首次登录标识

### Phase 3: 中间件和路由
1. 优化认证中间件
2. 配置路由认证策略
3. 测试认证流程

### Phase 4: 前端实现
1. 创建登录页面
2. 实现前端认证检查
3. 添加自动跳转逻辑
4. 测试用户体验

### Phase 5: 测试和文档
1. 编写单元测试
2. 编写集成测试
3. 更新用户文档
4. 性能测试

## 🎨 用户体验设计

### 登录页面UI
```
┌─────────────────────────────────┐
│                                 │
│        🎬 TMDB Crawler          │
│                                 │
│   ┌─────────────────────────┐   │
│   │  请输入管理员API密钥     │   │
│   │  ┌───────────────────┐  │   │
│   │  │                   │  │   │
│   │  └───────────────────┘  │   │
│   │                         │   │
│   │  ☑ 记住登录状态 (30天)  │   │
│   │                         │   │
│   │  [ 登录 ]               │   │
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

### 认证流程体验
1. **首次访问** → 自动跳转到登录页
2. **输入API Key** → 点击登录
3. **登录成功** → 跳转到首页，显示欢迎消息
4. **关闭浏览器** → 重新打开，自动登录（Cookie有效期内）
5. **Cookie过期** → 重新跳转到登录页

## 🔒 安全考虑

1. **Cookie安全**
   - 使用 `HttpOnly` 防止XSS攻击
   - 使用 `SameSite=Strict` 防止CSRF攻击
   - 生产环境使用 `Secure` 标志（HTTPS）

2. **会话管理**
   - 会话有效期：30天（可配置）
   - 支持主动登出
   - 支持多设备登录

3. **API Key保护**
   - 不在前端缓存API Key
   - 只在服务端验证
   - 支持环境变量配置

## 📊 配置选项

### 环境变量
```bash
# 认证配置
ADMIN_API_KEY=your-secret-key-here
ALLOW_REMOTE_ADMIN=false

# 会话配置
SESSION_DURATION=720h  # 30天
REMEMBER_ME_ENABLED=true
```

### 配置文件
```go
type AuthConfig struct {
    SecretKey      string
    AllowRemote    bool
    SessionDuration time.Duration  // 新增
    RememberMe     bool            // 新增
}
```

## ✅ 验收标准

1. ✅ 首次访问系统时，自动跳转到登录页面
2. ✅ 输入正确的API Key后，成功登录并跳转到首页
3. ✅ 登录后，关闭浏览器重新打开，自动保持登录状态
4. ✅ Cookie过期后，重新跳转到登录页面
5. ✅ 可以主动登出，清除登录状态
6. ✅ 支持多设备同时登录
7. ✅ 错误提示清晰友好
8. ✅ 移动端适配良好

## 📝 注意事项

1. **向后兼容**：现有的API Key认证方式继续支持
2. **渐进增强**：未配置API Key时，系统仍可运行（开发模式）
3. **性能优化**：认证检查不应影响页面加载速度
4. **可测试性**：提供测试模式，方便开发调试

## 🚀 后续优化

1. 支持多种认证方式（OAuth、LDAP等）
2. 添加双因素认证（2FA）
3. 实现设备管理（查看/移除登录设备）
4. 添加登录历史记录
5. 支持会话超时自动续期

## 📅 时间估算

- Phase 1: 2小时
- Phase 2: 3小时
- Phase 3: 2小时
- Phase 4: 4小时
- Phase 5: 2小时

**总计**: 约13小时

---

**创建时间**: 2026-01-12
**状态**: 待实施
**优先级**: 高
