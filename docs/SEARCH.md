# 搜索功能文档

## 概述

项目支持在 SQLite 和 PostgreSQL 数据库上进行不区分大小写的搜索功能。搜索功能会自动根据数据库类型选择合适的 SQL 语句。

## 实现原理

### 数据库类型检测

搜索功能通过 `db.Dialector.Name()` 检测当前使用的数据库类型：

```go
if r.db.Dialector.Name() == "sqlite" {
    // SQLite 特定实现
} else {
    // PostgreSQL 特定实现
}
```

### SQLite 实现

SQLite 不支持 `ILIKE` 操作符，因此使用 `LOWER()` 函数配合 `LIKE`：

```go
q := strings.ToLower(search)
query = query.Where("LOWER(name) LIKE ? OR LOWER(original_name) LIKE ?", 
    "%"+q+"%", "%"+q+"%")
```

**SQL 示例**：
```sql
SELECT * FROM shows 
WHERE LOWER(name) LIKE '%breaking%' 
   OR LOWER(original_name) LIKE '%breaking%';
```

### PostgreSQL 实现

PostgreSQL 原生支持 `ILIKE` 操作符（不区分大小写的 LIKE）：

```go
query = query.Where("name ILIKE ? OR original_name ILIKE ?", 
    "%"+search+"%", "%"+search+"%")
```

**SQL 示例**：
```sql
SELECT * FROM shows 
WHERE name ILIKE '%breaking%' 
   OR original_name ILIKE '%breaking%';
```

## 搜索功能

### Search 方法

按名称或原始名称搜索剧集：

```go
shows, total, err := showRepo.Search("breaking", 1, 10)
```

**参数**：
- `query`: 搜索关键词（不区分大小写）
- `page`: 页码（从 1 开始）
- `pageSize`: 每页数量

**返回**：
- `shows`: 匹配的剧集列表
- `total`: 总匹配数量
- `err`: 错误信息

### ListFiltered 方法

按状态和关键词过滤剧集：

```go
shows, total, err := showRepo.ListFiltered("Returning Series", "walking", 1, 10)
```

**参数**：
- `status`: 状态过滤（可选，空字符串表示不过滤）
- `search`: 搜索关键词（可选，空字符串表示不搜索）
- `page`: 页码（从 1 开始）
- `pageSize`: 每页数量

**返回**：
- `shows`: 匹配的剧集列表
- `total`: 总匹配数量
- `err`: 错误信息

## 搜索特性

### 1. 不区分大小写

搜索功能不区分大小写，以下搜索词返回相同结果：

```go
showRepo.Search("breaking", 1, 10)    // 小写
showRepo.Search("BREAKING", 1, 10)    // 大写
showRepo.Search("Breaking", 1, 10)    // 首字母大写
showRepo.Search("BrEaKiNg", 1, 10)    // 混合大小写
```

### 2. 部分匹配

搜索支持部分匹配，不需要完整输入：

```go
showRepo.Search("walk", 1, 10)   // 匹配 "The Walking Dead"
showRepo.Search("thing", 1, 10)  // 匹配 "Stranger Things", "Breaking News"
```

### 3. 多字段搜索

搜索同时检查 `name` 和 `original_name` 字段：

```go
// 如果剧集的 name 或 original_name 包含搜索词，都会被返回
showRepo.Search("game", 1, 10)  // 匹配 "Game of Thrones"
```

### 4. 分页支持

搜索结果支持分页：

```go
// 第一页，每页 10 条
shows, total, err := showRepo.Search("breaking", 1, 10)

// 第二页，每页 20 条
shows, total, err := showRepo.Search("breaking", 2, 20)
```

## 使用示例

### 基本搜索

```go
// 搜索包含 "breaking" 的剧集
shows, total, err := showRepo.Search("breaking", 1, 10)
if err != nil {
    log.Printf("Search failed: %v", err)
    return
}

log.Printf("Found %d shows", total)
for _, show := range shows {
    log.Printf("- %s", show.Name)
}
```

### 组合过滤

```go
// 搜索正在播出且包含 "dead" 的剧集
shows, total, err := showRepo.ListFiltered("Returning Series", "dead", 1, 10)
if err != nil {
    log.Printf("Search failed: %v", err)
    return
}

log.Printf("Found %d returning shows with 'dead'", total)
```

### 空搜索

```go
// 空搜索词返回所有剧集
shows, total, err := showRepo.Search("", 1, 10)

// 等价于
shows, total, err := showRepo.List(1, 10)
```

### 仅状态过滤

```go
// 仅按状态过滤，不搜索
shows, total, err := showRepo.ListFiltered("Returning Series", "", 1, 10)

// 等价于
shows, total, err := showRepo.ListByStatus("Returning Series", 1, 10)
```

## API 端点

### 搜索剧集

```http
GET /api/v1/shows?search=breaking&page=1&page_size=10
```

