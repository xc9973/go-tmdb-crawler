# 剧集智能纠错机制设计文档

**日期:** 2025-01-23
**状态:** 设计阶段
**作者:** AI + User

---

## 一、概述

这是一个自动检测和修复"卡住"剧集数据的系统。核心思想是通过分析每个剧集的历史更新频率，识别那些应该有新更新但 TMDB 数据没有同步的剧集，然后自动加入刷新队列进行修复。

### 核心价值

- **自动维护数据新鲜度**：减少手动干预，自动发现过期数据
- **基于真实播出规律**：误报率低，适应各种更新频率
- **批量处理避免 API 压力**：通过队列机制控制刷新速率

### 核心流程

```
每日凌晨定时任务启动
    ↓
遍历所有剧集，分析最近 10 集的播出间隔
    ↓
计算正常更新频率（间隔众数）
    ↓
检查最新一集是否超过 1.5 倍正常间隔
    ↓
将异常剧集加入刷新队列
    ↓
批量刷新并记录结果
```

---

## 二、检测算法设计

### 2.1 计算历史更新间隔

**输入：** 剧集 ID

**步骤：**
1. 查询剧集的所有 episodes，按 `air_date ASC` 排序
2. 如果 episode 数量 < 3，跳过（数据不足）
3. 取最近 10 集的播出日期
4. 计算相邻两集之间的天数间隔 [d1, d2, d3, ...]
5. 过滤异常间隔（> 60 天的视为季间空档期，不计入）
6. 计算众数（most common value）作为"正常更新间隔"

### 2.2 判断是否需要刷新

**步骤：**
1. 获取最新一集的 `air_date`
2. 计算距离今天的天数 `gap_days`
3. 如果 `gap_days > 正常更新间隔 * 1.5`：
   - → 标记为"可能过期"
   - → 加入刷新队列

### 2.3 检测示例

| 剧集类型 | 历史间隔 | 正常间隔 | 触发阈值 |
|---------|---------|---------|---------|
| 周更剧 | [7,7,8,7,7] | 7天 | >10.5天 |
| 日更剧 | [1,1,1,2,1] | 1天 | >1.5天 |
| 月更剧 | [30,30,31,30] | 30天 | >45天 |

---

## 三、边界条件处理

| 场景 | 策略 |
|------|------|
| 新剧集（<3集） | 跳过检测，等待数据积累 |
| 完结剧集（Ended） | 仍检测（TMDB 状态可能有误） |
| 有空档期 | 只看最近 10 集的间隔，忽略 >60 天的间隔 |
| 已取消/停播 | 检测但降低优先级 |

---

## 四、数据模型

### 4.1 数据库变更

**新增字段（shows 表）：**

```sql
ALTER TABLE shows ADD COLUMN refresh_threshold INT DEFAULT NULL;
ALTER TABLE shows ADD COLUMN stale_detected_at DATETIME DEFAULT NULL;
ALTER TABLE shows ADD COLUMN last_correction_result VARCHAR(50) DEFAULT NULL;
```

**字段说明：**
| 字段 | 类型 | 说明 |
|------|------|------|
| `refresh_threshold` | INT | 自动计算的刷新阈值（天数），NULL 表示自动计算 |
| `stale_detected_at` | DATETIME | 最近一次被检测为过期的时间 |
| `last_correction_result` | VARCHAR(50) | 最近一次纠错结果（pending/success/failed） |

### 4.2 CrawlTask 扩展

**新增任务类型：**
```go
const TaskTypeCorrection = "correction"
```

**任务状态流转：**
```
pending → processing → success/failed
```

---

## 五、API 设计

### 5.1 新增端点

| 方法 | 路径 | 功能 | 权限 |
|------|------|------|------|
| GET | `/api/v1/correction/status` | 获取纠错状态统计 | 公开 |
| POST | `/api/v1/correction/run-now` | 立即运行检测 | 管理员 |
| GET | `/api/v1/correction/stale` | 获取过期剧集列表 | 管理员 |
| POST | `/api/v1/correction/:id/override` | 手动设置阈值 | 管理员 |
| DELETE | `/api/v1/correction/:id/stale` | 清除过期标记 | 管理员 |

### 5.2 响应示例

**GET /api/v1/correction/status**
```json
{
  "code": 200,
  "data": {
    "total_shows": 150,
    "stale_count": 5,
    "pending_refresh": 3,
    "last_check_at": "2025-01-22T02:00:00Z",
    "next_check_at": "2025-01-23T02:00:00Z",
    "stale_shows": [
      {
        "id": 1,
        "name": "剧集A",
        "tmdb_id": 12345,
        "normal_interval": 7,
        "days_overdue": 12,
        "latest_episode_date": "2025-01-10T00:00:00Z"
      }
    ]
  }
}
```

---

## 六、调度与任务处理

### 6.1 定时任务配置

```go
// 每天凌晨 2:00 运行纠错检测
scheduler.AddJob("0 2 * * *", CorrectionDetectionJob)
```

### 6.2 检测任务流程

