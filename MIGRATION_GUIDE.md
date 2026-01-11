# 数据迁移指南

本文档说明如何将Python版本的数据迁移到Go版本。

## 迁移概述

Python版本使用Excel存储数据,Go版本使用PostgreSQL数据库。迁移过程分为两个步骤:

1. **Python脚本**: 从Excel导出为CSV
2. **Go程序**: 从CSV导入到PostgreSQL

## 前置准备

### 1. 备份现有数据

```bash
# 备份整个data目录
cp -r py/data py/data.backup.$(date +%Y%m%d_%H%M%S)

# 或者备份单个文件
cp py/data/剧集列表.xlsx py/data/剧集列表.xlsx.backup
```

### 2. 确保Go服务已启动

```bash
cd go-tmdb-crawler
docker-compose up -d

# 检查服务状态
docker-compose ps
```

### 3. 验证数据库连接

```bash
# 进入容器
docker-compose exec tmdb-crawler sh

# 测试数据库连接
# (根据实际情况验证)
```

## 步骤1: Excel导出为CSV

### 使用Python脚本导出

```bash
cd go-tmdb-crawler/scripts/migrate

# 创建输出目录
mkdir -p output

# 运行导出脚本
python3 excel_to_csv.py
```

### 手动导出(备选)

如果Python脚本不可用,可以使用以下方法:

1. 打开 `py/data/剧集列表.xlsx`
2. 文件 -> 另存为 -> CSV (逗号分隔)
3. 保存为 `go-tmdb-crawler/scripts/migrate/output/shows.csv`

对于剧集详情,需要对每个Excel文件重复上述步骤。

## 步骤2: CSV导入到PostgreSQL

### 方法1: 使用Docker容器运行

```bash
cd go-tmdb-crawler

# 将CSV文件复制到容器
docker-compose cp scripts/migrate/output tmdb-crawler:/tmp/migrate

# 在容器中运行导入程序
docker-compose exec tmdb-crawler sh -c "cd /tmp/migrate && go run import.go"
```

### 方法2: 使用Makefile命令

```bash
cd go-tmdb-crawler

# 运行迁移
make migrate
```

### 方法3: 本地编译运行

```bash
cd go-tmdb-crawler/scripts/migrate

# 编译导入程序
go build -o import import.go

# 运行导入程序
./import
```

## 数据验证

### 验证剧集数据

```bash
# 进入Go服务容器
docker-compose exec tmdb-crawler sh

# 使用psql查询
psql $DATABASE_URL -c "SELECT COUNT(*) FROM shows;"
psql $DATABASE_URL -c "SELECT name, tmdb_id FROM shows LIMIT 10;"
```

### 验证剧集详情

```bash
psql $DATABASE_URL -c "SELECT COUNT(*) FROM episodes;"
psql $DATABASE_URL -c "SELECT s.name, e.season_number, e.episode_number FROM episodes e JOIN shows s ON e.show_id = s.id LIMIT 10;"
```

### 对比数据数量

```bash
# Python版本数据量
cd py/data
# 查看剧集列表.xlsx的行数(不包括表头)
# 使用Excel或LibreOffice打开查看

# Go版本数据量
cd go-tmdb-crawler
docker-compose exec tmdb-crawler psql $DATABASE_URL -c "
SELECT 
  (SELECT COUNT(*) FROM shows) as shows_count,
  (SELECT COUNT(*) FROM episodes) as episodes_count;
"
```

## 常见问题

### Q1: 导入程序找不到CSV文件

**问题**: `open scripts/migrate/output/shows.csv: no such file or directory`

**解决**:
```bash
# 确保输出目录存在
mkdir -p go-tmdb-crawler/scripts/migrate/output

# 确保CSV文件已生成
ls -lh go-tmdb-crawler/scripts/migrate/output/
```

### Q2: 数据库连接失败

**问题**: `connection refused` 或 `database does not exist`

**解决**:
```bash
# 确保数据库服务运行
docker-compose ps

# 确保数据库已创建
docker-compose exec tmdb-crawler psql $DATABASE_URL -c "SELECT 1;"

# 如果使用PostgreSQL容器
docker-compose exec postgres psql -U tmdb -d tmdb -c "SELECT 1;"
```

### Q3: 重复导入导致数据重复

**问题**: 第二次导入时出现重复数据

**解决**: 导入程序已经处理了重复数据,会自动跳过已存在的记录。或者可以先清空表:

```bash
# ⚠️ 警告: 此操作会删除所有数据
docker-compose exec tmdb-crawler psql $DATABASE_URL -c "
TRUNCATE TABLE episodes RESTART IDENTITY CASCADE;
TRUNCATE TABLE shows RESTART IDENTITY CASCADE;
"
```

### Q4: 字段名不匹配

**问题**: CSV列名与程序期望的不一致

**解决**: 检查CSV文件的表头,确保包含以下列:

**shows.csv必需列**:
- TMDB_ID
- 名称
- 原名
- 状态
- 简介
- 海报路径
- 背景路径
- 类型
- 评分

**episodes.csv必需列**:
- TMDB_ID
- 季数
- 集数
- 名称
- 简介
- 剧照路径
- 播出日期
- 评分

## 回滚方案

如果迁移出现问题,可以回滚到Python版本:

```bash
# 1. 停止Go服务
docker-compose down

# 2. 恢复Python数据
rm -rf py/data
cp -r py/data.backup.YYYYMMDD_HHMMSS py/data

# 3. 重启Python服务(如果需要)
cd py
./run.sh
```

## 完成检查清单

迁移完成后,请确认以下项目:

- [ ] CSV文件已生成
- [ ] 导入程序成功运行
- [ ] 数据库中有正确的记录数
- [ ] 随机抽查几条数据,内容正确
- [ ] Web界面可以正常显示剧集
- [ ] 爬虫功能正常工作
- [ ] 原始Python数据已备份

## 下一步

迁移完成后:

1. 运行完整的功能测试
2. 监控日志确认没有错误
3. 测试爬虫功能
4. 验证Telegraph发布
5. 确认定时任务正常

## 技术支持

如果遇到问题:

1. 查看日志: `docker-compose logs -f tmdb-crawler`
2. 检查数据库: `docker-compose exec tmdb-crawler psql $DATABASE_URL`
3. 查看任务文档: `任务文档2.0.md`
4. 查看设计文档: `设计文档2.0.md`

---

**最后更新**: 2026-01-11  
**版本**: 1.0
