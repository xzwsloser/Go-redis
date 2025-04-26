# 基于 alpine 构建 Docker 镜像
FROM alpine:latest
# 创建工作目录
WORKDIR /app
# 复制可执行程序
COPY bin/Go-Redis /app/Go-Redis
# 复制配置文件
COPY redis.yaml /app/redis.yaml
# 修改执行权限
RUN chmod +x /app/Go-Redis
# 配置启动命令
CMD ["/app/Go-Redis"]