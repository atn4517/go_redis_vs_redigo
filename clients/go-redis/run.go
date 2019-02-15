package go_redis_test

import (
	"sync"

	"go_redis_vs_redigo/clients"
	"go_redis_vs_redigo/utils"

	goRedis "github.com/go-redis/redis"
	"github.com/thoas/go-funk"
)

// Run 运行
func Run(conf *utils.ConfigObj) {
	logger := utils.GetLogger()
	oneNewClients, err := newClients(conf.GetRedisURL(), conf.GetSentinelURL(), conf.GetClusterURL())
	if err != nil {
		logger.Errorf("create redigo test Client Error: %+v", err)
	}
	defer oneNewClients.Close()
	configs := oneNewClients.GetBase()
	logger.Infof("go-redis client, redis config: %+v", configs.RedisOption)
	logger.Infof("go-redis client, sentinel config: %+v", configs.SentinelOption)
	logger.Infof("go-redis client, cluster config: %+v", configs.ClusterOption)
	clients.Run(conf, oneNewClients)
}

// goRedisTester 测试对象
type goRedisTester struct {
	baseTester     *clients.BaseTester
	sentinelClient *goRedis.Client
	redisClient    *goRedis.Client
	clusterClient  *goRedis.ClusterClient
	number         int
}

// newClients 创建一个新测试对象
func newClients(redisURL string, sentinelURL string, clusterURL string) (*goRedisTester, error) {
	baseTester, err := clients.NewTester(redisURL, sentinelURL, clusterURL)
	goRedis.SetLogger(nil)
	if err != nil {
		var oneNewClients *goRedisTester
		return oneNewClients, err
	}
	newGoRedisTester := goRedisTester{baseTester: baseTester}
	newGoRedisTester.createRedisClient()
	newGoRedisTester.createSentinelClient()
	newGoRedisTester.createClusterClient()
	return &newGoRedisTester, nil
}

// Close 关闭
func (g *goRedisTester) Close() {
	if g.redisClient != nil {
		utils.NeverMind(g.redisClient.Close())
	}
	if g.sentinelClient != nil {
		utils.NeverMind(g.sentinelClient.Close())
	}
	if g.clusterClient != nil {
		utils.NeverMind(g.clusterClient.Close())
	}
}

// GetBase 获取 baseTester
func (g *goRedisTester) GetBase() *clients.BaseTester {
	return g.baseTester
}

// Run goRedisTester 的 run 方法
func (g *goRedisTester) Run(c chan int) []error {
	var wg sync.WaitGroup
	errorList := [3]error{nil}
	type runFunc func(chan int) error
	funcArray := []runFunc{g.runRedis, g.runSentinel, g.runCluster}
	for index, oneFunc := range funcArray {
		wg.Add(1)
		go func(index int, oneFunc runFunc, c chan int) {
			defer wg.Done()
			err := oneFunc(c)
			errorList[index] = err
		}(index, oneFunc, c)
	}
	wg.Wait()
	return funk.Filter(errorList, func(err error) bool {
		return err != nil
	}).([]error)
}
