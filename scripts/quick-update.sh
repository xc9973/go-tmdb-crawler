#!/bin/bash
# 快速更新脚本 - 仅更新前端文件，无需重新构建镜像
# 适用于：只修改了 HTML/CSS/JS 等前端文件的情况

set -e

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_DIR="/opt/go-tmdb-crawler"
CONTAINER_NAME="tmdb-crawler"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  快速更新前端文件${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查是否以 root 权限运行
if [ "$EUID" -ne 0 ]; then 
    echo "错误: 请使用 root 权限运行此脚本"
    exit 1
fi

cd "$PROJECT_DIR"

# 1. 拉取最新代码
echo -e "${YELLOW}[1/3] 拉取最新代码...${NC}"
git fetch origin
git reset --hard origin/main
echo -e "${GREEN}✓ 代码已更新${NC}"
echo ""

# 2. 复制前端文件到容器
echo -e "${YELLOW}[2/3] 更新容器内的前端文件...${NC}"
docker cp web/welcome.html ${CONTAINER_NAME}:/app/web/
docker cp web/css/welcome.css ${CONTAINER_NAME}:/app/web/css/
docker cp web/js/welcome.js ${CONTAINER_NAME}:/app/web/js/
docker cp web/index.html ${CONTAINER_NAME}:/app/web/
docker cp web/js/api.js ${CONTAINER_NAME}:/app/web/js/
echo -e "${GREEN}✓ 前端文件已更新${NC}"
echo ""

# 3. 重启容器（可选）
echo -e "${YELLOW}[3/3] 重启容器...${NC}"
docker restart ${CONTAINER_NAME}
echo -e "${GREEN}✓ 容器已重启${NC}"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  更新完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "访问地址: http://localhost:8888"
echo ""
