# golang 客户端简单示例


## go-redis 示例

```
import (
	"fmt"
	"time"
	
	goRedis "github.com/go-redis/redis" // 引入 go-redis
)


// 创建单节点 Redis 服务连接对象
redisClient := goRedis.NewClient(&goRedis.Options{
	Addr:     "127.0.0.1:6379",
	Password: "password",
	DB:       1,
	PoolSize: 10000,
})

// 创建 sentinel 类型 Redis 服务连接对象
sentinelClient := goRedis.NewFailoverClient(&goRedis.FailoverOptions{
	MasterName:    "test_master",
	SentinelAddrs: []string{"127.0.0.1:26379", "127.0.0.1:26380", "127.0.0.1:26381"},
	Password:      "password",
	DB:            0,
	PoolSize:      10000,
})

// 创建 cluster 类型 Redis 服务连接对象
clusterClient := goRedis.NewClusterClient(&goRedis.ClusterOptions{
	Addrs: []string{
		"127.0.0.1:6379", "127.0.0.1:6380", "127.0.0.1:6381",
		"127.0.0.1:6382", "127.0.0.1:6383", "127.0.0.1:6384",
	},
	Password: "password",
	PoolSize: 10000,
})

// 定义接口
type goRedisClient interface {
	Set(string, interface{}, time.Duration) *goRedis.StatusCmd
	Get(string) *goRedis.StringCmd
}

// 使用 Redis 连接对象操作 Redis
func runClient(client goRedisClient) {
	err := client.Set("key", "value", 0).Err()
	if err != nil {
		doSomething(err)
	}
	res := client.Get(keyStr)
	fmt.Println(res.String())
}

runClient(redisClient)
runClient(sentinelClient)
runClient(clusterClient)
```

## redigo 示例

```
import (
	"time"
	"fmt"

    // 引入 redigo 及相关支持库
	redigoSentinel "github.com/FZambia/sentinel"
	redigoCluster "github.com/chasex/redis-go-cluster"
	redigo "github.com/gomodule/redigo/redis"
)

// 返回 Redis 的 Dial 函数，将密码等参数传入后包裹在闭包内
type dialFunc func() (redigo.Conn, error)
func getDial(addr string, password string, database int) dialFunc {
	dial := func() (redigo.Conn, error) {
		c, err := redigo.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		if password != "" {
			if _, errAuth := c.Do("AUTH", password); err != nil {
				_ = c.Close()
				return nil, errAuth
			}
		}
		if _, errSelect := c.Do("SELECT", database); err != nil {
			_ = c.Close()
			return nil, errSelect
		}
		return c, nil
	}
	return dial
}

// 创建单节点 Redis 连接池
redisClient := &redigo.Pool{
	MaxIdle:         1000,
	MaxActive:       10000,
	Wait:            true,
	IdleTimeout:     240 * time.Second,
	MaxConnLifetime: 600 * time.Second,
	Dial:            getDial("127.0.0.1:6379", "password", 1),
	TestOnBorrow: func(c redigo.Conn, t time.Time) error {
		if _, err := c.Do("PING"); err != nil {
			logger := utils.GetLogger()
			logger.Errorf("%+v", err)
			return err
		}
		return nil
	},
}

// 获取连接
conn := redisClient.Get()
// 将连接归还
defer conn.Close()
// 操作 redis
_, err1 := conn.Do("SET", "key", "1")
res, err2 := conn.Do("GET", "key")
resInt, transError := redigo.Int(res, err2)


// 返回 Sentinel 的 Dial 函数
type sentinelDialFunc func(string) (redigo.Conn, error)
func getSentinelDial() sentinelDialFunc {
	dial := func(addr string) (redigo.Conn, error) {
		c, err := redigo.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	return dial
}

// 创建 Sentinel 类型 Redis 服务连接对象，和单节点 Redis 连接池一样使用
sentinelClient := func() *redigo.Pool {
	_sentinelClient := &redigoSentinel.Sentinel{
		Addrs:      []string{"127.0.0.1:26379", "127.0.0.1:26380", "127.0.0.1:26381"},
		MasterName: "test_master",
		Dial:       getSentinelDial(),
	}
	return &redigo.Pool{
		MaxIdle:     1000,
		MaxActive:   10000,
		Wait:        true,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {
			masterAddr, err := _sentinelClient.MasterAddr()
			if err != nil {
				return nil, err
			}
			return getDial(masterAddr, "password", 1)()
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			if !redigoSentinel.TestRole(c, "master") {
				return errors.New("role check failed")
			}
			return nil
		},
	}
}()


// 创建 Cluster 类型 Redis 服务连接对象
clusterClient, err = redigoCluster.NewCluster(&redigoCluster.Options{
	StartNodes:   []string{
		"127.0.0.1:6379", "127.0.0.1:6380", "127.0.0.1:6381",
		"127.0.0.1:6382", "127.0.0.1:6383", "127.0.0.1:6384",
	},
	ConnTimeout:  5 * time.Second,
	ReadTimeout:  5 * time.Second,
	WriteTimeout: 5 * time.Second,
	KeepAlive:    10000,
	AliveTime:    60 * time.Second,
})

// Cluster 类型 Redis 服务连接对象使用自动连接池，无需自己进行申请/释放
_, err1 := clusterClient.Do("SET", "key", "1")
res, err2 := clusterClient.Do("GET", "key")
resInt, transError := redigo.Int(res, err2)

```