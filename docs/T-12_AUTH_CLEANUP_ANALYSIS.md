# T-12 Auth 失败记录清理 - 分析报告

## 任务概述

**任务编号**: T-12  
**优先级**: 低优先级 (Low)  
**问题来源**: 代码审查报告2.0.md  
**影响范围**: `middleware/auth.go`

## 问题描述

### 当前问题
根据代码审查报告（第69-71行）：

> **Auth 中间件失败记录无淘汰**
> - 位置: `go-tmdb-crawler/middleware/auth.go`
> - 现象: `failedAttempts` map 永不清理，长期运行可能增长。
> - 建议: 定期清理过期记录或设置最大容量。

### 问题分析

#### 1. 数据结构
```go
// middleware/auth.go:39
failedAttempts map[string]*attemptInfo

// middleware/auth.go:48-52
type attemptInfo struct {
    count        int
    blockedUntil time.Time
    lastActivity time.Time
}
```

#### 2. 当前行为

**记录创建**（第223-243行）：
- 当认证失败时，在 `recordFailure()` 中创建新记录
- 记录包含：失败次数、封禁截止时间、最后活动时间
- 失败次数达到 5 次后，封禁 30 分钟

**记录清除**（第245-254行）：
- 仅在认证成功时调用 `clearFailure()`
- 清除操作：重置 count 为 0，清空 blockedUntil
- **但不会从 map 中删除记录**

**记录查询**（第78-97行）：
- 每次请求都会检查 `failedAttempts[clientIP]`
- 即使记录已过期，仍保留在 map 中

#### 3. 问题影响

**内存泄漏风险**：
- 每个唯一 IP 地址都会在 map 中创建一条记录
- 记录永不删除，长期运行会累积大量记录
- 假设每天有 1000 个不同 IP 访问，一年将累积 365,000 条记录

**性能影响**：
- Map 查询时间复杂度为 O(1)，但大量记录会影响缓存性能
- 每次 GC 都需要扫描这些对象
- 内存占用持续增长

**实际场景**：
- 正常用户：认证成功后记录被清空（count=0），但记录仍存在
- 攻击者：被封禁后，30 分钟后解封，但记录永久保留
- 误操作用户：多次失败后不再访问，记录永久保留

## 解决方案

### 方案一：定期清理过期记录（推荐）

#### 实现方式
添加后台 goroutine，定期清理过期记录。

```go
// 在 AdminAuth 结构体中添加
type AdminAuth struct {
    config         *AuthConfig
    envSecret      string
    mu             sync.Mutex
    failedAttempts map[string]*attemptInfo
    authService    AuthService
    cleanupStop    chan struct{}  // 新增：停止清理通道
}

// 启动清理任务
func (a *AdminAuth) startCleanupRoutine() {
    a.cleanupStop = make(chan struct{})
    
    go func() {
        ticker := time.NewTicker(5 * time.Minute)  // 每5分钟清理一次
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                a.cleanupExpiredAttempts()
            case <-a.cleanupStop:
                return
            }
        }
    }()
}

// 清理过期记录
func (a *AdminAuth) cleanupExpiredAttempts() {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    now := time.Now()
    const retentionPeriod = 24 * time.Hour  // 保留24小时
    
    for ip, ai := range a.failedAttempts {
        // 删除条件：
        // 1. 最后活动时间超过保留期
        // 2. 且当前未被封禁
        if now.Sub(ai.lastActivity) > retentionPeriod && 
           (ai.blockedUntil.IsZero() || now.After(ai.blockedUntil)) {
            delete(a.failedAttempts, ip)
        }
    }
}

// 停止清理任务
func (a *AdminAuth) Stop() {
    if a.cleanupStop != nil {
        close(a.cleanupStop)
    }
}
```

#### 优点
- 自动清理，无需手动干预
- 可配置清理频率和保留期
- 不影响正常功能

#### 缺点
- 增加一个后台 goroutine
- 需要在应用关闭时正确停止

### 方案二：LRU 缓存（推荐）

#### 实现方式
使用 LRU（Least Recently Used）缓存限制 map 大小。

