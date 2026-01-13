# 数据备份导入导出功能设计

**日期**: 2026-01-13
**版本**: 1.0
**作者**: Claude

---

## 1. 需求概述

### 1.1 功能目标
- 支持全量数据导出为 JSON 文件
- 支持从 JSON 文件导入恢复数据
- 提供手动触发的 Web 界面和 API

### 1.2 数据范围
全量备份包含以下表：
- `shows` - 剧集基本信息
- `episodes` - 剧集详情
- `crawl_logs` - 爬取日志
- `telegraph_posts` - Telegraph 发布记录

---

## 2. API 设计

### 2.1 导出 API

```
GET /api/v1/backup/export
```

**响应**:
- Content-Type: `application/json`
- Content-Disposition: `attachment; filename="tmdb-backup-YYYYMMDD-HHMMSS.json"`

**成功响应**: JSON 文件下载

**错误响应**:
```json
{
  "error": "导出失败",
  "details": "数据库查询错误: ..."
}
```

---

### 2.2 导入 API

```
POST /api/v1/backup/import
Content-Type: multipart/form-data
```

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | File | 是 | JSON 备份文件 |
| mode | string | 否 | 导入模式: `replace`(默认) / `merge` |

**成功响应**:
```json
{
  "success": true,
  "message": "导入成功",
  "details": {
    "shows_imported": 50,
    "episodes_imported": 1200,
    "crawl_logs_imported": 300,
    "telegraph_posts_imported": 100,
    "conflicts_skipped": 5
  }
}
```

**错误响应**:
```json
{
  "success": false,
  "error": "导入失败",
  "details": "JSON 格式错误: ..."
}
```

---

### 2.3 备份状态 API

```
GET /api/v1/backup/status
```

**响应**:
```json
{
  "last_backup": "2026-01-13T12:00:00Z",
  "stats": {
    "shows": 50,
    "episodes": 1200,
    "crawl_logs": 300,
    "telegraph_posts": 100
  }
}
```

---

## 3. 数据格式

### 3.1 导出 JSON 结构

```json
{
  "version": "1.0",
  "exported_at": "2026-01-13T12:00:00Z",
  "app_version": "2.0.0",
  "stats": {
    "shows": 50,
    "episodes": 1200,
    "crawl_logs": 300,
    "telegraph_posts": 100
  },
  "data": {
    "shows": [
      {
        "id": 1,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-13T12:00:00Z",
        "tmdb_id": 95479,
        "name": "咒术回战",
        "name_cn": "",
        "overview": "...",
        "poster_path": "/xxx.jpg",
        "backdrop_path": "/yyy.jpg",
        "first_air_date": "2020-10-02",
        "status": "Returning Series",
        "type": "anime",
        "custom_notes": "",
        "tracking": false
      }
    ],
    "episodes": [
      {
        "id": 1,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-13T12:00:00Z",
        "show_id": 1,
        "season_number": 1,
        "episode_number": 1,
        "air_date": "2020-10-02",
        "name": "咒术回战",
        "overview": "...",
        "still_path": "/zzz.jpg",
        "runtime": 24,
        "vote_average": 8.5,
        "vote_count": 100
      }
    ],
    "crawl_logs": [...],
    "telegraph_posts": [...]
  }
}
```

### 3.2 格式验证规则
| 字段 | 规则 |
|------|------|
| version | 必填，仅支持 "1.0" |
| exported_at | 必填，ISO 8601 格式 |
| data.shows | 必填，至少为空数组 |
| data.episodes | 可选 |
| data.crawl_logs | 可选 |
| data.telegraph_posts | 可选 |

---

## 4. 导入逻辑

### 4.1 导入顺序（处理外键依赖）

```
1. Shows          (episodes/telegraph_posts/crawl_logs 依赖 show_id)
2. TelegraphPosts (依赖 show_id)
3. Episodes       (依赖 show_id)
4. CrawlLogs      (依赖 show_id)
```

### 4.2 导入模式

| 模式 | 行为 |
|------|------|
| **replace** | 清空所有 4 张表 → 按顺序导入 |
| **merge** | 保留现有数据，ID 冲突时跳过该记录 |

### 4.3 ID 处理
- 保持原 ID：导入时显式指定 `id`，保证关联正确
- 使用 GORM 的 `Clauses(clause.OnConflict{...})` 处理冲突

### 4.4 事务保护
```go
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()
// ... 导入逻辑
tx.Commit()
```

---

## 5. 错误处理

### 5.1 导出错误
| 场景 | 处理 |
|------|------|
| 数据库查询失败 | 返回 500，记录日志 |
| JSON 序列化失败 | 返回 500，记录日志 |

### 5.2 导入错误
| 场景 | 处理 |
|------|------|
| 文件上传失败 | 返回 400 |
| JSON 格式错误 | 返回 400，具体错误位置 |
| 版本不匹配 | 返回 400，提示不支持 |
| 单表导入失败 | **整件事务回滚**，返回详细错误 |
| 部分记录冲突 | merge 模式下记录日志，返回冲突计数 |

---

## 6. Web 界面

### 6.1 位置
新增 `/web/backup.html` 页面，导航栏添加"数据备份"入口。

### 6.2 功能
- **导出区域**: 显示当前数据统计，"导出备份"按钮
- **导入区域**: 文件上传控件，模式选择（下拉框），"导入备份"按钮
- **状态显示**: 显示最近备份时间、数据统计

### 6.3 交互反馈
- 导出：loading 状态 → 自动下载文件
- 导入：上传进度 → loading → 成功/失败提示

---

## 7. 文件结构

```
services/
  └── backup/
      ├── backup.go          # 备份服务核心逻辑
      ├── export.go          # 导出逻辑
      └── import.go          # 导入逻辑

api/
  └── backup.go              # 备份 API 处理器

models/
  └── backup.go              # 备份数据模型

web/
  └── backup.html            # 备份管理页面
```

---

## 8. 安全考虑

- **认证**: 所有备份 API 需要登录认证
- **文件大小限制**: 导入文件最大 50MB
- **格式校验**: 严格的 JSON schema 验证
- **事务保护**: 导入失败完整回滚

---

## 9. 测试要点

- 导出后能成功导入
- 导入 ID 关联正确性
- replace 模式清空验证
- merge 模式冲突处理
- 事务回滚验证
- 大数据量性能测试

---

## 10. 未来扩展

- 支持 Excel 格式导出（可作为前端选项）
- 定时自动备份
- 增量备份（仅导出变更数据）
- 云存储集成（S3、阿里云 OSS）
