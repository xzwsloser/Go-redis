# Go-Redis
> 本项目利用 Golang 仿写 Redis,实现了一个单机NoSQL数据库
 
目前,本项目实现的数据库支持如下功能:

- 支持各种数据结构,包括 `string` , `list` 以及 `zset` 等
- 支持 `aof` 持久化以及 `aof` 重写功能
- 支持键的过期时间设置
- 支持事务

## 基本使用
> 可以直接使用 `redis-cli` 工具连接到目标数据库,并且进行相应的操作
### 构建方式
- 可以直接利用 `MakeFile` 构建:
```shell
make build
```
- 同时也可使用 `Dockerfile` 构建(该命令会创建一个镜像并且运行容器):
```shell
make docker_run
```
同时可以使用如下命令删除镜像和容器:
```shell
make docker_rm
```
更多功能可以利用如下命令查看:
```shell
make help
```
### 配置文件
配置文件为项目根目录下的 `redis.yaml`,配置文件实例如下,可以对其中的选项进行修改:
```yaml
# 配置 Redis 服务器信息
Redis:
  Address: 127.0.0.1
  Port: 8080

# 配置 Log 信息
Log:
  Stdout: on
  File: off
  Filename: redis.log
  Color:  off
  Level:  Info

# 配置数据库信息
DB:
  Number: 16

# 配置 Aof 相关信息
Aof:
  Load: on
  AppendOnly: on
  AppendFileName: appendonly.aof
  AppendFileSync: always
```
## 测试
利用 `Redis` 官方提供的工具: `redis-benchmark` 对于数据库性能进行测试,利用如下命令对于数据库进行压力测试(使用的 aof 同步等级为 `everysec`):
```shell
redis-benchmark -h 127.0.0.1 -p 8080  -t set,get,mset,lpush,lpop,zadd -c 100  -n 100000 -q
```
最终得到如下结果:
```shell
# Go-Redis:
SET: 165016.50 requests per second, p50=0.295 msec                    
GET: 165837.48 requests per second, p50=0.295 msec                    
LPUSH: 166666.66 requests per second, p50=0.295 msec                    
LPOP: 159489.64 requests per second, p50=0.303 msec                    
ZADD: 160513.64 requests per second, p50=0.295 msec                    
MSET (10 keys): 162074.56 requests per second, p50=0.303 msec  

# Redis(commit-hash: 319bbcc1a780b836889a71b80313e039140b11d1)
SET: 175438.59 requests per second, p50=0.279 msec                    
GET: 172117.05 requests per second, p50=0.287 msec                    
LPUSH: 173010.38 requests per second, p50=0.287 msec                    
LPOP: 175131.36 requests per second, p50=0.279 msec                    
ZADD: 173010.38 requests per second, p50=0.287 msec                    
MSET (10 keys): 176678.45 requests per second, p50=0.287 msec 
```

本项目参考开源项目 [`Godis`](https://github.com/HDT3213/godis)