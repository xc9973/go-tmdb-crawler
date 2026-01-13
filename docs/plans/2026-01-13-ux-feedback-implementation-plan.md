# 前端状态反馈体验增强实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 提升 Web 端状态反馈体验，包括批量操作进度条、友好错误提示、统一确认框和 API 重试机制。

**Architecture:** 新增 `web/js/feedback.js` 模块，集成四大核心反馈组件。通过在 `api.js` 和 `shows.js` 中引入和调用这些组件，实现渐进式的 UI/UX 增强。

**Tech Stack:** JavaScript (ES6+), Bootstrap 5, Bootstrap Icons.

---

### Task 1: 创建基础反馈模块 (feedback.js)

**Files:**
- Create: `web/js/feedback.js`

**Step 1: 编写基础结构与 ProgressBar 类**
实现一个 `ProgressBar` 类，用于管理页面顶部的进度条展示。

```javascript
class ProgressBar {
    constructor() {
        this.containerId = 'feedbackProgressContainer';
        this.progressBarId = 'feedbackProgressBar';
    }
    start(total, message = '正在处理...') { /* 实现代码 */ }
    update(current, success = true) { /* 实现代码 */ }
    complete(summary) { /* 实现代码 */ }
}
```

**Step 2: 实现 ErrorHandler 类**
实现错误码映射和友好消息转换。

**Step 3: 实现 ConfirmDialog 类**
封装 Bootstrap Modal 逻辑，返回 Promise。

**Step 4: 实现 RetryHandler 类**
实现简单的重试逻辑。

**Step 5: 导出全局 feedback 对象**

```javascript
window.feedback = {
    progress: new ProgressBar(),
    error: new ErrorHandler(),
    confirm: new ConfirmDialog(),
    retry: new RetryHandler()
};
```

**Step 6: 提交**
```bash
git add web/js/feedback.js
git commit -m "feat: add basic feedback module with progress, error, confirm and retry components"
```

---

### Task 2: 更新 HTML 结构与资源引入

**Files:**
- Modify: `web/index.html`
- Modify: `web/today.html`
- Modify: `web/logs.html`
- Modify: `web/show_detail.html`

**Step 1: 添加进度条和确认框容器**
在 `toastContainer` 附近添加必要的 HTML 片段。

**Step 2: 引入 feedback.js**
在 `js/api.js` 之后引入新脚本。

**Step 3: 提交**
```bash
git add web/*.html
git commit -m "feat: update HTML files to include feedback.js and UI containers"
```

---

### Task 3: 集成 API 重试与友好错误处理

**Files:**
- Modify: `web/js/api.js`

**Step 1: 在 request 方法中集成重试逻辑**
修改 `request` 方法，在 catch 块中调用 `feedback.retry.execute()`。

**Step 2: 使用 ErrorHandler 包装错误消息**
在抛出 Error 之前，使用 `feedback.error.getFriendlyMessage(data.message)`。

**Step 3: 提交**
```bash
git add web/js/api.js
git commit -m "feat: integrate retry logic and friendly error handling into API client"
```

---

### Task 4: 改进剧集列表页面的交互

**Files:**
- Modify: `web/js/shows.js`

**Step 1: 替换所有原生 confirm**
使用 `await feedback.confirm.show(...)` 替换 `window.confirm(...)`。

**Step 2: 在批量操作中接入进度条**
修改 `batchRefresh()` 和 `batchDelete()`，在循环中调用 `feedback.progress.update()`。

**Step 3: 提交**
```bash
git add web/js/shows.js
git commit -m "feat: enhance shows page with progress bars and unified confirm dialogs"
```

---

### Task 5: 验证与清理

**Step 1: 手动验证批量刷新进度**
**Step 2: 手动验证错误提示翻译**
**Step 3: 手动验证重试触发**
**Step 4: 清理冗余代码**
**Step 5: 最终提交**
