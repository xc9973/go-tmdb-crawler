# 集成测试文档

## 测试概述

本文档描述了TMDB剧集爬取系统的集成测试场景和步骤。

---

## 测试环境准备

### 1. 环境变量配置

```bash
# .env
TMDB_API_KEY=your_tmdb_api_key
TELEGRAPH_ACCESS_TOKEN=your_telegraph_token
DATABASE_URL=postgres://user:password@localhost/tmdb_test
SERVER_PORT=8080
```

### 2. 数据库初始化

```bash
# 创建测试数据库
createdb tmdb_test

# 运行迁移
psql tmdb_test < migrations/001_init_schema.sql
```

### 3. 启动服务

```bash
# 编译
go build -o tmdb-crawler main.go

# 启动服务
./tmdb-crawler server
```

---

## 测试场景

### 场景1: 完整爬取流程

**目标**: 测试从TMDB爬取数据到数据库的完整流程

**步骤**:

1. **通过CLI爬取单个剧集**
```bash
./tmdb-crawler crawl 1668
```

**验证**:
- [ ] 控制台显示爬取成功消息
- [ ] 数据库shows表中有新记录
- [ ] episodes表中有剧集数据

2. **验证数据**
```sql
SELECT id, name, tmdb_id FROM shows WHERE tmdb_id = 1668;
SELECT COUNT(*) FROM episodes WHERE show_id = (SELECT id FROM shows WHERE tmdb_id = 1668);
```

**预期结果**:
- 剧集名称正确
- 至少有1个季度的数据
- 每个季度有正确的集数

---

### 场景2: Web界面操作流程

**目标**: 测试Web界面的完整操作流程

**步骤**:

1. **访问剧集列表**
```
URL: http://localhost:8080/
```

**验证**:
- [ ] 页面正常加载
- [ ] 显示剧集列表
- [ ] 搜索功能正常
- [ ] 分页功能正常

2. **添加新剧集**
```
1. 点击"添加剧集"按钮
2. 输入TMDB ID: 1396 (Breaking Bad)
3. 点击"搜索"按钮
4. 验证自动填充信息
5. 点击"保存"按钮
```

**验证**:
- [ ] TMDB查询成功
- [ ] 表单自动填充
- [ ] 保存成功
- [ ] 列表刷新显示新剧集

3. **查看剧集详情**
```
1. 点击剧集名称
2. 验证详情页面
```

**验证**:
- [ ] 详情页显示完整信息
- [ ] 海报正常显示
- [ ] 剧集列表正确显示

4. **刷新剧集**
```
1. 在详情页点击"刷新"按钮
2. 验证刷新成功
```

**验证**:
- [ ] 显示加载动画
- [ ] 刷新成功提示
- [ ] 数据已更新

---

### 场景3: Telegraph发布流程

**目标**: 测试从数据爬取到Telegraph发布的完整流程

**步骤**:

1. **生成今日更新**
```bash
./tmdb-crawler publish today
```

**验证**:
- [ ] 命令执行成功
- [ ] 生成Markdown内容
- [ ] 返回Telegraph URL

2. **通过Web界面发布**
```
1. 访问 http://localhost:8080/today.html
2. 验证今日更新列表
3. 点击"发布到Telegraph"按钮
4. 验证发布成功
```

**验证**:
- [ ] 显示今日更新的剧集
- [ ] 点击发布按钮后显示成功提示
- [ ] 显示Telegraph链接
- [ ] 点击链接可以访问文章

3. **验证Telegraph文章**
```
访问返回的Telegraph URL
```

**验证**:
- [ ] 文章标题正确
- [ ] 剧集信息完整
- [ ] 海报图片正常显示
- [ ] 格式正确

---

### 场景4: API集成测试

**目标**: 测试REST API的完整功能

**测试用例**:

1. **获取剧集列表**
```bash
curl http://localhost:8080/api/v1/shows?page=1&page_size=10
```

**预期响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [...],
    "total": 10,
    "page": 1,
    "page_size": 10
  }
}
```

2. **获取单个剧集**
```bash
curl http://localhost:8080/api/v1/shows/1
```

**预期响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "Show Name",
    "tmdb_id": 12345,
    ...
  }
}
```

3. **添加剧集**
```bash
curl -X POST http://localhost:8080/api/v1/shows \
  -H "Content-Type: application/json" \
  -d '{
    "tmdb_id": 1396,
    "name": "Breaking Bad",
    "status": "Returning Series"
  }'
```

