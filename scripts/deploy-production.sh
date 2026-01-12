#!/bin/bash
# 一键部署脚本 - 用于生产环境 /opt/go-tmdb-crawler
# 功能：拉取最新代码、重新构建镜像（如果需要）、启动 Docker 容器

set -e  # 遇到错误立即退出

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目配置
PROJECT_DIR="/opt/go-tmdb-crawler"
REPO_URL="https://github.com/xc9973/go-tmdb-crawler.git"
BRANCH="main"
IMAGE_NAME="go-tmdb-crawler"
CONTAINER_NAME="tmdb-crawler"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  TMDB Crawler 一键部署脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查是否以 root 权限运行
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}错误: 请使用 root 权限运行此脚本${NC}"
    echo "使用命令: sudo bash $0"
    exit 1
fi

# 1. 检查并创建项目目录
echo -e "${YELLOW}[1/6] 检查项目目录...${NC}"
if [ ! -d "$PROJECT_DIR" ]; then
    echo "项目目录不存在，正在克隆仓库..."
    mkdir -p "$PROJECT_DIR"
    git clone -b "$BRANCH" "$REPO_URL" "$PROJECT_DIR"
else
    echo "项目目录已存在: $PROJECT_DIR"
fi

cd "$PROJECT_DIR"
echo -e "${GREEN}✓ 当前目录: $(pwd)${NC}"
echo ""

# 2. 拉取最新代码
echo -e "${YELLOW}[2/6] 拉取最新代码...${NC}"
git fetch origin
git reset --hard origin/"$BRANCH"
echo -e "${GREEN}✓ 代码已更新到最新版本${NC}"
echo ""

# 3. 检查环境变量文件
echo -e "${YELLOW}[3/6] 检查环境配置...${NC}"
if [ ! -f ".env" ]; then
    if [ -f ".env.production.example" ]; then
        echo "创建 .env 文件从 .env.production.example..."
        cp .env.production.example .env
        echo -e "${YELLOW}⚠ 请编辑 .env 文件配置您的环境变量${NC}"
    else
        echo -e "${RED}错误: 未找到 .env 或 .env.production.example 文件${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✓ 环境配置文件已存在${NC}"
fi
echo ""

# 4. 停止并删除旧容器
echo -e "${YELLOW}[4/6] 停止旧容器...${NC}"
if [ "$(docker ps -aq -f name=${CONTAINER_NAME})" ]; then
    echo "停止并删除旧容器..."
    docker stop ${CONTAINER_NAME} 2>/dev/null || true
    docker rm ${CONTAINER_NAME} 2>/dev/null || true
    echo -e "${GREEN}✓ 旧容器已删除${NC}"
else
    echo "未找到运行中的容器"
fi
echo ""

# 5. 重新构建镜像
echo -e "${YELLOW}[5/6] 构建 Docker 镜像...${NC}"
echo "这可能需要几分钟时间..."
docker build -t ${IMAGE_NAME}:latest .
echo -e "${GREEN}✓ 镜像构建完成${NC}"
echo ""

# 6. 启动新容器
echo -e "${YELLOW}[6/6] 启动 Docker 容器...${NC}"
docker run -d \
    --name ${CONTAINER_NAME} \
    --restart unless-stopped \
    -p 8888:8888 \
    -v $(pwd)/data:/app/data \
    -v $(pwd)/logs:/app/logs \
    --env-file .env \
    ${IMAGE_NAME}:latest

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 容器启动成功${NC}"
else
    echo -e "${RED}✗ 容器启动失败${NC}"
    exit 1
fi
echo ""

# 显示容器状态
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "容器信息:"
docker ps -f name=${CONTAINER_NAME} --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "查看日志:"
echo "  docker logs -f ${CONTAINER_NAME}"
echo ""
echo "停止容器:"
echo "  docker stop ${CONTAINER_NAME}"
echo ""
echo "重启容器:"
echo "  docker restart ${CONTAINER_NAME}"
echo ""
echo -e "${GREEN}访问地址: http://localhost:8888${NC}"
echo ""
