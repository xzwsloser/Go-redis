# Go Makefile

ProjectName := "Go-Redis"

MainFile := ./main.go
PROJECTBASE := $(shell pwd)
PROJECTBIN := $(PROJECTBASE)/bin

# CGO_ENABLED: 禁止使用 CGO,确保生成的 可执行文件不依赖于特定的 C 语言环境,便于跨平台
# GOOS=linux: 指定目标操作系统
# GOARCH=amd64: 指定目标指令集架构
build: clean
	@go mod tidy
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o $(PROJECTBIN)/$(ProjectName) $(MainFile)
	@chmod +x $(PROJECTBIN)/


# go fmt: 用于格式话项目代码
fmt:
	@go fmt $(PROJECTBASE)/...
	@go mod tidy

# go vet: 静态分析代码
vet:
	@go vet $(PROJECTBASE)/...
	@go mod tidy

# $>/dev/null 表示丢弃标准输入和错误输出
clean:
	@rm -rf $(PROJECTBIN)/*
depend:
	go mod download
# 构建并且运行 docker 容器
docker_run: docker_build
	sudo docker run -d --name go-redis -p 6399:6399 go-redis
# 构建 docker 镜像
docker_build: build
	sudo docker build -t go-redis .
# 删除 docker 容器并且删除镜像
docker_rm:
	-@sudo docker kill go-redis
	-@sudo docker rm go-redis
	-@sudo docker rmi go-redis
help:
	@echo "make build: 			构建项目"
	@echo "make fmt:   			格式化代码"
	@echo "make vet:   			静态代码分析"
	@echo "make clean: 			清除构建目标"
	@echo "make depend:  		下载依赖到本地缓存"
	@echo "make docker_run: 	构建镜像并且启动容器"
	@echo "make docker_build: 	构建镜像"
	@echo "make docker_rm:      删除容器和镜像"

.PHONY: fmt clean vet