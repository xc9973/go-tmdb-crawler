# Docker部署配置总结

**任务**: 5.3 Docker部署配置  
**完成时间**: 2026-01-11  
**状态**: ✅ 已完成

---

## 已创建的文件

### 1. Dockerfile

**位置**: `go-tmdb-crawler/Dockerfile`  
**用途**: 定义Go应用的Docker镜像构建

**特点**:
- 多阶段构建,减小镜像体积
- 基于Alpine Linux,轻量高效
- 包含SQLite支持
- 暴露8080端口
- 自动创建数据目录

**构建命令**:
```bash
docker build -t tmdb-crawler:latest .
```

### 2. docker-compose.yml

**位置**: `go-tmdb-crawler/docker-compose.yml`  
**用途**: 定义多容器应用编排

**服务**:
- `tmdb-crawler`: Go应用主服务
- `nginx`: Nginx反向代理(可选)
- `postgres`: PostgreSQL数据库(可选)

**特点**:
- 支持多种部署模式(Profile)
- 数据持久化
- 健康检查
- 自动重启
- 网络隔离

### 3. .dockerignore

**位置**: `go-tmdb-crawler/.dockerignore`  
**用途**: 排除不需要的文件,减小构建上下文

**排除内容**:
- Git文件
- 文档文件
- 编辑器配置
- 编译产物
- 敏感配置
- 日志和数据

### 4. Nginx配置

**位置**: `go-tmdb-crawler/nginx/nginx.conf`  
**用途**: Nginx反向代理配置

**功能**:
- HTTP重定向到HTTPS
- SSL/TLS支持
- Gzip压缩
- 静态文件缓存
- API代理
- WebSocket支持
- 安全头设置

### 5. 部署文档

**位置**: `go-tmdb-crawler/DEPLOYMENT.md`  
**用途**: 完整的部署指南

**内容**:
- 环境要求
- 快速开始
- 配置说明
- 部署模式
- 常用命令
- 故障排查
- 生产环境部署

---

## 部署模式

### 模式1: 基础模式

**命令**:
```bash
docker-compose up -d
```

**特点**:
- 仅运行Go应用
- 直接暴露8080端口
- 使用SQLite数据库
- 适合开发测试

**访问**: http://localhost:8080

### 模式2: Nginx反向代理

**命令**:
```bash
docker-compose --profile with-nginx up -d
```

**特点**:
- 包含Nginx前端
- HTTPS支持
- 静态文件缓存
- 负载均衡能力

**访问**: https://localhost

### 模式3: PostgreSQL数据库

**命令**:
```bash
docker-compose --profile with-postgres up -d
```

**特点**:
- 使用PostgreSQL
- 更高性能
- 支持集群
- 适合生产环境

### 模式4: 完整部署

**命令**:
```bash
docker-compose --profile with-nginx --profile with-postgres up -d
```

**特点**:
- 所有服务启用
- 完整的生产配置
- 最高性能和安全性

---

## Makefile命令

### Docker相关命令

| 命令 | 说明 |
|------|------|
| `make docker-build` | 构建Docker镜像 |
| `make docker-run` | 启动Docker容器 |
| `make docker-stop` | 停止容器 |
| `make docker-down` | 停止并删除容器 |
| `make docker-logs` | 查看日志 |
| `make docker-rebuild` | 重新构建并启动 |
| `make docker-shell` | 进入容器Shell |
| `make docker-ps` | 查看容器状态 |
| `make docker-clean` | 清理Docker资源 |

---

## 配置文件结构

```
go-tmdb-crawler/
├── Dockerfile                 # Go应用镜像定义
├── docker-compose.yml         # 容器编排配置
├── .dockerignore             # 构建排除文件
├── DEPLOYMENT.md             # 部署文档
├── nginx/
│   ├── nginx.conf            # Nginx配置
│   └── ssl/                  # SSL证书目录
├── data/                     # 数据持久化目录
├── logs/                     # 日志目录
└── .env                      # 环境变量配置
```

