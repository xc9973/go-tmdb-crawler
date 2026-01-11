# 多阶段构建 - 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o tmdb-crawler main.go

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# 从构建阶段复制可执行文件
COPY --from=builder /app/tmdb-crawler .

# 复制web文件
COPY --from=builder /app/web ./web

# 创建数据目录
RUN mkdir -p /root/data /root/logs

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV APP_ENV=production
ENV APP_PORT=8080
ENV DB_TYPE=sqlite
ENV DB_PATH=/root/data/tmdb.db
ENV WEB_DIR=./web
ENV LOG_DIR=/root/logs
ENV DATA_DIR=/root/data

# 运行应用
CMD ["./tmdb-crawler", "server"]
