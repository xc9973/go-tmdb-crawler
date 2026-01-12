# 时区配置指南

## 概述

本项目支持统一的时区配置，确保日期范围查询和时间比较在不同环境下的一致性。

## 配置

### 环境变量

在 `.env` 文件中设置 `DEFAULT_TIMEZONE` 变量：

```bash
# 使用 UTC 时区（默认）
DEFAULT_TIMEZONE=UTC

# 使用中国时区
DEFAULT_TIMEZONE=Asia/Shanghai

# 使用美国东部时区
DEFAULT_TIMEZONE=America/New_York
```

### 支持的时区

项目使用 IANA 时区数据库（也称为 tz database）。常见时区包括：

- `UTC` - 协调世界时
- `Asia/Shanghai` - 中国标准时间
- `Asia/Tokyo` - 日本标准时间
- `America/New_York` - 美国东部时间
- `America/Los_Angeles` - 美国太平洋时间
- `Europe/London` - 英国时间
- `Europe/Paris` - 欧洲中部时间

完整时区列表请参考：[IANA Time Zone Database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)

## 使用场景

### 1. 今日更新查询

`GetTodayUpdates()` 方法使用配置的时区来确定"今天"的范围：

```go
// 在 UTC 时区下
// 今天：2024-01-15 00:00:00 UTC 到 2024-01-15 23:59:59 UTC

// 在 Asia/Shanghai 时区下
// 今天：2024-01-15 00:00:00 CST 到 2024-01-15 23:59:59 CST
```

### 2. 日期范围查询

`GetByDateRange()` 方法使用配置的时区来解释日期边界：

```go
startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

// 查询将返回在配置时区下，该日期范围内所有播出的剧集
```

### 3. 发布服务

发布服务使用配置的时区来生成标题和计算日期范围：

```go
// PublishTodayUpdates - 发布今日更新
// PublishWeeklyUpdates - 发布过去7天的更新
// PublishMonthlyUpdates - 发布过去30天的更新
```

## 日期范围定义

### 今日更新 (GetTodayUpdates)

- **范围**：[startOfDay, endOfDay)
- **定义**：从当天 00:00:00 开始，到第二天 00:00:00 结束（不包含）
- **示例**：
  - UTC 时区：`2024-01-15 00:00:00 UTC` <= air_date < `2024-01-16 00:00:00 UTC`
  - Asia/Shanghai 时区：`2024-01-15 00:00:00 CST` <= air_date < `2024-01-16 00:00:00 CST`

### 日期范围查询 (GetByDateRange)

- **范围**：[startDate, endDate]
- **定义**：从 startDate 00:00:00 开始，到 endDate 23:59:59 结束（包含）
- **示例**：
  - 查询 2024-01-01 到 2024-01-31
  - 实际范围：`2024-01-01 00:00:00` <= air_date <= `2024-01-31 23:59:59`

## 最佳实践

### 1. 使用 UTC 作为默认时区

对于国际化的应用，建议使用 UTC 作为默认时区：

```bash
DEFAULT_TIMEZONE=UTC
```

### 2. 根据用户群体选择时区

如果主要用户在特定地区，可以使用该地区的时区：

```bash
# 中国用户
DEFAULT_TIMEZONE=Asia/Shanghai

# 美国用户
DEFAULT_TIMEZONE=America/New_York
```

### 3. 数据库存储时间

- 所有时间在数据库中以 UTC 格式存储
- 查询时根据配置的时区进行转换
- 确保时间比较的一致性

### 4. 时区验证

应用启动时会验证配置的时区是否有效：

```go
if _, err := time.LoadLocation(cfg.Timezone.Default); err != nil {
    return nil, fmt.Errorf("invalid DEFAULT_TIMEZONE: %s", cfg.Timezone.Default)
}
```

## API 使用示例

### 查询今日更新

```bash
# 使用配置的时区查询今日更新
GET /api/v1/calendar/today
```

### 查询日期范围

```bash
# 查询指定日期范围的更新
GET /api/v1/calendar?start=2024-01-01&end=2024-01-31
```

### 发布今日更新

```bash
# 发布今日更新到 Telegraph
POST /api/v1/publish/today
```

## 注意事项

1. **时区一致性**：确保所有使用日期范围的服务都使用相同的时区配置
2. **边界测试**：在时区边界（如午夜）进行充分测试
3. **夏令时**：某些时区有夏令时，使用 IANA 时区数据库会自动处理
4. **数据库时区**：确保数据库连接也使用 UTC 时区，避免混淆

## 迁移指南

如果从旧版本升级，需要：

1. 在 `.env` 文件中添加 `DEFAULT_TIMEZONE` 配置
2. 重启应用以应用新配置
3. 验证日期查询结果是否符合预期
4. 如需更改时区，更新配置后重启即可

## 故障排查

### 问题：查询结果不正确

**可能原因**：
- 时区配置错误
- 数据库中的时间格式不正确

**解决方案**：
1. 检查 `.env` 文件中的 `DEFAULT_TIMEZONE` 配置
2. 验证数据库中的时间是否以 UTC 格式存储
3. 查看应用日志中的时区初始化信息

### 问题：时区验证失败

**可能原因**：
- 时区名称拼写错误
- 使用了不支持的时区名称

**解决方案**：
1. 检查时区名称是否正确（区分大小写）
2. 参考 IANA 时区数据库使用正确的时区名称
3. 测试时区：`time.LoadLocation("Asia/Shanghai")`
