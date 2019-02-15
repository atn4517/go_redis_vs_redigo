package redigo_test

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"go_redis_vs_redigo/clients"
	"go_redis_vs_redigo/utils"

	redigoSentinel "github.com/FZambia/sentinel"
	redigoCluster "github.com/chasex/redis-go-cluster"
	redigo "github.com/gomodule/redigo/redis"
)

type dialFunc func() (redigo.Conn, error)

func getDial(option *clients.Option) dialFunc {
	var IPPortArray []string
	for ip, port := range option.Hosts {
		IPPortArray = append(IPPortArray, ip)
		IPPortArray = append(IPPortArray, strconv.Itoa(port))
	}
	server := strings.Join(IPPortArray, ":")
	password := option.Password
	db := option.DataBase
	dial := func() (redigo.Conn, error) {
		c, err := redigo.Dial("tcp", server)
		if err != nil {
			return nil, err
		}
		if password != "" {
			if _, errAuth := c.Do("AUTH", password); err != nil {
				utils.NeverMind(c.Close())
				return nil, errAuth
			}
		}
		if _, errSelect := c.Do("SELECT", db); err != nil {
			_error := c.Close()
			utils.NeverMind(_error)
			return nil, errSelect
		}
		return c, nil
	}
	return dial
}

type sentinelDialFunc func(string) (redigo.Conn, error)

func getSentinelDial() sentinelDialFunc {
	dial := func(server string) (redigo.Conn, error) {
		c, err := redigo.Dial("tcp", server)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	return dial
}

func (g *goRedisTester) createRedisClient() {
	if g.baseTester.RedisOption != nil {
		dial := getDial(g.baseTester.RedisOption)
		g.redisClient = &redigo.Pool{
			MaxIdle:         1000,
			MaxActive:       10000,
			Wait:            true,
			IdleTimeout:     240 * time.Second,
			MaxConnLifetime: 600 * time.Second,
			Dial:            dial,
			TestOnBorrow: func(c redigo.Conn, t time.Time) error {
				if _, err := c.Do("PING"); err != nil {
					logger := utils.GetLogger()
					logger.Errorf("%+v", err)
					return err
				}
				return nil
			},
		}
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
		g.sentinelClient = func(option *clients.Option, addrArray []string) *redigo.Pool {
			sentinelClient := &redigoSentinel.Sentinel{
				Addrs:      addrArray,
				MasterName: option.MasterName,
				Dial:       getSentinelDial(),
			}
			return &redigo.Pool{
				MaxIdle:     1000,
				MaxActive:   10000,
				Wait:        true,
				IdleTimeout: 240 * time.Second,
				Dial: func() (redigo.Conn, error) {
					masterAddr, err := sentinelClient.MasterAddr()
					if err != nil {
						return nil, err
					}
					IPPortArray := strings.Split(masterAddr, ":")
					port, err := strconv.Atoi(IPPortArray[1])
					if err != nil {
						return nil, err
					}
					option.Hosts = clients.HostIPPortPair{IPPortArray[0]: port}
					return getDial(option)()
				},
				TestOnBorrow: func(c redigo.Conn, t time.Time) error {
					if !redigoSentinel.TestRole(c, "master") {
						return errors.New("role check failed")
					}
					return nil
				},
			}
		}(g.baseTester.SentinelOption, AddrArray)
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
		var err error
		g.clusterClient, err = redigoCluster.NewCluster(&redigoCluster.Options{
			StartNodes:   AddrArray,
			ConnTimeout:  5 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			KeepAlive:    10000,
			AliveTime:    60 * time.Second,
		})
		if err != nil {
			logger := utils.GetLogger()
			logger.Errorf("create clusterClient Error: %+v", err)
		}
	}
}

func (g *goRedisTester) runRedis(c chan int) error {
	if g.redisClient == nil {
		return errors.New("redisClient is nil")
	}
	conn := g.redisClient.Get()
	defer func() {
		err := conn.Close()
		utils.NeverMind(err)
	}()
	runClient(conn, c, "redisClient")
	return nil
}

func (g *goRedisTester) runSentinel(c chan int) error {
	if g.sentinelClient == nil {
		return errors.New("sentinelClient is nil")
	}
	conn := g.sentinelClient.Get()
	defer func() {
		err := conn.Close()
		utils.NeverMind(err)
	}()
	runClient(conn, c, "sentinelClient")
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
	Do(string, ...interface{}) (interface{}, error)
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
		_, err := client.Do("SET", keyStr, valueStr)
		if err != nil {
			logger.Errorf("%s error: %+v", clientType, err)
		} else {
			logger.Infof("%s Set: %s -> %s", clientType, keyStr, valueStr)
		}
		res, err := client.Do("GET", keyStr)
		if err != nil {
			logger.Errorf("%s error: %+v", clientType, err)
		} else {
			resInt, transError := redigo.Int(res, err)
			if transError != nil {
				logger.Errorf("%s trans error: %+v", clientType, transError)
			} else {
				logger.Infof("%s Get: %s -> %d", clientType, keyStr, resInt)
			}
		}
		time.Sleep(100 * time.Millisecond)
		key = key + 1
	}
}