---

## 环境变量配置

### 必需变量

```bash
TMDB_API_KEY=your_api_key_here
```

### 可选变量

```bash
# 应用配置
APP_ENV=production
APP_PORT=8080
APP_LOG_LEVEL=info

# 数据库配置
DB_TYPE=sqlite              # sqlite 或 postgres
DB_PATH=/root/data/tmdb.db

# TMDB API
TMDB_LANGUAGE=zh-CN

# Telegraph
TELEGRAPH_TOKEN=your_token

# 调度器
ENABLE_SCHEDULER=true
DAILY_CRON=0 8 * * *
```

---

## 数据持久化

### SQLite模式

**数据目录**: `./data`  
**数据库文件**: `/root/data/tmdb.db`

**备份**:
```bash
docker-compose exec tmdb-crawler cp /root/data/tmdb.db ./backup/
```

### PostgreSQL模式

**数据卷**: `postgres-data`  
**数据目录**: `/var/lib/postgresql/data`

**备份**:
```bash
docker-compose exec postgres pg_dump -U tmdb tmdb > backup.sql
```

---

## 网络配置

### 网络名称

`go-tmdb-crawler_tmdb-network`

### 网络类型

Bridge模式

### 端口映射

| 服务 | 容器端口 | 主机端口 |
|------|----------|----------|
| tmdb-crawler | 8080 | 8080 |
| nginx | 80 | 80 |
| nginx | 443 | 443 |
| postgres | 5432 | - (内部) |

---

## 健康检查

### Go应用

```yaml
healthcheck:
  test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", 
         "http://localhost:8080/api/v1/crawler/status"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### 检查命令

```bash
# 检查健康状态
docker inspect tmdb-crawler | grep -A 10 Health

# 手动测试
curl http://localhost:8080/api/v1/crawler/status
```

---

## 日志管理

### 查看日志

```bash
# 所有日志
docker-compose logs

# 特定服务
docker-compose logs tmdb-crawler

# 实时跟踪
docker-compose logs -f tmdb-crawler

# 最近行
docker-compose logs --tail=100 tmdb-crawler
```

### 日志配置

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

---

## 性能优化

### 资源限制

```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 1G
    reservations:
      cpus: '0.5'
      memory: 512M
```

### 构建优化

- 多阶段构建
- Alpine基础镜像
- .dockerignore排除
- 依赖缓存

### 运行优化

- Gzip压缩
- 静态文件缓存
- 连接池
- 健康检查

---

## 安全配置

### 1. 最小权限

- 使用非root用户
- 限制网络访问
- 文件权限控制

### 2. 敏感信息

- .env文件不提交
- 密钥使用环境变量
- Docker secrets

### 3. 网络安全

- 容器网络隔离
- TLS/SSL加密
- 安全头设置

### 4. 镜像安全

- 定期更新基础镜像
- 扫描漏洞
- 使用官方镜像

---

## 监控和维护

### 容器监控

```bash
# 资源使用
docker stats tmdb-crawler

# 容器状态
docker-compose ps

# 事件日志
docker events
```

### 数据备份

```bash
# 自动备份脚本
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
docker-compose exec tmdb-crawler \
  cp /root/data/tmdb.db ./backup/tmdb_$DATE.db
find ./backup -name "tmdb_*.db" -mtime +30 -delete
```

### 日志轮转

```bash
# 清理旧日志
docker-compose exec tmdb-crawler \
  find /root/logs -name "*.log" -mtime +7 -delete
```

---

## 更新部署

### 滚动更新

```bash
# 1. 拉取代码
git pull

# 2. 重新构建
docker-compose build

# 3. 重启服务
docker-compose up -d

# 4. 清理旧镜像
docker image prune -f
```

### 零停机更新

```bash
# 扩容
docker-compose up -d --scale tmdb-crawler=2

