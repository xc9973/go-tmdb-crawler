# TMDB剧集爬取系统 v2.0

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Status](https://img.shields.io/badge/Status-Active-success)

**一个基于Go语言的TMDB剧集自动爬取和管理系统**

[功能特性](#功能特性) • [快速开始](#快速开始) • [文档](#文档) • [贡献](#贡献)

</div>

---

## 📋 项目简介

TMDB剧集爬取系统是一个功能完整的服务端应用程序,用于自动爬取、管理和发布TMDB剧集数据。系统提供Web界面和RESTful API,支持定时任务、日历生成和Telegraph发布等功能。

### 主要特点

- 🎬 **自动爬取** - 从TMDB API自动获取最新剧集数据
- 📅 **日历生成** - 自动生成剧集更新日历
- 📝 **Telegraph发布** - 一键发布更新到Telegraph
- ⏰ **定时任务** - 自动定时刷新和发布
- 🎨 **Web界面** - 响应式设计的Web管理界面
- 🔌 **RESTful API** - 完整的API接口
- 🐳 **Docker支持** - 支持Docker和Docker Compose部署
- 💾 **多数据库** - 支持PostgreSQL和SQLite

---

## ✨ 功能特性

### 核心功能

#### 1. 剧集管理
- ✅ 添加、编辑、删除剧集
- ✅ 通过TMDB ID自动获取剧集信息
- ✅ 支持批量操作
- ✅ 剧集搜索和过滤
- ✅ 查看剧集详情和集数列表

#### 2. 自动爬取
- ✅ 从TMDB API爬取剧集数据
- ✅ 自动获取季度和集数信息
- ✅ 支持单个和批量爬取
- ✅ 爬取日志记录
- ✅ 错误处理和重试机制

#### 3. 日历生成
- ✅ 生成今日更新清单
- ✅ 生成未来N天的更新日历
- ✅ 导出Markdown格式
- ✅ Web界面展示

#### 4. Telegraph发布
- ✅ 自动生成发布内容
- ✅ 一键发布到Telegraph
- ✅ 避免重复发布
- ✅ 发布历史记录

#### 5. 定时任务
- ✅ 可配置的定时任务
- ✅ 自动刷新剧集数据
- ✅ 自动生成更新清单
- ✅ 自动发布到Telegraph

### 技术特性

- **高性能**: Go语言编写,性能优异
- **易部署**: 支持Docker一键部署
- **可扩展**: 模块化设计,易于扩展
- **类型安全**: 静态类型,编译时检查
- **并发支持**: 原生支持并发操作

---

## 🚀 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 15+ 或 SQLite 3+
- TMDB API Key

### 方式1: Docker部署 (推荐)

```bash
# 克隆项目
git clone https://github.com/yourusername/go-tmdb-crawler.git
cd go-tmdb-crawler

# 复制配置文件
cp .env.example .env

# 编辑.env文件,填入TMDB API Key
vim .env

# 启动服务
docker-compose up -d

# 访问Web界面
open http://localhost:8080
```

### 方式2: 本地运行

```bash
# 克隆项目
git clone https://github.com/yourusername/go-tmdb-crawler.git
cd go-tmdb-crawler

# 安装依赖
go mod download

# 配置环境变量
cp .env.example .env
vim .env

# 运行服务
go run main.go server

# 访问Web界面
open http://localhost:8080
```

### 方式3: 编译运行

```bash
# 编译
go build -o tmdb-crawler main.go

# 运行
./tmdb-crawler server
```

---

## 📖 使用指南

### 添加第一个剧集

1. 访问 http://localhost:8080
2. 点击"添加剧集"按钮
3. 输入TMDB ID (例如: 95479 - 咒术回战)
4. 点击"查询并添加"
5. 系统自动获取剧集信息

### 查看今日更新

1. 点击导航栏的"今日更新"
2. 查看今日更新的剧集列表
3. 点击"发布到Telegraph"按钮发布

### API使用示例

```bash
# 获取剧集列表
curl http://localhost:8080/api/v1/shows

# 添加新剧集
curl -X POST http://localhost:8080/api/v1/shows \
  -H "Content-Type: application/json" \
  -d '{"tmdb_id": 95479}'

# 刷新剧集
curl -X POST http://localhost:8080/api/v1/shows/1/refresh

# 获取今日更新
curl http://localhost:8080/api/v1/calendar/today
```

---

## 📚 文档

- **[API文档](docs/API.md)** - RESTful API完整文档
- **[部署文档](docs/DEPLOYMENT.md)** - 详细的部署指南
- **[用户手册](docs/USER_GUIDE.md)** - 用户使用手册
- **[数据迁移指南](MIGRATION_GUIDE.md)** - Python到Go的数据迁移

---

## 🏗️ 项目结构

```
go-tmdb-crawler/
├── api/                    # API处理器
│   ├── crawler.go         # 爬虫API
│   ├── show.go            # 剧集API
│   └── publish.go         # 发布API
├── config/                # 配置管理
│   └── config.go
├── models/                # 数据模型
│   ├── show.go
│   ├── episode.go
│   ├── crawl_log.go
│   └── telegraph.go
├── repositories/          # 数据仓储层
│   ├── show.go
│   ├── episode.go
│   └── crawl_log.go
├── services/              # 业务逻辑层
│   ├── tmdb.go           # TMDB API服务
│   ├── crawler.go        # 爬虫服务
│   ├── publisher.go      # 发布服务
│   └── scheduler.go      # 定时任务
├── web/                   # Web界面
│   ├── index.html        # 剧集列表
│   ├── show_detail.html  # 剧集详情
│   ├── today.html        # 今日更新
│   ├── logs.html         # 爬取日志
│   ├── css/              # 样式文件
│   └── js/               # JavaScript文件
├── migrations/            # 数据库迁移
│   └── 001_init_schema.sql
├── scripts/               # 脚本工具
│   └── migrate/          # 数据迁移脚本
├── docs/                  # 文档
│   ├── API.md
│   ├── DEPLOYMENT.md
│   └── USER_GUIDE.md
├── docker-compose.yml     # Docker Compose配置
├── Dockerfile            # Docker镜像配置
├── Makefile              # 构建脚本
├── main.go               # 程序入口
└── README.md             # 项目说明
```

---

## 🔧 配置说明

主要环境变量:

```bash
# 应用配置
APP_ENV=production          # 运行环境
APP_PORT=8080              # 服务端口

# 数据库配置
DB_TYPE=sqlite             # 数据库类型: postgresql/sqlite
DB_PATH=./tmdb.db          # SQLite数据库路径
DB_HOST=localhost          # PostgreSQL主机
DB_PORT=5432               # PostgreSQL端口
DB_USER=tmdb_user          # PostgreSQL用户名
DB_PASSWORD=password       # PostgreSQL密码
DB_NAME=tmdb_db            # PostgreSQL数据库名

# TMDB API
TMDB_API_KEY=your_key      # TMDB API密钥(必填)

# Telegraph
TELEGRAPH_SHORT_NAME=tmdb_crawler
TELEGRAPH_AUTHOR_NAME=剧集更新助手

# 定时任务
ENABLE_SCHEDULER=true
SCHEDULE_CRON=0 8 * * *    # 每天早上8点
```

---

## 🛠️ 开发指南

### 本地开发

```bash
# 安装依赖
go mod download

# 运行测试
go test ./...

# 运行服务
go run main.go server

# 代码格式化
go fmt ./...

# 代码检查
go vet ./...
```

### 构建生产版本

```bash
# 编译
go build -ldflags="-s -w" -o tmdb-crawler main.go

# 运行
./tmdb-crawler server
```

### 运行测试

```bash
# 单元测试
go test -v ./...

# 集成测试
go test -v ./tests/...

# 测试覆盖率
go test -cover ./...
```

---

## 📦 API端点

### 剧集管理
- `GET /api/v1/shows` - 获取剧集列表
- `GET /api/v1/shows/:id` - 获取剧集详情
- `POST /api/v1/shows` - 添加剧集
- `PUT /api/v1/shows/:id` - 更新剧集
- `DELETE /api/v1/shows/:id` - 删除剧集
- `POST /api/v1/shows/:id/refresh` - 刷新剧集

### 爬虫控制
- `POST /api/v1/crawler/show/:tmdb_id` - 爬取指定剧集
- `POST /api/v1/crawler/refresh-all` - 刷新所有剧集 (异步, 返回 task_id)
- `POST /api/v1/crawler/crawl-by-status` - 按状态爬取 (异步, 返回 task_id)
- `GET /api/v1/crawler/tasks/:id` - 查询异步任务状态
- `GET /api/v1/crawler/logs` - 获取爬取日志
- `GET /api/v1/crawler/status` - 获取爬虫状态

### 日历和发布
- `GET /api/v1/calendar/today` - 获取今日更新
- `GET /api/v1/calendar` - 获取更新日历
- `POST /api/v1/telegraph/publish` - 发布到Telegraph
- `GET /api/v1/telegraph/posts` - 获取发布历史

完整API文档请参考: [docs/API.md](docs/API.md)

---

## 🐳 Docker部署

### 使用Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重启服务
docker-compose restart
```

### 单独使用Docker

```bash
# 构建镜像
docker build -t tmdb-crawler .

# 运行容器
docker run -d \
  --name tmdb-crawler \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e TMDB_API_KEY=your_key \
  tmdb-crawler
```

---

## 🔄 数据迁移

从Python版本迁移到Go版本:

```bash
# 1. 导出Excel数据为CSV
cd go-tmdb-crawler/scripts/migrate
python3 export_excel_to_csv.py

# 2. 导入CSV到数据库
cd /Volumes/1disk/爬去/go-tmdb-crawler
go run scripts/migrate/import.go

# 3. 验证数据
sqlite3 tmdb.db "SELECT COUNT(*) FROM shows;"
```

详细迁移指南: [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md)

---

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议!

### 贡献流程

1. Fork本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交Pull Request

### 开发规范

- 遵循Go代码规范
- 添加必要的注释
- 编写单元测试
- 更新相关文档

---

## 📝 更新日志

### v2.0.0 (2026-01-11)

**新功能**:
- ✨ 完整的剧集管理功能
- ✨ 自动爬取TMDB数据
- ✨ 日历生成功能
- ✨ Telegraph发布功能
- ✨ 定时任务支持
- ✨ Web管理界面
- ✨ RESTful API
- ✨ Docker支持

**改进**:
- 🎨 优化用户界面
- ⚡ 提升性能
- 🐛 修复已知问题

---

## 📄 许可证

本项目采用MIT许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 👥 作者

**您的名字** - *初始工作* - [YourUsername](https://github.com/yourusername)

---

## 🙏 致谢

- [TMDB](https://www.themoviedb.org/) - 提供剧集数据API
- [Telegraph](https://telegra.ph/) - 提供发布平台
- [Gin](https://gin-gonic.com/) - Web框架
- [GORM](https://gorm.io/) - ORM框架

---

## 📞 联系方式

- **项目地址**: [https://github.com/yourusername/go-tmdb-crawler](https://github.com/yourusername/go-tmdb-crawler)
- **问题反馈**: [Issues](https://github.com/yourusername/go-tmdb-crawler/issues)
- **邮件**: your-email@example.com

---

## 🔗 相关链接

- [设计文档](设计文档2.0.md)
- [需求文档](需求文档.md)
- [任务文档](任务文档2.0.md)
- [API文档](docs/API.md)
- [部署文档](docs/DEPLOYMENT.md)

---

<div align="center">

**如果这个项目对您有帮助,请给个⭐️**

Made with ❤️ by [Your Name]

</div>