```go
import (
    "github.com/hashicorp/golang-lru/v2"
)

type AdminAuth struct {
    config         *AuthConfig
    envSecret      string
    mu             sync.Mutex
    failedAttempts *lru.Cache[string, *attemptInfo]  // 改为 LRU 缓存
    authService    AuthService
}

// 初始化
func NewAdminAuth(secretKey string, allowRemote bool) *AdminAuth {
    envSecret := strings.TrimSpace(os.Getenv("ADMIN_API_KEY"))
    
    // 创建 LRU 缓存，最多保留 10000 条记录
    cache, _ := lru.New[string, *attemptInfo](10000)
    
    return &AdminAuth{
        config: &AuthConfig{
            SecretKey:   secretKey,
            AllowRemote: allowRemote,
        },
        envSecret:      envSecret,
        failedAttempts: cache,
    }
}

// 修改 recordFailure
func (a *AdminAuth) recordFailure(clientIP string) {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    const maxFailures = 5
    const banDuration = 30 * time.Minute
    
    ai, exists := a.failedAttempts.Get(clientIP)
    if !exists {
        ai = &attemptInfo{}
        a.failedAttempts.Add(clientIP, ai)
    }
    
    ai.count++
    ai.lastActivity = time.Now()
    
    if ai.count >= maxFailures {
        ai.blockedUntil = time.Now().Add(banDuration)
        ai.count = 0
    }
}

// 修改 clearFailure
func (a *AdminAuth) clearFailure(clientIP string) {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    if ai, ok := a.failedAttempts.Get(clientIP); ok {
        ai.count = 0
        ai.blockedUntil = time.Time{}
    }
}
```

#### 优点
- 自动限制内存使用
- 无需后台 goroutine
- 性能优秀（O(1) 操作）
- 自动淘汰最久未使用的记录

#### 缺点
- 需要引入外部依赖（golang-lru）
- 可能淘汰仍需要的记录（但概率很低）

### 方案三：惰性清理（简单）

#### 实现方式
在每次操作时顺便清理过期记录。

```go
// 在 recordFailure 和查询时顺便清理
func (a *AdminAuth) recordFailure(clientIP string) {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    const maxFailures = 5
    const banDuration = 30 * time.Minute
    const retentionPeriod = 24 * time.Hour
    
    // 每次记录失败时，顺便清理一些过期记录
    // 限制每次最多清理 10 条，避免影响性能
    a.cleanupExpiredAttemptsLocked(10)
    
    ai := a.failedAttempts[clientIP]
    if ai == nil {
        ai = &attemptInfo{}
        a.failedAttempts[clientIP] = ai
    }
    
    ai.count++
    ai.lastActivity = time.Now()
    
    if ai.count >= maxFailures {
        ai.blockedUntil = time.Now().Add(banDuration)
        ai.count = 0
    }
}

// 清理过期记录（内部方法，已持有锁）
func (a *AdminAuth) cleanupExpiredAttemptsLocked(maxClean int) {
    now := time.Now()
    const retentionPeriod = 24 * time.Hour
    
    cleaned := 0
    for ip, ai := range a.failedAttempts {
        if cleaned >= maxClean {
            break
        }
        
        if now.Sub(ai.lastActivity) > retentionPeriod && 
           (ai.blockedUntil.IsZero() || now.After(ai.blockedUntil)) {
            delete(a.failedAttempts, ip)
            cleaned++
        }
    }
}
```

#### 优点
- 无需额外 goroutine
- 无需外部依赖
- 实现简单

#### 缺点
- 清理不及时，依赖请求频率
- 可能影响请求性能（但影响很小）

### 方案四：混合方案（最优）

结合方案一和方案二，使用 LRU 缓存 + 定期清理。

```go
type AdminAuth struct {
    config         *AuthConfig
    envSecret      string
    mu             sync.Mutex
    failedAttempts *lru.Cache[string, *attemptInfo]
    authService    AuthService
    cleanupStop    chan struct{}
}

func NewAdminAuth(secretKey string, allowRemote bool) *AdminAuth {
    envSecret := strings.TrimSpace(os.Getenv("ADMIN_API_KEY"))
    
    // LRU 缓存限制最大记录数
    cache, _ := lru.New[string, *attemptInfo](10000)
    
    auth := &AdminAuth{
        config: &AuthConfig{
            SecretKey:   secretKey,
            AllowRemote: allowRemote,
        },
        envSecret:      envSecret,
        failedAttempts: cache,
        cleanupStop:    make(chan struct{}),
    }
    
    // 启动定期清理
    auth.startCleanupRoutine()
    
    return auth
}

func (a *AdminAuth) startCleanupRoutine() {
    go func() {
        ticker := time.NewTicker(10 * time.Minute)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                a.cleanupExpiredAttempts()
            case <-a.cleanupStop:
                return
            }
        }
    }()
}

func (a *AdminAuth) cleanupExpiredAttempts() {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    now := time.Now()
    const retentionPeriod = 24 * time.Hour
    
    // LRU 缓存需要遍历所有键
    keys := a.failedAttempts.Keys()
    for _, ip := range keys {
        if ai, ok := a.failedAttempts.Peek(ip); ok {
            if now.Sub(ai.lastActivity) > retentionPeriod && 
               (ai.blockedUntil.IsZero() || now.After(ai.blockedUntil)) {
                a.failedAttempts.Remove(ip)
            }
        }
    }
}

func (a *AdminAuth) Stop() {
    if a.cleanupStop != nil {
        close(a.cleanupStop)
    }
}
```

