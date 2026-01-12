# 调度器并发控制指南

## 概述

调度器实现了完善的并发控制机制，防止同类任务重复并发执行，避免资源堆积和任务冲突。

## 并发控制机制

### 1. 互斥锁（Mutex）

调度器为不同类型的任务使用独立的互斥锁：

- **crawlJobMutex**: 控制爬取任务（dailyCrawlJob, weeklyCrawlJob）
- **publishJobMutex**: 控制发布任务（dailyPublishJob, weeklyPublishJob）

### 2. TryLock 机制

使用 `TryLock()` 而非 `Lock()`，实现非阻塞的并发控制：

```go
func (s *Scheduler) dailyCrawlJob() {
    // 尝试获取锁，如果失败则跳过本次执行
    if !s.crawlJobMutex.TryLock() {
        s.logger.Warn("Daily crawl job already running, skipping")
        return
    }
    defer s.crawlJobMutex.Unlock()
    
    // 执行任务逻辑...
}
```

**优势**：
- 非阻塞：不会等待锁释放
- 快速失败：立即返回，避免任务堆积
- 日志记录：记录跳过事件，便于监控

### 3. 同步执行

所有调度任务改为同步执行，移除了之前的 `go func()` 异步执行：

```go
// 修复前：异步执行
go func() {
    if err := s.crawler.RefreshAll(); err != nil {
        s.logger.Errorf("Daily crawl failed: %v", err)
    }
}()

// 修复后：同步执行
if err := s.crawler.RefreshAll(); err != nil {
    s.logger.Errorf("Daily crawl failed: %v", err)
}
```

**优势**：
- 可追踪：任务执行状态清晰
- 可控：避免 goroutine 泄漏
- 简单：不需要额外的同步机制

## 超时控制

### 默认超时设置

调度器为不同类型的任务设置了默认超时：

| 任务类型 | 默认超时 | 说明 |
|---------|---------|------|
| 爬取任务 | 30 分钟 | RefreshAll 操作可能较慢 |
| 发布任务 | 10 分钟 | 发布到 Telegraph 相对较快 |

### 自定义超时

可以通过 `SetTimeouts` 方法自定义超时时间：

```go
import "time"

scheduler := services.NewScheduler(crawler, publisher, logger)

// 设置自定义超时
scheduler.SetTimeouts(
    15 * time.Minute,  // 爬取任务超时
    5 * time.Minute,   // 发布任务超时
)
```

### 超时执行

使用 `runJobWithTimeout` 辅助函数执行带超时的任务：

```go
func (s *Scheduler) runJobWithTimeout(
    jobName string,
    timeout time.Duration,
    job func() error,
) error {
    // 创建结果通道
    done := make(chan error, 1)
    
    // 在 goroutine 中执行任务
    go func() {
        done <- job()
    }()
    
    // 等待任务完成或超时
    select {
    case err := <-done:
        // 任务完成
        return err
    case <-time.After(timeout):
        // 任务超时
        return fmt.Errorf("%s timed out after %v", jobName, timeout)
    }
}
```

## 任务状态追踪

### 状态字段

调度器维护以下状态字段：

```go
type Scheduler struct {
    // ... 其他字段
    
    // 并发控制
    crawlJobRunning    bool
    publishJobRunning  bool
    crawlJobMutex      sync.Mutex
    publishJobMutex    sync.Mutex
    
    // 超时设置
    crawlTimeout    time.Duration
    publishTimeout  time.Duration
}
```

### 状态查询

通过 `GetStatus()` 方法获取调度器状态：

```go
status := scheduler.GetStatus()
// 返回：
// {
//     "running": true,
//     "last_crawl_time": "2024-01-15T08:00:00Z",
//     "last_publish_time": "2024-01-15T20:30:00Z",
//     "crawl_job_running": false,
//     "publish_job_running": false
// }
```

### 超时设置查询

通过 `GetTimeouts()` 方法获取当前超时设置：

```go
timeouts := scheduler.GetTimeouts()
// 返回：
// {
//     "crawl_timeout": "30m0s",
//     "publish_timeout": "10m0s"
// }
```

## 并发场景

### 场景 1：同类任务并发触发

**问题**：如果调度器配置了多个相同类型的任务，可能同时触发。

**解决方案**：互斥锁确保同一时间只有一个任务执行。

