# TMDB剧集爬取系统 - 部署文档

## 目录

- [环境要求](#环境要求)
- [本地部署](#本地部署)
- [Docker部署](#docker部署)
- [服务器部署](#服务器部署)
- [配置说明](#配置说明)
- [常见问题](#常见问题)

---

## 环境要求

### 硬件要求

- **CPU**: 1核或更高
- **内存**: 512MB或更高
- **磁盘**: 100MB或更高

### 软件要求

#### 本地部署
- Go 1.21+
- PostgreSQL 15+ 或 SQLite 3+
- Git

#### Docker部署
- Docker 20.10+
- Docker Compose 2.0+

#### 服务器部署
- Linux (Ubuntu 20.04+ / CentOS 8+)
- Go 1.21+
- PostgreSQL 15+ 或 SQLite 3+
- systemd

### 第三方服务

- **TMDB API Key**: [申请地址](https://www.themoviedb.org/settings/api)
- **Telegraph账号**: [访问地址](https://telegra.ph/)

---

## 本地部署

### 1. 克隆项目

```bash
git clone https://github.com/yourusername/go-tmdb-crawler.git
cd go-tmdb-crawler
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件:

```bash
# 应用配置
APP_ENV=production
APP_PORT=8080
APP_LOG_LEVEL=info

# 数据库配置 (PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_USER=tmdb_user
DB_PASSWORD=your_password
DB_NAME=tmdb_db

# 或使用SQLite
DB_TYPE=sqlite
DB_PATH=./tmdb.db

# TMDB API
TMDB_API_KEY=your_tmdb_api_key

# Telegraph
TELEGRAPH_SHORT_NAME=tmdb_crawler
TELEGRAPH_AUTHOR_NAME=剧集更新助手

# 定时任务
ENABLE_SCHEDULER=true
SCHEDULE_CRON=0 8 * * *

# 日志
LOG_FILE_PATH=./logs/app.log
```

### 4. 初始化数据库

```bash
# PostgreSQL
psql -U tmdb_user -d tmdb_db -f migrations/001_init_schema.sql

# SQLite (自动创建)
# 数据库会在首次运行时自动创建
```

### 5. 编译运行

```bash
# 编译
go build -o tmdb-crawler main.go

# 运行
./tmdb-crawler server
```

### 6. 验证部署

```bash
# 检查服务状态
curl http://localhost:8080/health

# 检查API
curl http://localhost:8080/api/v1/shows
```

---

## Docker部署

### 1. 使用Docker Compose (推荐)

#### 步骤1: 准备配置文件

```bash
# 复制环境配置
cp .env.example .env

# 编辑配置
vim .env
```

#### 步骤2: 启动服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 查看服务状态
docker-compose ps
```

#### 步骤3: 初始化数据库

```bash
# 进入容器
docker-compose exec app bash

# 运行迁移
go run main.go migrate
```

#### 步骤4: 验证部署

```bash
# 检查服务
curl http://localhost:8080/health

# 检查API
curl http://localhost:8080/api/v1/shows
```

### 2. 使用Docker单独部署

#### 构建镜像

```bash
docker build -t tmdb-crawler:latest .
```

#### 运行容器

```bash
docker run -d \
  --name tmdb-crawler \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -e APP_ENV=production \
  -e DB_TYPE=sqlite \
  -e DB_PATH=/app/data/tmdb.db \
  -e TMDB_API_KEY=your_key \
  tmdb-crawler:latest
```

#### 查看日志

```bash
docker logs -f tmdb-crawler
```

---

## 服务器部署

### 1. 准备服务器

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 安装PostgreSQL (可选)
sudo apt install postgresql postgresql-contrib -y
```

### 2. 部署应用

#### 创建用户

```bash
sudo useradd -m -s /bin/bash tmdb
sudo su - tmdb
```

#### 克隆项目

```bash
cd /home/tmdb
git clone https://github.com/yourusername/go-tmdb-crawler.git
cd go-tmdb-crawler
```

#### 配置环境

```bash
cp .env.example .env
vim .env
```

#### 编译

```bash
go mod download
go build -o tmdb-crawler main.go
```

### 3. 配置systemd服务

创建服务文件:

```bash
sudo vim /etc/systemd/system/tmdb-crawler.service
```

内容:

```ini
[Unit]
Description=TMDB Crawler Service
After=network.target postgresql.service

[Service]
Type=simple
User=tmdb
WorkingDirectory=/home/tmdb/go-tmdb-crawler
Environment="APP_ENV=production"
ExecStart=/home/tmdb/go-tmdb-crawler/tmdb-crawler server
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### 4. 启动服务

```bash
# 重载systemd配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start tmdb-crawler

# 设置开机自启
sudo systemctl enable tmdb-crawler

# 查看状态
sudo systemctl status tmdb-crawler

# 查看日志
sudo journalctl -u tmdb-crawler -f
```

### 5. 配置Nginx反向代理 (可选)

安装Nginx:

```bash
sudo apt install nginx -y
```

配置文件:

```bash
sudo vim /etc/nginx/sites-available/tmdb-crawler
```

内容:

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api {
        proxy_pass http://localhost:8080/api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

启用配置:

```bash
sudo ln -s /etc/nginx/sites-available/tmdb-crawler /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

配置SSL (Let's Encrypt):

```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d your-domain.com
```

---

## 配置说明

### 环境变量详解

| 变量名 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| APP_ENV | 运行环境: development/production | development | 否 |
| APP_PORT | 服务端口 | 8080 | 否 |
| APP_LOG_LEVEL | 日志级别: debug/info/warn/error | info | 否 |
| DB_TYPE | 数据库类型: postgresql/sqlite | postgresql | 否 |
| DB_HOST | PostgreSQL主机地址 | localhost | 使用PostgreSQL时 |
| DB_PORT | PostgreSQL端口 | 5432 | 使用PostgreSQL时 |
| DB_USER | PostgreSQL用户名 | - | 使用PostgreSQL时 |
| DB_PASSWORD | PostgreSQL密码 | - | 使用PostgreSQL时 |
| DB_NAME | PostgreSQL数据库名 | tmdb | 使用PostgreSQL时 |
| DB_PATH | SQLite数据库路径 | ./tmdb.db | 使用SQLite时 |
| TMDB_API_KEY | TMDB API密钥 | - | **是** |
| TELEGRAPH_SHORT_NAME | Telegraph短名称 | tmdb_crawler | 否 |
| TELEGRAPH_AUTHOR_NAME | Telegraph作者名 | 剧集更新助手 | 否 |
| ENABLE_SCHEDULER | 是否启用定时任务 | true | 否 |
| SCHEDULE_CRON | 定时任务cron表达式 | 0 8 * * * | 否 |
| LOG_FILE_PATH | 日志文件路径 | ./logs/app.log | 否 |

### Cron表达式说明

```
SCHEDULE_CRON="0 8 * * *"
│ │ │ │ │
│ │ │ │ └─ 星期几 (0-6, 0=周日)
│ │ │ └─── 月份 (1-12)
│ │ └───── 日期 (1-31)
│ └─────── 小时 (0-23)
└───────── 分钟 (0-59)
```

常用示例:
- `0 8 * * *` - 每天早上8点
- `0 */6 * * *` - 每6小时
- `0 0 * * 1` - 每周一凌晨

---

## 常见问题

### 1. 端口被占用

**问题**: `Error: listen tcp :8080: bind: address already in use`

**解决**:
```bash
# 查看占用端口的进程
lsof -i :8080

# 杀死进程
kill -9 <PID>

# 或修改端口
export APP_PORT=8081
```

### 2. 数据库连接失败

**问题**: `connection refused` 或 `authentication failed`

**解决**:
```bash
# 检查PostgreSQL状态
sudo systemctl status postgresql

# 检查连接
psql -U tmdb_user -d tmdb_db -h localhost

# 检查防火墙
sudo ufw allow 5432
```

### 3. TMDB API调用失败

**问题**: `TMDB API call failed: 401 Unauthorized`

**解决**:
```bash
# 检查API Key
echo $TMDB_API_KEY

# 确认API Key有效
curl "https://api.themoviedb.org/3/movie/550?api_key=$TMDB_API_KEY"

# 重新申请API Key
# 访问 https://www.themoviedb.org/settings/api
```

### 4. 定时任务不执行

**问题**: 定时任务没有按时执行

**解决**:
```bash
# 检查scheduler是否启用
echo $ENABLE_SCHEDULER

# 检查cron表达式
echo $SCHEDULE_CRON

# 查看日志
tail -f logs/app.log

# 手动触发测试
curl -X POST http://localhost:8080/api/v1/crawler/refresh-all
```

### 5. Docker容器无法启动

**问题**: `docker-compose up` 失败

**解决**:
```bash
# 查看详细日志
docker-compose logs

# 重建容器
docker-compose down
docker-compose build
docker-compose up -d

# 清理并重启
docker-compose down -v
docker-compose up -d
```

### 6. 内存不足

**问题**: 爬取大量剧集时内存溢出

**解决**:
```bash
# 限制并发数
export MAX_CONCURRENT_CRAWLS=5

# 使用SQLite代替PostgreSQL
export DB_TYPE=sqlite

# 增加交换空间
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

### 7. 日志文件过大

**问题**: 日志文件占用过多磁盘空间

**解决**:
```bash
# 配置logrotate
sudo vim /etc/logrotate.d/tmdb-crawler

# 内容:
/home/tmdb/go-tmdb-crawler/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 tmdb tmdb
}
```

---

## 性能优化

### 1. 数据库优化

```sql
-- 创建索引
CREATE INDEX idx_shows_tmdb_id ON shows(tmdb_id);
CREATE INDEX idx_shows_status ON shows(status);
CREATE INDEX idx_episodes_show_id ON episodes(show_id);
CREATE INDEX idx_episodes_air_date ON episodes(air_date);

-- 定期清理日志
DELETE FROM crawl_logs WHERE created_at < NOW() - INTERVAL '30 days';
```

### 2. 应用优化

```bash
# 启用Gzip压缩
export ENABLE_GZIP=true

# 设置缓存
export CACHE_ENABLED=true
export CACHE_TTL=3600

# 调整worker数量
export WORKER_COUNT=4
```

### 3. 系统优化

```bash
# 调整文件描述符限制
vim /etc/security/limits.conf
* soft nofile 65536
* hard nofile 65536

# 调整内核参数
vim /etc/sysctl.conf
net.ipv4.tcp_max_syn_backlog = 4096
net.core.somaxconn = 1024
```

---

## 监控和维护

### 1. 健康检查

```bash
# 检查服务状态
curl http://localhost:8080/health

# 检查API
curl http://localhost:8080/api/v1/crawler/status
```

### 2. 日志监控

```bash
# 实时查看日志
tail -f logs/app.log

# 查看错误日志
grep ERROR logs/app.log

# 查看今天的日志
grep "$(date +%Y-%m-%d)" logs/app.log
```

### 3. 数据库维护

```bash
# 备份数据库
pg_dump -U tmdb_user tmdb_db > backup_$(date +%Y%m%d).sql

# 恢复数据库
psql -U tmdb_user tmdb_db < backup_20260111.sql

# SQLite备份
cp tmdb.db tmdb.db.backup_$(date +%Y%m%d)
```

### 4. 更新应用

```bash
# 拉取最新代码
git pull origin main

# 更新依赖
go mod download

# 重新编译
go build -o tmdb-crawler main.go

# 重启服务
sudo systemctl restart tmdb-crawler
```

---

## 安全建议

1. **使用环境变量** - 不要在代码中硬编码敏感信息
2. **限制访问** - 使用防火墙限制数据库访问
3. **HTTPS** - 生产环境务必使用HTTPS
4. **定期更新** - 及时更新系统和依赖包
5. **备份** - 定期备份数据库和配置文件
6. **监控** - 配置日志监控和告警

---

## 参考资料

- [Go官方文档](https://golang.org/doc/)
- [PostgreSQL文档](https://www.postgresql.org/docs/)
- [Docker文档](https://docs.docker.com/)
- [Nginx文档](https://nginx.org/en/docs/)
- [TMDB API文档](https://developers.themoviedb.org/3)

---

**文档版本**: 1.0  
**最后更新**: 2026-01-11  
**维护者**: TMDB剧集爬取系统团队