## 推荐方案

### 短期方案（快速实施）
**方案三：惰性清理**
- 无需外部依赖
- 实现简单
- 对现有代码改动最小
- 性能影响可忽略

### 长期方案（推荐）
**方案四：混合方案**
- 使用 LRU 缓存限制最大记录数
- 定期清理过期记录
- 最佳性能和内存控制

## 实现建议

### 1. 添加配置选项

```go
// 在 AuthConfig 中添加
type AuthConfig struct {
    SecretKey     string
    AllowRemote   bool
    LocalPassword string
    
    // 新增配置
    MaxFailedAttempts int           // 最大失败记录数（LRU 缓存大小）
    CleanupInterval  time.Duration  // 清理间隔
    RetentionPeriod  time.Duration  // 记录保留期
}

// 默认值
const (
    DefaultMaxFailedAttempts = 10000
    DefaultCleanupInterval  = 10 * time.Minute
    DefaultRetentionPeriod  = 24 * time.Hour
)
```

### 2. 添加监控指标

```go
// 获取失败记录统计
func (a *AdminAuth) GetFailedAttemptsStats() map[string]interface{} {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    activeCount := 0
    blockedCount := 0
    expiredCount := 0
    now := time.Now()
    
    for _, ai := range a.failedAttempts {
        if !ai.blockedUntil.IsZero() && now.Before(ai.blockedUntil) {
            blockedCount++
        } else if ai.count > 0 {
            activeCount++
        } else {
            expiredCount++
        }
    }
    
    return map[string]interface{}{
        "total_records":  len(a.failedAttempts),
        "active_count":   activeCount,
        "blocked_count":  blockedCount,
        "expired_count":  expiredCount,
    }
}
```

### 3. 添加测试

```go
func TestFailedAttemptsCleanup(t *testing.T) {
    auth := NewAdminAuth("test-key", true)
    defer auth.Stop()
    
    // 模拟多次失败
    for i := 0; i < 10; i++ {
        auth.recordFailure("192.168.1.1")
    }
    
    // 检查记录存在
    stats := auth.GetFailedAttemptsStats()
    assert.Equal(t, 1, stats["total_records"])
    
    // 等待清理
    time.Sleep(11 * time.Minute)
    
    // 检查记录已清理
    stats = auth.GetFailedAttemptsStats()
    assert.Equal(t, 0, stats["total_records"])
}
```

## 验收标准

根据任务清单，T-12 的验收标准是：
- ✅ 长期运行内存稳定

### 具体验证方法

1. **内存测试**
   ```bash
   # 运行服务并模拟大量 IP 访问
   # 监控内存使用情况
   go tool pprof -http=:8081 <heap_profile>
   ```

2. **记录数测试**
   ```go
   // 检查记录数是否稳定
   stats := auth.GetFailedAttemptsStats()
   fmt.Printf("Total records: %d\n", stats["total_records"])
   ```

3. **压力测试**
   ```bash
   # 模拟 10000 个不同 IP 访问
   # 验证内存不会持续增长
   ```

## 风险评估

### 当前风险
- **风险等级**: 中
- **影响**: 长期运行可能导致内存泄漏
- **概率**: 高（必然发生）

### 实施风险
- **风险等级**: 低
- **影响**: 可能引入新的 bug
- **缓解措施**: 充分测试，逐步推出

### 不实施风险
- **风险等级**: 中
- **影响**: 内存持续增长，最终可能导致 OOM
- **时间**: 数周到数月（取决于访问量）

## 相关文件

- `middleware/auth.go` - 认证中间件实现
- `代码审查报告2.0.md` - 问题来源（第69-71行）
- `代码优化任务.md` - 任务定义（第129-137行）

## 总结

### 问题严重性
- **当前状态**: `failedAttempts` map 永不清理，存在内存泄漏风险
- **影响范围**: 长期运行的服务
- **紧急程度**: 低（但需要解决）

### 推荐行动
1. **短期**: 实施惰性清理（方案三），快速解决问题
2. **长期**: 升级到混合方案（方案四），提供更好的性能和控制

### 实施优先级
- **高优先级**: 如果服务需要长期运行（数月以上）
- **中优先级**: 如果服务定期重启（每周/每月）
- **低优先级**: 如果服务访问量很小（每天 < 100 个不同 IP）

### 预期效果
- 内存使用稳定，不会持续增长
- 性能影响可忽略（< 1%）
- 代码复杂度略微增加