```cron
# 配置示例：每天 8:00 和 12:00 执行爬取
0 0 8,12 * * *

# 如果 8:00 的任务还在运行，12:00 的任务会被跳过
# 日志输出： "Daily crawl job already running, skipping"
```

### 场景 2：任务执行时间过长

**问题**：任务执行时间超过调度间隔。

**解决方案**：
1. 互斥锁防止新任务启动
2. 超时机制强制终止长时间运行的任务
3. 日志记录超时事件

```go
// 任务超时示例
// 日志输出： "Daily crawl timed out after 30m0s (limit: 30m0s)"
```

### 场景 3：手动触发与调度任务冲突

**问题**：手动触发任务时，调度任务可能正在执行。

**解决方案**：手动触发任务也使用相同的互斥锁。

```go
// 手动触发爬取
func (s *Scheduler) RunCrawlNow() error {
    // 使用相同的互斥锁
    if !s.crawlJobMutex.TryLock() {
        return fmt.Errorf("crawl job already running")
    }
    defer s.crawlJobMutex.Unlock()
    
    // 执行爬取...
}
```

## 最佳实践

### 1. 合理设置超时时间

根据实际任务执行时间设置超时：

```go
// 爬取任务：通常需要 5-15 分钟
scheduler.SetTimeouts(20 * time.Minute, 10 * time.Minute)

// 发布任务：通常需要 1-5 分钟
scheduler.SetTimeouts(30 * time.Minute, 5 * time.Minute)
```

### 2. 监控任务执行

定期检查调度器状态：

```go
status := scheduler.GetStatus()
if status["crawl_job_running"] == true {
    log.Warn("Crawl job is currently running")
}
```

### 3. 处理超时事件

超时后需要：

1. 检查任务是否真的失败
2. 清理可能残留的资源
3. 记录详细的错误信息
4. 考虑是否需要重试

### 4. 避免频繁调度

设置合理的调度间隔，避免任务堆积：

```cron
# ✅ 推荐：合理的调度间隔
0 0 8,12,20 * * *  # 每天 3 次

# ⚠️ 不推荐：过于频繁
*/5 * * * * *      # 每 5 秒
```

### 5. 使用日志监控

关注以下日志：

- `"already running, skipping"` - 任务被跳过
- `"timed out after"` - 任务超时
- `"completed in"` - 任务正常完成

## 故障排查

### 问题：任务总是被跳过

**可能原因**：
1. 任务执行时间过长
2. 调度间隔过短
3. 任务卡死或死锁

**解决方案**：
1. 检查任务执行日志
2. 增加超时时间
3. 优化任务性能
4. 检查是否有死锁

### 问题：任务频繁超时

**可能原因**：
1. 超时时间设置过短
2. 网络延迟
3. 外部 API 响应慢

**解决方案**：
1. 增加超时时间
2. 检查网络连接
3. 优化 API 调用
4. 添加重试机制

### 问题：goroutine 泄漏

**可能原因**：
1. 使用了 `go func()` 但没有正确管理
2. 通道未关闭
3. 无限循环

**解决方案**：
1. 使用同步执行而非异步
2. 使用 context 取消机制
3. 定期监控 goroutine 数量

## 性能考虑

### 内存使用

- **互斥锁**：每个锁占用约 24 字节
- **状态字段**：bool 字段占用 1 字节
- **总开销**：可忽略不计

### CPU 使用

- **TryLock**：非阻塞操作，CPU 开销极小
- **超时检查**：使用 select 和 time.After，CPU 开销小

### 并发性能

- **同步执行**：避免 goroutine 开销
- **互斥锁**：确保任务串行执行
- **吞吐量**：受任务执行时间限制

## 相关文档

- [Cron 表达式配置指南](CRON.md)
- [时区配置指南](TIMEZONE.md)
- [robfig/cron 文档](https://github.com/robfig/cron)

## 注意事项

1. **互斥锁范围**：不同类型的任务（爬取和发布）使用不同的锁，可以并行执行
2. **超时设置**：超时只取消等待，不会终止正在执行的任务
3. **日志记录**：所有并发控制事件都会记录到日志
4. **状态查询**：状态查询使用读锁，不会阻塞任务执行
5. **手动触发**：手动触发的任务也受并发控制约束