**参数**：
- `search`: 搜索关键词（可选）
- `page`: 页码（默认 1）
- `page_size`: 每页数量（默认 20）

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "tmdb_id": 1396,
        "name": "Breaking Bad",
        "original_name": "Breaking Bad",
        "status": "Ended"
      },
      {
        "id": 5,
        "tmdb_id": 12345,
        "name": "Breaking News",
        "original_name": "Breaking News",
        "status": "Ended"
      }
    ],
    "total": 2
  }
}
```

### 按状态过滤和搜索

```http
GET /api/v1/shows?status=Returning%20Series&search=walking&page=1&page_size=10
```

**参数**：
- `status`: 状态过滤（可选）
- `search`: 搜索关键词（可选）
- `page`: 页码（默认 1）
- `page_size`: 每页数量（默认 20）

## 性能考虑

### 索引建议

为了提高搜索性能，建议在数据库中创建适当的索引：

**SQLite**：
```sql
-- 创建索引以提高搜索性能
CREATE INDEX idx_shows_name_lower ON shows(LOWER(name));
CREATE INDEX idx_shows_original_name_lower ON shows(LOWER(original_name));
```

**PostgreSQL**：
```sql
-- PostgreSQL 的 ILIKE 可以使用普通索引
CREATE INDEX idx_shows_name ON shows(name);
CREATE INDEX idx_shows_original_name ON shows(original_name);

-- 或者使用表达式索引
CREATE INDEX idx_shows_name_lower ON shows(LOWER(name));
CREATE INDEX idx_shows_original_name_lower ON shows(LOWER(original_name));
```

### 搜索优化

1. **限制结果数量**：使用分页避免返回过多数据
2. **使用具体关键词**：更具体的搜索词返回更少的结果
3. **添加状态过滤**：结合状态过滤减少搜索范围
4. **考虑全文搜索**：对于大量数据，考虑使用全文搜索引擎

## 测试

项目包含完整的搜索功能测试：

```bash
# 运行搜索相关测试
go test -v ./repositories -run TestShowRepository
```

**测试覆盖**：
- ✅ 不区分大小写搜索
- ✅ 部分匹配
- ✅ 空搜索词
- ✅ 无匹配结果
- ✅ 状态过滤
- ✅ 组合过滤
- ✅ 分页功能

## 故障排查

### 问题：搜索不返回结果

**可能原因**：
1. 搜索词拼写错误
2. 数据库中没有匹配的数据
3. 大小写敏感（已解决，不应出现）

**解决方案**：
1. 检查搜索词拼写
2. 使用更短的搜索词进行部分匹配
3. 检查数据库中是否有数据

### 问题：搜索性能慢

**可能原因**：
1. 数据量过大
2. 缺少索引
3. 搜索词过于通用

**解决方案**：
1. 创建适当的索引
2. 使用更具体的搜索词
3. 添加状态过滤减少搜索范围
4. 考虑使用全文搜索引擎

### 问题：PostgreSQL 搜索失败

**可能原因**：
1. 数据库连接问题
2. ILIKE 操作符不支持（旧版本）

**解决方案**：
1. 检查数据库连接
2. 升级 PostgreSQL 到支持 ILIKE 的版本（9.1+）
3. 使用 `LOWER()` + `LIKE` 作为替代方案

## 最佳实践

### 1. 使用部分匹配

```go
// ✅ 推荐：使用部分匹配
showRepo.Search("walk", 1, 10)

// ⚠️ 不推荐：要求完整匹配
showRepo.Search("The Walking Dead", 1, 10)
```

### 2. 结合状态过滤

```go
// ✅ 推荐：结合状态过滤提高性能
showRepo.ListFiltered("Returning Series", "new", 1, 10)

// ⚠️ 不推荐：搜索所有数据
showRepo.Search("new", 1, 10)
```

### 3. 合理设置分页

```go
// ✅ 推荐：合理的分页大小
showRepo.Search("breaking", 1, 20)

// ⚠️ 不推荐：过大的分页
showRepo.Search("breaking", 1, 1000)
```

### 4. 处理空结果

```go
shows, total, err := showRepo.Search("xyz", 1, 10)
if err != nil {
    return err
}

if total == 0 {
    // 处理无结果情况
    return fmt.Errorf("no shows found")
}
```

## 相关文档

- [Show Repository API](../repositories/show.go)
- [API 文档](API.md)
- [数据库迁移指南](MIGRATION_GUIDE.md)

## 注意事项

1. **数据库兼容性**：搜索功能已针对 SQLite 和 PostgreSQL 进行优化
2. **性能**：对于大量数据，建议创建索引
3. **大小写**：搜索不区分大小写，无需转换输入
4. **分页**：始终使用分页避免返回过多数据
5. **空搜索**：空搜索词返回所有结果，注意性能影响
