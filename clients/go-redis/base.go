package go_redis_test

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"go_redis_vs_redigo/utils"

	goRedis "github.com/go-redis/redis"
)

func (g *goRedisTester) createRedisClient() {
	if g.baseTester.RedisOption != nil {
		var IPPortArray []string
		for ip, port := range g.baseTester.RedisOption.Hosts {
			IPPortArray = append(IPPortArray, ip)
			IPPortArray = append(IPPortArray, strconv.Itoa(port))
		}
		g.redisClient = goRedis.NewClient(&goRedis.Options{
			Addr:     strings.Join(IPPortArray, ":"),
			Password: g.baseTester.RedisOption.Password,
			DB:       g.baseTester.RedisOption.DataBase,
			PoolSize: 10000,
		})
	}
}

func (g *goRedisTester) createSentinelClient() {
	if g.baseTester.SentinelOption != nil {
		var AddrArray []string
		for ip, port := range g.baseTester.SentinelOption.Hosts {
			var IPPortArray []string
			IPPortArray = append(IPPortArray, ip)
			IPPortArray = append(IPPortArray, strconv.Itoa(port))
			Addr := strings.Join(IPPortArray, ":")
			AddrArray = append(AddrArray, Addr)
		}
		g.sentinelClient = goRedis.NewFailoverClient(&goRedis.FailoverOptions{
			MasterName:    g.baseTester.SentinelOption.MasterName,
			SentinelAddrs: AddrArray,
			Password:      g.baseTester.SentinelOption.Password,
			DB:            g.baseTester.SentinelOption.DataBase,
			PoolSize:      10000,
		})
	}
}

func (g *goRedisTester) createClusterClient() {
	if g.baseTester.ClusterOption != nil {
		var AddrArray []string
		for ip, port := range g.baseTester.ClusterOption.Hosts {
			var IPPortArray []string
			IPPortArray = append(IPPortArray, ip)
			IPPortArray = append(IPPortArray, strconv.Itoa(port))
			Addr := strings.Join(IPPortArray, ":")
			AddrArray = append(AddrArray, Addr)
		}
		g.clusterClient = goRedis.NewClusterClient(&goRedis.ClusterOptions{
			Addrs:    AddrArray,
			Password: g.baseTester.ClusterOption.Password,
			PoolSize: 10000,
		})
	}
}

func (g *goRedisTester) runRedis(c chan int) error {
	if g.redisClient == nil {
		return errors.New("redisClient is nil")
	}
	runClient(g.redisClient, c, "redisClient")
	return nil
}

func (g *goRedisTester) runSentinel(c chan int) error {
	if g.sentinelClient == nil {
		return errors.New("sentinelClient is nil")
	}
	runClient(g.sentinelClient, c, "sentinelClient")
	return nil
}

func (g *goRedisTester) runCluster(c chan int) error {
	if g.clusterClient == nil {
		return errors.New("clusterClient is nil")
	}
	runClient(g.clusterClient, c, "clusterClient")
	return nil
}

type goRedisClient interface {
	Set(string, interface{}, time.Duration) *goRedis.StatusCmd
	Get(string) *goRedis.StringCmd
}

func runClient(client goRedisClient, c chan int, clientType string) {
	logger := utils.GetLogger()
	key := 0
	for {
		if len(c) > 0 {
			<-c
			return
		}
		keyStr := strconv.Itoa(key)
		valueStr := strconv.Itoa(key + 100)
		err := client.Set(keyStr, valueStr, 0).Err()
		if err != nil {
			logger.Errorf("%s error: %+v", clientType, err)
		} else {
			logger.Infof("%s Set: %s -> %s", clientType, keyStr, valueStr)
		}
		res := client.Get(keyStr)
		logger.Infof("%s Get: %s -> %v", clientType, keyStr, res)
		time.Sleep(100 * time.Millisecond)
		key = key + 1
	}
}