**验证**:
- [ ] 返回成功响应
- [ ] 剧集已添加到数据库

4. **刷新剧集**
```bash
curl -X POST http://localhost:8080/api/v1/shows/1/refresh
```

**验证**:
- [ ] 返回成功响应
- [ ] 数据已更新

5. **删除剧集**
```bash
curl -X DELETE http://localhost:8080/api/v1/shows/1
```

**验证**:
- [ ] 返回成功响应
- [ ] 数据库中记录已删除

---

### 场景5: 定时任务测试

**目标**: 测试定时调度功能

**步骤**:

1. **启动调度器**
```bash
./tmdb-crawler scheduler
```

**验证**:
- [ ] 服务正常启动
- [ ] 显示调度任务信息

2. **手动触发定时任务**
```bash
curl -X POST http://localhost:8080/api/v1/scheduler/daily-job
```

**验证**:
- [ ] 任务执行成功
- [ ] 数据已更新
- [ ] Telegraph已发布

---

### 场景6: 错误处理测试

**目标**: 测试系统的错误处理能力

**测试用例**:

1. **无效的TMDB ID**
```bash
./tmdb-crawler crawl 999999999
```

**预期**: 显示错误消息,不崩溃

2. **网络错误处理**
```
断开网络连接后尝试爬取
```

**预期**: 显示网络错误,有重试机制

3. **数据库连接失败**
```
停止PostgreSQL服务后运行系统
```

**预期**: 显示数据库错误,优雅退出

4. **无效的API请求**
```bash
curl http://localhost:8080/api/v1/shows/invalid
```

**预期**: 返回400错误和错误消息

---

## 性能测试

### 测试指标

1. **API响应时间**
- 单个剧集查询: < 100ms
- 剧集列表查询: < 200ms
- 爬取操作: < 30秒

2. **并发测试**
```bash
# 使用Apache Bench进行并发测试
ab -n 1000 -c 10 http://localhost:8080/api/v1/shows
```

**预期**:
- 所有请求成功
- 无错误响应
- 平均响应时间 < 500ms

3. **批量爬取性能**
```bash
./tmdb-crawler batch-crawl 1668,1396,1399
```

**预期**:
- 3个剧集在60秒内完成
- 并发爬取正常工作

---

## 测试报告模板

```markdown
## 集成测试报告

**测试日期**: 2026-01-11
**测试人员**: [姓名]
**测试环境**: 开发环境

### 测试结果总结

| 场景 | 状态 | 备注 |
|------|------|------|
| 完整爬取流程 | ✅ 通过 | - |
| Web界面操作 | ✅ 通过 | - |
| Telegraph发布 | ✅ 通过 | - |
| API集成测试 | ✅ 通过 | - |
| 定时任务测试 | ✅ 通过 | - |
| 错误处理测试 | ✅ 通过 | - |
| 性能测试 | ✅ 通过 | - |

### 发现的问题

1. [问题描述]
   - 严重程度: 高/中/低
   - 状态: 已修复/待修复

### 测试结论

系统功能完整,性能达标,可以投入使用。
```

---

## 自动化测试脚本

### Bash测试脚本示例

```bash
#!/bin/bash
# integration_test.sh

echo "开始集成测试..."

# 测试1: 启动服务
echo "测试1: 启动服务"
./tmdb-crawler server &
SERVER_PID=$!
sleep 5

# 测试2: API测试
echo "测试2: API测试"
curl http://localhost:8080/api/v1/shows | jq .

# 测试3: 爬取测试
echo "测试3: 爬取测试"
./tmdb-crawler crawl 1668

# 清理
kill $SERVER_PID

echo "测试完成"
```

---

## 持续集成配置

### GitHub Actions示例

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: tmdb_test
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v2
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Run tests
      run: |
        go test ./...
      env:
        DATABASE_URL: postgres://postgres:test@localhost/tmdb_test
```

---

## 注意事项

1. **测试数据隔离**: 使用独立的测试数据库
2. **API密钥安全**: 不要在测试代码中硬编码密钥
3. **清理测试数据**: 每次测试后清理数据
4. **并发测试**: 注意避免并发导致的冲突
5. **网络依赖**: 某些测试需要网络连接

---

## 下一步

完成集成测试后,可以进行:
- 性能优化
- 安全测试
- 用户验收测试
- 生产环境部署
