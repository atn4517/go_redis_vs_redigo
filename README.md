## 1. 简介

本项目用于测试 [redigo](https://github.com/gomodule/redigo "redigo") 和 [go-redis](https://github.com/go-redis/redis) 两个 Golang redis 库的使用，实现的功能很简单，就是分别用两个库去连接各种类型的 redis，并以 100ms 的间隔去对 redis 发送 GET/SET 命令，其间你可以对连接的 redis 进行各种操作，测试客户端是否对 redis 的高可用进行了适配。

同时本项目连接 redis 的代码也可以作为示例代码。

## 2. 使用方法

1. cd $GOPATH/src

2. git clone https://git-sa.nie.netease.com/atn4517/go_redis_vs_redigo

3. cp go_redis_vs_redigo/config.yaml.example config.yaml

4. 对 config.yaml 进行修改

5. go build go_redis_vs_redigo

6. ./go_redis_vs_redigo
