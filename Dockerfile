# 构建阶段
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置国内 Go 模块代理，解决网络不通问题
ENV GOPROXY=https://goproxy.cn,direct

# 复制 go.mod 和 go.sum，先下载依赖（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译成静态可执行文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shortlink ./cmd/main.go

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates，用于 HTTPS 请求
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从构建阶段复制可执行文件
COPY --from=builder /app/shortlink .

# 复制配置文件（可选，也可以完全依赖环境变量）
COPY --from=builder /app/config ./config
COPY --from=builder /app/web ./web

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./shortlink"]