```
1. 开启事务
2. 遍历所有 shows（状态为 Returning Series 或 Ended）
3. 对每个 show 运行检测算法：
   a. 获取最近 10 集数据
   b. 计算更新规律
   c. 判断是否过期
4. 如果检测为过期：
   - 创建/更新 CrawlTask 记录（type = "correction"）
   - 设置状态为 "pending"
   - 记录 stale_detected_at
   - 根据逾期天数设置优先级
5. 提交事务
6. 返回统计信息（总数、过期数、已加入队列数）
```

### 6.3 刷新任务执行

```
1. TaskManager 定期扫描 type = "correction" 且 status = "pending" 的任务
2. 按优先级排序（ overdue 天数越多优先级越高）
3. 批量调用 CrawlerService.CrawlShow()（每批最多 10 个）
4. 更新 CrawlTask 状态和结果
5. 清除 stale_detected_at 标记
```

### 6.4 错误处理

| 错误类型 | 处理方式 |
|---------|---------|
| TMDB API 失败 | 标记为 failed，保留任务以便重试 |
| 数据库错误 | 记录日志，跳过该剧集，继续处理下一个 |
| 网络超时 | 重试 3 次，失败后标记为 failed |
| 限流 | 每次最多处理 10 个过期剧集，间隔 5 秒 |

---

## 七、前端界面设计

### 7.1 首页健康卡片

```html
<div class="health-card">
  <h3>📊 数据健康状态</h3>
  <div class="stats">
    <span>总剧集: <strong>150</strong></span>
    <span>过期: <strong class="warning">5</strong></span>
    <span>正常: <strong class="success">145</strong></span>
  </div>
  <div class="actions">
    <button onclick="location.href='/correction.html'">查看过期剧集</button>
    <button onclick="runCorrectionNow()">立即检测</button>
  </div>
  <div class="last-check">上次检测: 2小时前</div>
</div>
```

### 7.2 过期剧集列表页面

**路径：** `/correction.html`

**表格列：**
- 剧集名称
- 海报
- 正常更新间隔
- 最新一集日期
- 逾期天数
- 状态
- 操作（立即刷新、忽略、编辑阈值）

### 7.3 日志页面增强

在现有 `logs.html` 中添加"纠错任务"标签页，显示：
- 检测时间
- 发现过期数
- 刷新成功/失败数
- 执行时长

### 7.4 通知机制

- 检测完成后在顶部显示 toast 通知
- 示例：`"检测完成：发现 5 个过期剧集，已加入刷新队列"`
- 使用现有的轮询机制获取实时状态

### 7.5 视觉标识

- 剧集列表页面：过期剧集显示黄色警告图标
- 状态栏显示"上次检测: X小时前"

---

## 八、实现计划

### Phase 1: 核心检测逻辑
```
services/correction/
  - correction.go         # CorrectionService 主服务
  - detector.go           # 检测算法
  - pattern.go            # 更新规律计算
```

### Phase 2: 数据库与任务
```
migrations/
  - 006_add_correction_fields.sql
repositories/
  - correction_task.go    # 纠错任务仓储
services/
  - task_manager.go       # 扩展支持 correction 类型
```

### Phase 3: API 层
```
api/
  - correction.go         # CorrectionAPI 处理器
```

### Phase 4: 调度集成
```
services/
  - scheduler.go          # 扩展添加纠错任务
```

### Phase 5: 前端界面
```
web/
  - correction.html       # 过期剧集列表页（新建）
  - js/correction.js      # 纠错相关逻辑
  - css/correction.css    # 样式
  - index.html            # 添加健康卡片
  - js/index.js           # 添加健康卡片逻辑
```

---

## 九、测试策略

### 9.1 单元测试

**检测算法测试用例：**
```go
func TestCalculateUpdatePattern(t *testing.T) {
    tests := []struct {
        name           string
        intervals      []int
        expectedMode   int
        expectedThreshold int
    }{
        {"周更", []int{7,7,8,7,7}, 7, 10},
        {"日更", []int{1,1,1,2,1}, 1, 1},
        {"月更", []int{30,30,31,30}, 30, 45},
        {"有空档期", []int{7,7,90,7,7}, 7, 10}, // 90天应被过滤
    }
    // ...
}
```

### 9.2 集成测试

- 完整检测+刷新流程
- TMDB API 失败重试
- 限流控制

### 9.3 Mock 测试

- Mock TMDB API 验证限流和重试
- 时间旅行测试（模拟不同日期）

---

## 十、性能考虑

| 指标 | 目标 | 实现方式 |
|------|------|---------|
| 检测耗时 | < 30秒 | 只查询最近 10 集，使用索引 |
| API 调用 | < 100次/天 | 批量处理，去重 |
| 内存占用 | < 50MB | 流式处理，不缓存全部数据 |

---

## 十一、后续优化

- [ ] 支持自定义检测规则（按国家/类型）
- [ ] 机器学习预测播出间隔
- [ ] 与第三方数据源交叉验证
- [ ] 自动调整 TMDB API 缓存策略
