# 服务器部署指南

本指南将帮助你在Linux服务器上部署TMDB剧集爬取系统。

## 目录

- [前置要求](#前置要求)
- [方式一: 使用Docker Compose部署(推荐)](#方式一使用docker-compose部署推荐)
- [方式二: 手动部署](#方式二手动部署)
- [方式三: 使用Nginx反向代理](#方式三使用nginx反向代理)
- [部署后配置](#部署后配置)
- [常见问题](#常见问题)

---

## 前置要求

### 服务器环境
- **操作系统**: Linux (Ubuntu 20.04+ / CentOS 7+ / Debian 10+)
- **内存**: 最低 512MB,推荐 1GB+
- **磁盘**: 最低 1GB 可用空间
- **权限**: sudo 或 root 权限

### 必需软件

#### 方式一(Docker部署):
```bash
# 安装 Docker
curl -fsSL https://get.docker.com | sh

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 验证安装
docker --version
docker-compose --version
```

#### 方式二(手动部署):
```bash
# 安装 Go 1.21+
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 安装 SQLite 3
sudo apt-get update
sudo apt-get install -y sqlite3

# 验证安装
go version
sqlite3 --version
```

---

## 方式一:使用Docker Compose部署(推荐)

这是最简单、最推荐的部署方式。

### 步骤1: 克隆项目

```bash
# 克隆项目到服务器
cd /opt
git clone https://github.com/xc9973/go-tmdb-crawler.git
cd go-tmdb-crawler
```

### 步骤2: 配置环境变量

```bash
# 复制环境变量示例文件
cp .env.example .env

# 编辑环境变量
nano .env
```

修改以下关键配置:

```bash
# Application Configuration
APP_ENV=production
APP_PORT=8888
APP_LOG_LEVEL=info

# Database Configuration (使用SQLite)
DB_TYPE=sqlite
DB_PATH=/app/data/tmdb.db

# TMDB API
TMDB_API_KEY=your_tmdb_api_key_here
TMDB_BASE_URL=https://api.themoviedb.org/3
TMDB_LANGUAGE=zh-CN

# Telegraph (可选)
TELEGRAPH_TOKEN=
TELEGRAPH_AUTHOR_NAME=剧集更新助手
TELEGRAPH_AUTHOR_URL=

# Scheduler (可选)
ENABLE_SCHEDULER=false
DAILY_CRON=0 8 * * *

# File Paths
WEB_DIR=/app/web
LOG_DIR=/app/logs
DATA_DIR=/app/data

# CORS (生产环境建议限制域名)
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=*
```

### 步骤3: 构建并启动服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 步骤4: 验证部署

```bash
# 检查服务是否运行
curl http://localhost:8888/api/v1/shows

# 应该返回:
# {"code":0,"data":{"items":[],"total":0},"message":"success"}
```

### 步骤5: 设置防火墙

```bash
# Ubuntu/Debian
sudo ufw allow 8888/tcp
sudo ufw reload

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=8888/tcp
sudo firewall-cmd --reload
```

### 常用Docker命令

```bash
# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f app

# 进入容器
docker-compose exec app bash

# 更新代码
git pull
docker-compose down
docker-compose up -d --build
```

---

## 方式二:手动部署

如果你不想使用Docker,可以手动部署。

### 步骤1: 克隆项目

```bash
cd /opt
git clone https://github.com/xc9973/go-tmdb-crawler.git
cd go-tmdb-crawler
```

### 步骤2: 安装依赖

```bash
# 下载Go模块依赖
go mod download

# 编译项目
go build -o tmdb-crawler main.go
```

### 步骤3: 配置环境变量

```bash
cp .env.example .env
nano .env
```

修改配置(参考Docker部署中的配置说明)

### 步骤4: 创建必要目录

```bash
mkdir -p logs data
chmod +x tmdb-crawler
```

### 步骤5: 创建Systemd服务

```bash
sudo nano /etc/systemd/system/tmdb-crawler.service
```

添加以下内容:

```ini
[Unit]
Description=TMDB Crawler Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/go-tmdb-crawler
Environment="APP_ENV=production"
ExecStart=/opt/go-tmdb-crawler/tmdb-crawler server
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### 步骤6: 启动服务

```bash
# 重新加载systemd配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start tmdb-crawler

# 设置开机自启
sudo systemctl enable tmdb-crawler

# 查看服务状态
sudo systemctl status tmdb-crawler

# 查看日志
sudo journalctl -u tmdb-crawler -f
```

---

## 方式三:使用Nginx反向代理

使用Nginx作为反向代理,可以提供更好的性能和安全性。

### 步骤1: 安装Nginx

```bash
sudo apt-get update
sudo apt-get install -y nginx
```

### 步骤2: 配置Nginx

```bash
sudo nano /etc/nginx/sites-available/tmdb-crawler
```

添加以下配置:

```nginx
server {
    listen 80;
    server_name your-domain.com;  # 替换为你的域名或服务器IP

    # 日志配置
    access_log /var/log/nginx/tmdb-crawler-access.log;
    error_log /var/log/nginx/tmdb-crawler-error.log;

    # 反向代理配置
    location / {
        proxy_pass http://localhost:8888;
        proxy_http_version 1.1;
        
        # WebSocket支持
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # 传递原始请求信息
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时配置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # 缓冲配置
        proxy_buffering off;
        proxy_request_buffering off;
    }

    # 静态文件缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://localhost:8888;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
```

### 步骤3: 启用站点

```bash
# 创建符号链接
sudo ln -s /etc/nginx/sites-available/tmdb-crawler /etc/nginx/sites-enabled/

# 测试配置
sudo nginx -t

# 重启Nginx
sudo systemctl restart nginx
```

### 步骤4: 配置SSL证书(推荐使用Let's Encrypt)

```bash
# 安装Certbot
sudo apt-get install -y certbot python3-certbot-nginx

# 获取并自动配置SSL证书
sudo certbot --nginx -d your-domain.com

# 自动续期
sudo certbot renew --dry-run
```

Certbot会自动修改Nginx配置,添加SSL支持。

---

## 部署后配置

### 1. 导入现有数据(可选)

如果你有旧的Excel数据需要导入:

```bash
# 使用Docker部署
docker-compose exec app go run scripts/migrate/import.go

# 手动部署
go run scripts/migrate/import.go
```

### 2. 访问Web界面

部署完成后,通过以下地址访问:

- **直接访问**: http://your-server-ip:8888
- **通过Nginx**: http://your-domain.com
- **HTTPS**: https://your-domain.com (如果配置了SSL)

### 3. 测试API

```bash
# 测试剧集列表API
curl http://localhost:8888/api/v1/shows

# 测试添加剧集
curl -X POST http://localhost:8888/api/v1/shows \
  -H "Content-Type: application/json" \
  -d '{"tmdb_id": 95479}'

# 测试今日更新
curl http://localhost:8888/api/v1/calendar/today
```

### 4. 监控日志

```bash
# Docker部署
docker-compose logs -f app

# 手动部署
tail -f logs/tmdb-crawler.log

# Systemd服务
sudo journalctl -u tmdb-crawler -f
```

---

## 常见问题

### 1. 端口被占用

**问题**: 启动时报错 "address already in use"

**解决**:
```bash
# 查找占用8888端口的进程
sudo lsof -i :8888

# 杀死进程
sudo kill -9 <PID>

# 或修改.env中的APP_PORT
```

### 2. 权限问题

**问题**: 无法写入数据库或日志文件

**解决**:
```bash
# Docker部署
sudo chown -R $USER:$USER /opt/go-tmdb-crawler

# 手动部署
sudo chown -R www-data:www-data /opt/go-tmdb-crawler
chmod -R 755 /opt/go-tmdb-crawler
```

### 3. 内存不足

**问题**: 服务器内存不足导致服务崩溃

**解决**:
```bash
# 创建Swap文件
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# 永久生效
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

### 4. 防火墙阻止访问

**问题**: 无法从外部访问服务

**解决**:
```bash
# 检查防火墙状态
sudo ufw status

# 允许端口
sudo ufw allow 8888/tcp
sudo ufw reload

# 或者临时关闭防火墙测试
sudo ufw disable
```

### 5. Docker容器无法启动

**问题**: Docker容器启动失败

**解决**:
```bash
# 查看详细日志
docker-compose logs app

# 重新构建
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# 检查Docker服务
sudo systemctl status docker
```

### 6. Nginx 502 Bad Gateway

**问题**: Nginx返回502错误

**解决**:
```bash
# 检查后端服务是否运行
sudo systemctl status tmdb-crawler
# 或
docker-compose ps

# 检查端口是否正确
netstat -tlnp | grep 8888

# 检查Nginx错误日志
sudo tail -f /var/log/nginx/tmdb-crawler-error.log
```

---

## 性能优化建议

### 1. 使用PostgreSQL数据库

对于生产环境,建议使用PostgreSQL替代SQLite:

```bash
# 修改.env
DB_TYPE=postgres
DB_HOST=postgres
DB_PORT=5432
DB_USER=tmdb
DB_PASSWORD=your_password
DB_NAME=tmdb_db

# 更新docker-compose.yml添加PostgreSQL服务
```

### 2. 启用缓存

考虑添加Redis缓存以提高性能:

```bash
# 在docker-compose.yml中添加Redis服务
redis:
  image: redis:alpine
  ports:
    - "6379:6379"
  volumes:
    - redis_data:/data
```

### 3. 设置日志轮转

防止日志文件过大:

```bash
sudo nano /etc/logrotate.d/tmdb-crawler
```

```
/opt/go-tmdb-crawler/logs/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 www-data www-data
    sharedscripts
}
```

---

## 安全建议

1. **修改默认端口**: 将8888改为其他不常用端口
2. **配置防火墙**: 只开放必要的端口
3. **使用HTTPS**: 配置SSL证书
4. **定期更新**: 定期更新系统和依赖
5. **备份数据**: 定期备份数据库
6. **监控日志**: 定期检查异常日志

---

## 备份与恢复

### 备份数据库

```bash
# SQLite备份
cp data/tmdb.db data/tmdb.db.backup.$(date +%Y%m%d)

# 定时备份(crontab)
0 2 * * * cp /opt/go-tmdb-crawler/data/tmdb.db /opt/backups/tmdb.db.$(date +\%Y\%m\%d)
```

### 恢复数据库

```bash
# 停止服务
docker-compose down

# 恢复备份
cp data/tmdb.db.backup.20260111 data/tmdb.db

# 重启服务
docker-compose up -d
```

---

## 更新部署

### Docker方式更新

```bash
cd /opt/go-tmdb-crawler
git pull
docker-compose down
docker-compose up -d --build
```

### 手动方式更新

```bash
cd /opt/go-tmdb-crawler
git pull
go build -o tmdb-crawler main.go
sudo systemctl restart tmdb-crawler
```

---

## 获取帮助

如果遇到问题:

1. 查看日志文件
2. 检查服务状态
3. 参考[GitHub Issues](https://github.com/xc9973/go-tmdb-crawler/issues)
4. 查看完整文档: [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)

---

**部署成功后,你就可以通过Web界面管理剧集了!** 🎉
