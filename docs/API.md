# TMDB剧集爬取系统 - API文档

## 概述

TMDB剧集爬取系统提供RESTful API接口,用于管理剧集数据、控制爬虫任务、生成日历和发布Telegraph文章。

**基础URL**: `http://localhost:8080/api/v1`

**数据格式**: JSON

**字符编码**: UTF-8

---

## 认证

当前版本暂不需要认证。未来版本将添加API Key认证。

---

## 通用响应格式

### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "错误描述",
  "error": "详细错误信息"
}
```

### HTTP状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 剧集管理API

### 1. 获取剧集列表

**端点**: `GET /api/v1/shows`

**描述**: 获取所有剧集列表,支持分页和过滤

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码,默认1 |
| page_size | int | 否 | 每页数量,默认20 |
| status | string | 否 | 状态过滤: `returning`/`ended` |
| search | string | 否 | 搜索关键词 |

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/shows?page=1&page_size=20"
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "shows": [
      {
        "id": 1,
        "tmdb_id": 95479,
        "name": "咒术回战",
        "original_name": "Jujutsu Kaisen",
        "status": "returning",
        "overview": "虎杖悠仁是一位体育万能的高中生...",
        "poster_path": "/pWsD91G2R1Da3AKM3ymr3UoIfRb.jpg",
        "backdrop_path": "/fiVW06jE7D9OyPv9p9Y8Jkl9SfD.jpg",
        "first_air_date": "2020-10-03",
        "vote_average": 8.7,
        "genres": "动画,动作,奇幻",
        "created_at": "2026-01-11T10:00:00Z",
        "updated_at": "2026-01-11T10:00:00Z"
      }
    ],
    "total": 25,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 2. 获取剧集详情

**端点**: `GET /api/v1/shows/:id`

**描述**: 根据ID获取单个剧集的详细信息

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 剧集ID |

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/shows/1"
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "show": {
      "id": 1,
      "tmdb_id": 95479,
      "name": "咒术回战",
      "original_name": "Jujutsu Kaisen",
      "status": "returning",
      "overview": "虎杖悠仁是一位体育万能的高中生...",
      "poster_path": "/pWsD91G2R1Da3AKM3ymr3UoIfRb.jpg",
      "backdrop_path": "/fWV06jE7D9OyPv9p9Y8Jkl9SfD.jpg",
      "first_air_date": "2020-10-03",
      "vote_average": 8.7,
      "genres": "动画,动作,奇幻",
      "episodes_count": 59,
      "created_at": "2026-01-11T10:00:00Z",
      "updated_at": "2026-01-11T10:00:00Z"
    },
    "episodes": [
      {
        "id": 1,
        "season_number": 1,
        "episode_number": 1,
        "name": "两面宿傩",
        "overview": "虎杖悠仁是一位体育万能的高中生...",
        "still_path": "/x7YF4XLHsjI7wYqZ9p9Y8Jkl9SfD.jpg",
        "air_date": "2020-10-03",
        "vote_average": 8.5
      }
    ]
  }
}
```

---

### 3. 添加剧集

**端点**: `POST /api/v1/shows`

**描述**: 添加新剧集到数据库

**请求体**:
```json
{
  "tmdb_id": 95479,
  "name": "咒术回战",
  "original_name": "Jujutsu Kaisen",
  "status": "returning",
  "overview": "剧集简介...",
  "poster_path": "/pWsD91G2R1Da3AKM3ymr3UoIfRb.jpg",
  "backdrop_path": "/fWV06jE7D9OyPv9p9Y8Jkl9SfD.jpg",
  "first_air_date": "2020-10-03",
  "genres": "动画,动作,奇幻"
}
```

**响应示例**:
```json
{
  "code": 201,
  "message": "剧集添加成功",
  "data": {
    "id": 26,
    "tmdb_id": 95479,
    "name": "咒术回战"
  }
}
```

---

### 4. 更新剧集

