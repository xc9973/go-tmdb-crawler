# 网页性能优化方案 A

## 概述

针对 go-tmdb-crawler 项目在公网部署时页面加载慢的问题，采用快速优化方案，通过添加 HTTP 缓存、合并 JS 文件、优化认证检查来提升用户体验。

**优化目标：**
- 静态资源（CSS/JS）设置缓存，减少重复下载
- 合并常用 JS 文件，减少 HTTP 请求
- 优化认证检查，避免阻塞页面渲染

## 问题诊断

**当前问题：**
- 使用 `router.Static()` 服务静态文件，无缓存头
- 6-7 个 JS 文件独立加载
- Bootstrap 从 CDN 加载，国内访问慢
- 每次 page navigation 都完整重新加载

**影响：**
- 页面跳转慢（需要重新下载 360KB+ 资源）
- 公网环境 CDN 访问不稳定
- 多个 HTTP 请求增加延迟

## 优化方案

### 1. 静态文件缓存

**文件：** `api/setup.go`

添加自定义静态文件处理器，为 CSS/JS 添加缓存头：

```go
func staticCacheMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if strings.HasSuffix(c.Request.URL.Path, ".css") ||
           strings.HasSuffix(c.Request.URL.Path, ".js") {
            c.Header("Cache-Control", "public, max-age=86400") // 1天
            c.Header("Vary", "Accept-Encoding")
        }
        c.Next()
    }
}

func setupCachedStaticFiles(router *gin.Engine, cfg *config.Config) {
    router.Use(staticCacheMiddleware())
    // ... 设置静态文件路由
}
```

**预期效果：** 回访时静态资源直接从浏览器缓存读取，几乎瞬时加载。

### 2. JS 文件合并

**当前结构：**
```
js/auth-check.js
js/api.js
js/feedback.js
js/auth-ui.js
js/modal.js
js/shows.js
js/today.js
js/logs.js
js/backup.js
```

**优化后结构：**
```
js/common.js      ← 合并 auth-check + api + feedback + auth-ui
js/modal.js       ← 保持独立
js/shows.js       ← 首页专用
js/today.js       ← 今日页专用
js/logs.js        ← 日志页专用
js/backup.js      ← 备份页专用
```

**auth-check.js 异步化：**
```javascript
// 同步阻塞 → 异步非阻塞
document.addEventListener('DOMContentLoaded', async () => {
    const isAuth = await checkAuth();
    updateAuthUI(isAuth);
});
```

### 3. API 响应优化

**文件：** `api/setup.go`

```go
import "github.com/gin-contrib/gzip"

router.Use(gzip.Gzip(gzip.DefaultCompression))
```

### 4. 资源预加载

**文件：** `web/*.html` 的 `<head>` 部分

```html
<link rel="dns-prefetch" href="//cdn.jsdelivr.net">
<link rel="preload" href="css/custom.css" as="style">
<link rel="preload" href="js/common.js" as="script">
```

### 5. 版本号管理

为 JS 文件添加版本号，缓存更新时强制刷新：

```html
<script src="js/common.js?v=2.1"></script>
```

## 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `api/setup.go` | 修改 | 添加 gzip、缓存中间件 |
| `web/js/common.js` | 新建 | 合并 4 个公共 JS 文件 |
| `web/js/auth-check.js` | 修改 | 异步化改造 |
| `web/index.html` | 修改 | 更新 JS 引用、添加预加载 |
| `web/today.html` | 修改 | 更新 JS 引用、添加预加载 |
| `web/logs.html` | 修改 | 更新 JS 引用、添加预加载 |
| `web/backup.html` | 修改 | 更新 JS 引用、添加预加载 |
| `web/show_detail.html` | 修改 | 更新 JS 引用、添加预加载 |

## 错误处理

1. **缓存失效** - 通过版本号 `?v=x.x` 强制刷新
2. **认证降级** - API 超时时不阻塞页面渲染
3. **CDN 降级** - CDN 失败时提供本地备用方案

## 性能验证

1. **浏览器 DevTools Network** - 查看资源加载时间
2. **Lighthouse** - 目标 Performance >80
3. **实际测试** - 首页 → 今日页 → 备份页，观察二次加载速度

## 预期效果

- **首次访问**：略慢（需下载完整资源）
- **后续跳转**：缓存后几乎秒开
- **API 响应**：gzip 压缩后减少 70% 传输量
- **总体提升**：50-70% 性能改善

## 工作量评估

| 任务 | 预计时间 |
|------|----------|
| 添加缓存中间件 | 30分钟 |
| 合并 JS 文件 | 30分钟 |
| 异步化认证 | 20分钟 |
| 更新 HTML 引用 | 20分钟 |
| 测试验证 | 20分钟 |
| **总计** | **约 2小时** |
