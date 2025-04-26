# Go-Redis
> 本项目利用 Golang 仿写 Redis,实现了一个单机NoSQL数据库
 
目前,本项目实现的数据库支持如下功能:

- 支持各种数据结构,包括 `string` , `list` 以及 `zset` 等
- 支持 `aof` 持久化以及 `aof` 重写功能
- 支持键的过期时间设置
- 支持事务

## 基本使用
> 可以直接使用 `redis-cli` 工具连接到目标数据库,并且进行相应的操作
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

本项目参考开源项目 [`Godis`](https://github.com/HDT3213/godis)