**端点**: `PUT /api/v1/shows/:id`

**描述**: 更新剧集信息

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 剧集ID |

**请求体**:
```json
{
  "name": "咒术回战 第二季",
  "status": "returning",
  "overview": "更新的简介..."
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "剧集更新成功",
  "data": {
    "id": 1,
    "name": "咒术回战 第二季"
  }
}
```

---

### 5. 删除剧集

**端点**: `DELETE /api/v1/shows/:id`

**描述**: 删除指定剧集及其所有剧集详情

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 剧集ID |

**响应示例**:
```json
{
  "code": 200,
  "message": "剧集删除成功",
  "data": {
    "deleted_id": 1
  }
}
```

---

### 6. 刷新剧集

**端点**: `POST /api/v1/shows/:id/refresh`

**描述**: 从TMDB API刷新指定剧集的数据

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 剧集ID |

**响应示例**:
```json
{
  "code": 200,
  "message": "刷新成功",
  "data": {
    "show_id": 1,
    "episodes_added": 10,
    "episodes_updated": 5
  }
}
```

---

## 爬虫控制API

### 1. 爬取指定剧集

**端点**: `POST /api/v1/crawler/show/:tmdb_id`

**描述**: 爬取指定TMDB ID的剧集数据

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| tmdb_id | int | 是 | TMDB剧集ID |

**请求示例**:
```bash
curl -X POST "http://localhost:8080/api/v1/crawler/show/95479"
```

**响应示例**:
```json
{
  "code": 200,
  "message": "爬取成功",
  "data": {
    "tmdb_id": 95479,
    "name": "咒术回战",
    "seasons_count": 1,
    "episodes_count": 59,
    "crawl_time": "2.5s"
  }
}
```

---

### 2. 刷新所有剧集

**端点**: `POST /api/v1/crawler/refresh-all`

**描述**: 刷新所有连载中剧集的数据

**响应示例**:
```json
{
  "code": 200,
  "message": "刷新完成",
  "data": {
    "total_shows": 25,
    "success_count": 23,
    "failed_count": 2,
    "total_episodes": 789,
    "duration": "45s"
  }
}
```

---

### 3. 获取爬取日志

**端点**: `GET /api/v1/crawler/logs`

**描述**: 获取爬取日志列表

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码,默认1 |
| page_size | int | 否 | 每页数量,默认20 |
| status | string | 否 | 状态过滤: `success`/`failed` |
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/crawler/logs?page=1&page_size=20"
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "logs": [
      {
        "id": 1,
        "tmdb_id": 95479,
        "show_name": "咒术回战",
        "status": "success",
        "error_message": "",
        "started_at": "2026-01-11T10:00:00Z",
        "completed_at": "2026-01-11T10:00:03Z",
        "duration": "3s"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 4. 获取爬虫状态

**端点**: `GET /api/v1/crawler/status`

**描述**: 获取当前爬虫运行状态

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "is_running": false,
    "total_shows": 25,
    "returning_shows": 17,
    "last_crawl_time": "2026-01-11T10:00:00Z",
    "next_crawl_time": "2026-01-12T08:00:00Z"
  }
}
```

---

## 日历API

### 1. 获取今日更新

**端点**: `GET /api/v1/calendar/today`

**描述**: 获取今日更新的剧集列表

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "date": "2026-01-11",
    "episodes": [
      {
        "show_id": 1,
        "show_name": "咒术回战",
        "season_number": 1,
        "episode_number": 1,
        "name": "两面宿傩",
        "air_date": "2026-01-11",
        "poster_path": "/pWsD91G2R1Da3AKM3ymr3UoIfRb.jpg"
      }
    ],
    "total": 11
  }
}
```

---

### 2. 获取更新日历

**端点**: `GET /api/v1/calendar`

