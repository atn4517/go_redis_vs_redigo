# golang �ͻ��˼�ʾ��


## go-redis ʾ��

```
import (
	"fmt"
	"time"
	
	goRedis "github.com/go-redis/redis" // ���� go-redis
)


// �������ڵ� Redis �������Ӷ���
redisClient := goRedis.NewClient(&goRedis.Options{
	Addr:     "127.0.0.1:6379",
	Password: "password",
	DB:       1,
	PoolSize: 10000,
})

// ���� Sentinel ���� Redis �������Ӷ���
sentinelClient := goRedis.NewFailoverClient(&goRedis.FailoverOptions{
	MasterName:    "test_master",
	SentinelAddrs: []string{"127.0.0.1:26379", "127.0.0.1:26380", "127.0.0.1:26381"},
	Password:      "password",
	DB:            0,
	PoolSize:      10000,
})

// ���� Cluster ���� Redis �������Ӷ���
clusterClient := goRedis.NewClusterClient(&goRedis.ClusterOptions{
	Addrs: []string{
		"127.0.0.1:6379", "127.0.0.1:6380", "127.0.0.1:6381",
		"127.0.0.1:6382", "127.0.0.1:6383", "127.0.0.1:6384",
	},
	Password: "password",
	PoolSize: 10000,
})

// ����ӿ�
type goRedisClient interface {
	Set(string, interface{}, time.Duration) *goRedis.StatusCmd
	Get(string) *goRedis.StringCmd
}

// ʹ�� Redis ���Ӷ������ Redis
func runClient(client goRedisClient) {
	err := client.Set("key", "value", 0).Err()
	if err != nil {
		doSomething(err)
	}
	res := client.Get("key")
	fmt.Println(res.String())
}

runClient(redisClient)
runClient(sentinelClient)
runClient(clusterClient)
```

## redigo ʾ��

```
import (
	"time"
	"fmt"

    // ���� redigo �����֧�ֿ�
	redigoSentinel "github.com/FZambia/sentinel"
	redigoCluster "github.com/chasex/redis-go-cluster"
	redigo "github.com/gomodule/redigo/redis"
)

// ���� Redis �� Dial ������������Ȳ������������ڱհ���
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

// �������ڵ� Redis ���ӳ�
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

// ��ȡ����
conn := redisClient.Get()
// �����ӹ黹
defer conn.Close()
// ���� Redis
_, err1 := conn.Do("SET", "key", "1")
res, err2 := conn.Do("GET", "key")
resInt, transError := redigo.Int(res, err2)


// ���� Sentinel �� Dial ����
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

// ���� Sentinel ���� Redis �������Ӷ��󣬺͵��ڵ� Redis ���ӳ�һ��ʹ��
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


// ���� Cluster ���� Redis �������Ӷ���
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

// Cluster ���� Redis �������Ӷ���ʹ���Զ����ӳأ������Լ���������/�ͷ�
_, err1 := clusterClient.Do("SET", "key", "1")
res, err2 := clusterClient.Do("GET", "key")
resInt, transError := redigo.Int(res, err2)

```