# 等待就绪
sleep 30

# 缩容
docker-compose up -d --scale tmdb-crawler=1
```

---

## 故障排查

### 常见问题

1. **容器无法启动**
   - 检查日志: `docker-compose logs`
   - 验证配置: `docker-compose config`
   - 检查端口: `lsof -i :8080`

2. **数据库连接失败**
   - SQLite: 检查文件权限
   - PostgreSQL: 检查容器状态

3. **健康检查失败**
   - 手动测试端点
   - 检查网络连接
   - 验证应用状态

---

## 测试验证

### 构建测试

```bash
# 构建镜像
docker-compose build

# 查看镜像
docker images | grep tmdb-crawler
```

### 运行测试

```bash
# 启动服务
docker-compose up -d

# 检查状态
docker-compose ps

# 测试访问
curl http://localhost:8080/api/v1/crawler/status
```

### 停止测试

```bash
# 停止服务
docker-compose down

# 清理资源
docker-compose down -v
```

---

## 最佳实践

### 开发环境

1. 使用SQLite简化配置
2. 挂载本地目录便于调试
3. 使用docker-compose快速启停

### 测试环境

1. 使用PostgreSQL模拟生产
2. 配置测试数据
3. 自动化测试脚本

### 生产环境

1. 使用PostgreSQL
2. 启用Nginx和HTTPS
3. 配置资源限制
4. 定期备份数据
5. 监控和告警

---

## 文件清单

### Docker配置文件

- ✅ `Dockerfile` - 镜像构建定义
- ✅ `docker-compose.yml` - 容器编排配置
- ✅ `.dockerignore` - 构建排除文件
- ✅ `nginx/nginx.conf` - Nginx配置
- ✅ `DEPLOYMENT.md` - 部署文档
- ✅ `Makefile` - 包含Docker命令

### 配置文件

- ✅ `.env.example` - 环境变量模板
- ✅ `.env` - 实际环境变量(不提交)

### 目录结构

- ✅ `nginx/ssl/` - SSL证书目录
- ✅ `data/` - 数据持久化目录
- ✅ `logs/` - 日志目录

---

## 验收标准

### 任务5.3验收标准

| 标准 | 状态 | 说明 |
|------|------|------|
| Dockerfile创建 | ✅ | 多阶段构建,Alpine基础 |
| docker-compose.yml | ✅ | 支持3种部署模式 |
| Nginx配置 | ✅ | HTTPS+反向代理 |
| .dockerignore | ✅ | 优化构建 |
| 部署文档 | ✅ | 完整详细 |
| Makefile命令 | ✅ | 9个Docker命令 |
| 配置示例 | ✅ | .env.example完整 |

---

## 总结

### 完成度: 100% ✅

任务5.3"Docker部署配置"已全部完成:

1. ✅ **Dockerfile**: 多阶段构建,优化镜像大小
2. ✅ **docker-compose.yml**: 支持4种部署模式
3. ✅ **Nginx配置**: 完整的反向代理配置
4. ✅ **.dockerignore**: 优化构建性能
5. ✅ **部署文档**: 详细的部署指南
6. ✅ **Makefile**: 添加9个Docker相关命令

### 特性

- 🚀 **快速部署**: 一键启动所有服务
- 🔒 **安全配置**: HTTPS、安全头、网络隔离
- 📊 **监控支持**: 健康检查、日志管理
- 🔄 **易于更新**: 滚动更新、零停机部署
- 📈 **可扩展**: 支持水平扩展
- 💾 **数据持久化**: 多种数据存储方案

### 下一步

1. 根据实际需求选择部署模式
2. 配置环境变量
3. 准备SSL证书(生产环境)
4. 部署并测试
5. 配置监控和备份

---

**文档版本**: 1.0  
**创建时间**: 2026-01-11  
**任务状态**: ✅ 已完成  
**维护者**: TMDB Crawler Team