**描述**: 获取指定天数的更新日历

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| days | int | 否 | 天数,默认30天 |

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/calendar?days=30"
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "start_date": "2026-01-11",
    "end_date": "2026-02-10",
    "calendar": [
      {
        "date": "2026-01-11",
        "day_of_week": "星期日",
        "episodes_count": 11,
        "episodes": [ ... ]
      }
    ],
    "total_episodes": 100
  }
}
```

---

## Telegraph API

### 1. 发布今日更新

**端点**: `POST /api/v1/telegraph/publish`

**描述**: 将今日更新发布到Telegraph

**请求体**:
```json
{
  "title": "今日剧集更新 - 2026年1月11日",
  "author_name": "剧集更新助手"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "发布成功",
  "data": {
    "url": "https://telegra.ph/xxxxx-xxxxx",
    "title": "今日剧集更新 - 2026年1月11日",
    "views": 0,
    "published_at": "2026-01-11T10:00:00Z"
  }
}
```

---

### 2. 获取发布历史

**端点**: `GET /api/v1/telegraph/posts`

**描述**: 获取Telegraph发布历史

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码,默认1 |
| page_size | int | 否 | 每页数量,默认20 |

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/telegraph/posts?page=1&page_size=20"
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "posts": [
      {
        "id": 1,
        "url": "https://telegra.ph/xxxxx-xxxxx",
        "title": "今日剧集更新 - 2026年1月11日",
        "content_hash": "abc123...",
        "views": 150,
        "published_at": "2026-01-11T10:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 1001 | TMDB API调用失败 |
| 1002 | TMDB ID不存在 |
| 1003 | 剧集已存在 |
| 2001 | 数据库操作失败 |
| 2002 | 数据验证失败 |
| 3001 | Telegraph发布失败 |
| 3002 | Telegraph配置错误 |
| 4001 | 爬虫任务执行失败 |
| 4002 | 日志生成失败 |

---

## 使用示例

### 使用curl

```bash
# 获取剧集列表
curl "http://localhost:8080/api/v1/shows"

# 添加新剧集
curl -X POST "http://localhost:8080/api/v1/shows" \
  -H "Content-Type: application/json" \
  -d '{"tmdb_id": 95479, "name": "咒术回战"}'

# 刷新剧集
curl -X POST "http://localhost:8080/api/v1/shows/1/refresh"

# 获取今日更新
curl "http://localhost:8080/api/v1/calendar/today"

# 发布到Telegraph
curl -X POST "http://localhost:8080/api/v1/telegraph/publish" \
  -H "Content-Type: application/json" \
  -d '{"title": "今日更新"}'
```

### 使用JavaScript (fetch)

```javascript
// 获取剧集列表
fetch('http://localhost:8080/api/v1/shows')
  .then(response => response.json())
  .then(data => console.log(data));

// 添加新剧集
fetch('http://localhost:8080/api/v1/shows', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    tmdb_id: 95479,
    name: '咒术回战'
  })
})
.then(response => response.json())
.then(data => console.log(data));
```

### 使用Python (requests)

```python
import requests

# 获取剧集列表
response = requests.get('http://localhost:8080/api/v1/shows')
data = response.json()
print(data)

# 添加新剧集
response = requests.post('http://localhost:8080/api/v1/shows', json={
    'tmdb_id': 95479,
    'name': '咒术回战'
})
data = response.json()
print(data)
```

---

## 注意事项

1. **时区**: 所有日期时间使用UTC时区
2. **分页**: 所有列表API都支持分页,建议使用分页避免数据量过大
3. **错误处理**: 调用API时务必检查响应状态码和错误信息
4. **并发**: 当前版本不支持并发爬取,建议按顺序调用爬虫API
5. **限流**: TMDB API有调用限制,建议控制爬取频率

---

## 更新日志

### v1.0.0 (2026-01-11)
- 初始版本发布
- 实现所有核心API
- 支持剧集管理、爬虫控制、日历生成、Telegraph发布

---

**文档版本**: 1.0  
**最后更新**: 2026-01-11  
**维护者**: TMDB剧集爬取系统团队
