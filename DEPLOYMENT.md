# Docker部署指南

## 目录

- [环境要求](#环境要求)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [部署模式](#部署模式)
- [常用命令](#常用命令)
- [故障排查](#故障排查)
- [生产环境部署](#生产环境部署)

---

## 环境要求

### 必需软件

- Docker: >= 20.10
- Docker Compose: >= 2.0

### 检查安装

```bash
docker --version
docker-compose --version
```

---

## 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd go-tmdb-crawler
```

### 2. 配置环境变量

复制环境变量模板:

```bash
cp .env.example .env
```

编辑`.env`文件,设置必要的配置:

```bash
# 必需: TMDB API密钥
TMDB_API_KEY=your_tmdb_api_key_here

# 可选: Telegraph令牌
TELEGRAPH_TOKEN=your_telegraph_token

# 数据库配置(默认使用SQLite)
DB_TYPE=sqlite
DB_PATH=/root/data/tmdb.db
```

### 3. 启动服务

**基础模式**(仅Go应用):

```bash
docker-compose up -d
```

**完整模式**(包含Nginx):

```bash
docker-compose --profile with-nginx up -d
```

**PostgreSQL模式**(使用PostgreSQL数据库):

```bash
docker-compose --profile with-postgres up -d
```

### 4. 验证部署

```bash
# 检查容器状态
docker-compose ps

# 查看日志
docker-compose logs -f tmdb-crawler

# 测试API
curl http://localhost:8080/api/v1/crawler/status
```

---

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 | 必需 |
|--------|------|--------|------|
| `APP_ENV` | 运行环境 | `production` | 否 |
| `APP_PORT` | 应用端口 | `8080` | 否 |
| `APP_LOG_LEVEL` | 日志级别 | `info` | 否 |
| `DB_TYPE` | 数据库类型 | `sqlite` | 否 |
| `DB_PATH` | SQLite路径 | `/root/data/tmdb.db` | 否 |
| `TMDB_API_KEY` | TMDB API密钥 | - | **是** |
| `TMDB_LANGUAGE` | 语言设置 | `zh-CN` | 否 |
| `TELEGRAPH_TOKEN` | Telegraph令牌 | - | 否 |
| `ENABLE_SCHEDULER` | 启用调度器 | `true` | 否 |
| `DAILY_CRON` | 定时任务表达式 | `0 8 * * *` | 否 |
| `ADMIN_API_KEY` | 管理员API密钥 | - | 否(生产环境推荐) |

### 数据库配置

#### SQLite(默认)

```yaml
environment:
  - DB_TYPE=sqlite
  - DB_PATH=/root/data/tmdb.db
```

#### PostgreSQL

```yaml
environment:
  - DB_TYPE=postgres
  - DB_HOST=postgres
  - DB_PORT=5432
  - DB_NAME=tmdb
  - DB_USER=tmdb
  - DB_PASSWORD=your_password
```

---

## 部署模式

### 模式1: 基础模式

仅运行Go应用,直接暴露8080端口。

```bash
docker-compose up -d
```

**访问地址**: http://localhost:8080

**适用场景**:
- 开发测试
- 内网部署
- 反向代理由其他服务提供

### 模式2: Nginx反向代理

包含Nginx作为前端服务器,提供HTTPS支持。

```bash
# 1. 准备SSL证书
cp cert.pem nginx/ssl/
cp key.pem nginx/ssl/

# 2. 启动服务
docker-compose --profile with-nginx up -d
```

**访问地址**: https://localhost

**适用场景**:
- 生产环境
- 需要HTTPS
- 需要静态文件缓存

### 模式3: PostgreSQL数据库

使用PostgreSQL代替SQLite。

```bash
docker-compose --profile with-postgres up -d
```

**适用场景**:
- 生产环境
- 需要更高性能
- 需要数据库集群

### 模式4: 完整部署

包含所有服务(Nginx + PostgreSQL)。

```bash
# 1. 配置.env文件
DB_TYPE=postgres
DB_PASSWORD=secure_password

# 2. 启动服务
docker-compose --profile with-nginx --profile with-postgres up -d
```

---

## 常用命令

### 容器管理

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose stop

# 重启服务
docker-compose restart

# 停止并删除容器
docker-compose down

# 停止并删除容器和卷
docker-compose down -v

# 重新构建镜像
docker-compose build

# 重新构建并启动
docker-compose up -d --build
```

### 日志查看

```bash
# 查看所有日志
docker-compose logs

# 查看特定服务日志
docker-compose logs tmdb-crawler
docker-compose logs nginx

# 实时跟踪日志
docker-compose logs -f tmdb-crawler

# 查看最近100行
docker-compose logs --tail=100 tmdb-crawler
```

### 容器操作

```bash
# 进入容器
docker-compose exec tmdb-crawler sh

# 在容器中执行命令
docker-compose exec tmdb-crawler ./tmdb-crawler --help

# 查看容器资源使用
docker stats tmdb-crawler

# 查看容器详细信息
docker inspect tmdb-crawler
```

### 数据管理

```bash
# 备份数据
docker-compose exec tmdb-crawler cp /root/data/tmdb.db ./backup/

# 恢复数据
docker-compose cp ./backup/tmdb.db tmdb-crawler:/root/data/tmdb.db

# 查看数据目录
docker-compose exec tmdb-crawler ls -lh /root/data
```

---

## 故障排查

### 容器无法启动

**问题**: 容器启动失败

**排查步骤**:

```bash
# 1. 查看容器状态
docker-compose ps

# 2. 查看日志
docker-compose logs tmdb-crawler

# 3. 检查配置
docker-compose config

# 4. 验证环境变量
docker-compose exec tmdb-crawler env | grep TMDB
```

### 无法访问Web界面

**问题**: 浏览器无法打开页面

**排查步骤**:

```bash
# 1. 检查端口是否被占用
lsof -i :8080

# 2. 检查防火墙
sudo ufw status

# 3. 测试本地访问
curl http://localhost:8080

# 4. 检查容器网络
docker network ls
docker network inspect go-tmdb-crawler_tmdb-network
```

### 数据库连接失败

**问题**: 无法连接到数据库

**SQLite排查**:

```bash
# 检查文件权限
docker-compose exec tmdb-crawler ls -lh /root/data/

# 检查数据库文件
docker-compose exec tmdb-crawler sqlite3 /root/data/tmdb.db "SELECT COUNT(*) FROM shows;"
```

**PostgreSQL排查**:

```bash
# 检查PostgreSQL容器
docker-compose logs postgres

# 测试连接
docker-compose exec postgres psql -U tmdb -d tmdb -c "SELECT 1;"
```

### 健康检查失败

**问题**: 容器不断重启

**排查步骤**:

```bash
# 1. 查看健康检查状态
docker inspect tmdb-crawler | grep -A 10 Health

# 2. 手动测试健康检查
docker-compose exec tmdb-crawler wget -O- http://localhost:8080/api/v1/crawler/status

# 3. 禁用健康检查(调试用)
# 编辑docker-compose.yml,注释掉healthcheck部分
```

---

## 生产环境部署

### 安全建议

1. **配置管理员认证**

生产环境强烈建议配置管理员API密钥,保护管理接口:

```bash
# 生成强密钥
openssl rand -base64 32

# 在.env文件中配置
ADMIN_API_KEY=your_generated_key_here
```

**认证方式**:

支持两种认证方式:

1. **Bearer Token** (推荐):
```bash
curl -H "Authorization: Bearer your_key" http://localhost:8080/api/v1/shows
```

2. **自定义Header**:
```bash
curl -H "X-Admin-API-Key: your_key" http://localhost:8080/api/v1/shows
```

**安全特性**:
- 失败尝试限制: 5次失败后IP封禁30分钟
- 本地客户端豁免: 127.0.0.1可配置本地密码
- 环境变量优先: ADMIN_API_KEY优先于配置文件
- Bcrypt支持: 支持bcrypt哈希存储密钥

**需要认证的端点**:
- POST /api/v1/shows - 创建剧集
- PUT /api/v1/shows/:id - 更新剧集
- DELETE /api/v1/shows/:id - 删除剧集
- POST /api/v1/shows/:id/refresh - 刷新剧集
- POST /api/v1/crawler/show/:tmdb_id - 爬取剧集
- GET /api/v1/crawler/logs - 查看日志

**公开端点** (无需认证):
- GET /api/v1/shows - 查询剧集列表
- GET /api/v1/shows/:id - 查询剧集详情
- GET /api/v1/calendar/today - 今日更新
- GET /api/v1/calendar - 日历视图
- GET /api/v1/crawler/status - 爬虫状态

2. **使用强密码**

```bash
# 生成随机密码
openssl rand -base64 32
```

2. **配置HTTPS**

```bash
# 使用Let's Encrypt获取免费证书
certbot certonly --standalone -d your-domain.com

# 复制证书到nginx目录
cp /etc/letsencrypt/live/your-domain.com/fullchain.pem nginx/ssl/cert.pem
cp /etc/letsencrypt/live/your-domain.com/privkey.pem nginx/ssl/key.pem
```

3. **限制网络访问**

```yaml
# docker-compose.yml
ports:
  - "127.0.0.1:8080:8080"  # 仅本地访问
```

4. **定期备份数据**

```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
docker-compose exec tmdb-crawler cp /root/data/tmdb.db ./backup/tmdb_$DATE.db
# 保留最近30天的备份
find ./backup -name "tmdb_*.db" -mtime +30 -delete
```

### 性能优化

1. **限制资源使用**

```yaml
# docker-compose.yml
services:
  tmdb-crawler:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

2. **配置日志轮转**

```yaml
services:
  tmdb-crawler:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

3. **启用缓存**

```nginx
# nginx.conf
# 静态文件缓存
location ~* \.(js|css|png|jpg|jpeg|gif|ico)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

### 监控和告警

1. **健康监控**

```bash
# 监控脚本
#!/bin/bash
while true; do
    if ! curl -f http://localhost:8080/api/v1/crawler/status; then
        echo "Service down, restarting..."
        docker-compose restart tmdb-crawler
    fi
    sleep 60
done
```

2. **日志监控**

```bash
# 监控错误日志
docker-compose logs -f | grep ERROR
```

3. **资源监控**

```bash
# 实时监控
docker stats tmdb-crawler --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"
```

---

## 更新部署

### 滚动更新

```bash
# 1. 拉取最新代码
git pull

# 2. 重新构建镜像
docker-compose build

# 3. 重启服务
docker-compose up -d

# 4. 清理旧镜像
docker image prune -f
```

### 零停机更新

```bash
# 1. 启动新容器
docker-compose up -d --scale tmdb-crawler=2 --no-recreate

# 2. 等待新容器就绪
sleep 30

# 3. 停止旧容器
docker-compose up -d --scale tmdb-crawler=1

# 4. 清理
docker-compose up -d --remove-orphans
```

---

## 附录

### 端口映射

| 服务 | 容器端口 | 主机端口 | 说明 |
|------|----------|----------|------|
| tmdb-crawler | 8080 | 8080 | Go应用 |
| nginx | 80 | 80 | HTTP |
| nginx | 443 | 443 | HTTPS |
| postgres | 5432 | - | PostgreSQL(内部) |

### 数据卷

| 卷 | 路径 | 说明 |
|----|------|------|
| ./data | /root/data | 数据目录 |
| ./logs | /root/logs | 日志目录 |
| postgres-data | /var/lib/postgresql/data | PostgreSQL数据 |

### 网络配置

- **网络名称**: tmdb-network
- **驱动类型**: bridge
- **子网**: 自动分配

### 相关文档

- [Docker官方文档](https://docs.docker.com/)
- [Docker Compose文档](https://docs.docker.com/compose/)
- [Nginx配置指南](https://nginx.org/en/docs/)

---

**文档版本**: 2.0  
**最后更新**: 2026-01-11  
**维护者**: TMDB Crawler Team
