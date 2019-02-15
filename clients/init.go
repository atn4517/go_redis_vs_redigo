package clients

import (
	"errors"
	"fmt"
	urlLib "net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"go_redis_vs_redigo/utils"
)

// tester 是客户端测试对象的 Interface
type tester interface {
	Close()
	Run(chan int) []error
	GetBase() *BaseTester
}

// Run 运行
func Run(conf *utils.ConfigObj, clients tester) {
	logger := utils.GetLogger()
	var wg sync.WaitGroup
	c := make(chan int, 3)
	wg.Add(1)
	go func() {
		defer wg.Done()
		errorList := clients.Run(c)
		for _, oneError := range errorList {
			logger.Errorf("find Error: %+v", oneError)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		second := conf.GetSecond()
		time.Sleep(time.Duration(second) * time.Second)
		for i := 0; i < 3; i++ {
			c <- 0
		}
	}()
	wg.Wait()
}

// HostIPPortPair ip 和端口键值对
type HostIPPortPair map[string]int

// Option redis 的连接选项
type Option struct {
	Hosts      HostIPPortPair
	Password   string
	DataBase   int
	MasterName string
}

// BaseTester 测试对象基类
type BaseTester struct {
	RedisURL       string
	SentinelURL    string
	ClusterURL     string
	RedisOption    *Option
	SentinelOption *Option
	ClusterOption  *Option
}

// NewTester 创建一个 BaseTester 对象
func NewTester(redisURL string, sentinelURL string, clusterURL string) (*BaseTester, error) {
	var (
		anyOne         bool
		err            error
		redisOption    *Option
		sentinelOption *Option
		clusterOption  *Option
		baseTester     *BaseTester
	)

	if redisURL != "" {
		redisOption, err = parseRedisURL(redisURL)
		if err != nil {
			return baseTester, err
		}
		anyOne = true
	}

	if sentinelURL != "" {
		sentinelOption, err = parseSentinelURL(sentinelURL)
		if err != nil {
			return baseTester, err
		}
		anyOne = true
	}

	if clusterURL != "" {
		clusterOption, err = parseClusterURL(clusterURL)
		if err != nil {
			return baseTester, err
		}
		anyOne = true
	}

	if !anyOne {
		return baseTester, errors.New("at last one")
	}

	baseTester = &BaseTester{
		RedisURL:       redisURL,
		SentinelURL:    sentinelURL,
		ClusterURL:     clusterURL,
		RedisOption:    redisOption,
		SentinelOption: sentinelOption,
		ClusterOption:  clusterOption,
	}
	return baseTester, nil
}

func baseParseURL(url string) (string, HostIPPortPair, string, string, int, error) {
	urlObj, err := urlLib.Parse(url)
	if err != nil {
		return "", nil, "", "", 0, fmt.Errorf("error parse url %s, detail: %+v", url, err)
	}

	scheme := strings.ToLower(urlObj.Scheme)
	hostName := urlObj.Hostname()
	portStr := urlObj.Port()
	port := 0
	if portStr == "" {
		port = 6379
	} else {
		_port, errTrans := strconv.Atoi(portStr)
		if errTrans != nil {
			return "", nil, "", "", 0, fmt.Errorf("redis port must between 1-65535, %s", portStr)
		}
		port = _port
	}
	host := HostIPPortPair{hostName: port}

	name := urlObj.User.Username()

	passWord, _ := urlObj.User.Password()

	dateBaseStr := strings.Trim(urlObj.EscapedPath(), "/ ")
	dateBase := 0
	if dateBaseStr != "" {
		_dateBase, errTrans := strconv.Atoi(dateBaseStr)
		if errTrans == nil {
			dateBase = _dateBase
		}
	}

	return scheme, host, name, passWord, dateBase, nil
}

// parseRedisURL 解析 redis url
func parseRedisURL(url string) (*Option, error) {
	scheme, host, _, passWord, dateBase, _ := baseParseURL(url)

	if scheme != "redis" {
		return nil, errors.New("redis url must starts with `redis://`")
	}

	return &Option{Hosts: host,
		Password:   passWord,
		DataBase:   dateBase,
		MasterName: "",
	}, nil
}

// parseSentinelURL 解析 sentinel url
func parseSentinelURL(url string) (*Option, error) {
	//  logger := utils.GetLogger()

	scheme, host, masterName, passWord, dateBase, _ := baseParseURL(url)

	if scheme != "sentinel" {
		return nil, errors.New("sentinel url must starts with `sentinel://`")
	}

	if masterName == "" {
		return nil, errors.New("sentinel url must have master name")
	}

	hosts := HostIPPortPair{}

	for h := range host {
		hostAndPorts := strings.Split(h, "+")
		for _, hostAndPort := range hostAndPorts {
			hostAndPortArray := strings.Split(hostAndPort, "*")
			_host := hostAndPortArray[0]
			_portStr := hostAndPortArray[1]
			_port, err := strconv.Atoi(_portStr)
			if err == nil {
				hosts[_host] = _port
			} else {
				return nil, errors.New("sentinel url port must be int")
			}
		}
		break
	}

	return &Option{Hosts: hosts,
		Password:   passWord,
		DataBase:   dateBase,
		MasterName: masterName,
	}, nil
}

// parseClusterURL 解析 sentinel url
func parseClusterURL(url string) (*Option, error) {
	// logger := utils.GetLogger()

	scheme, host, _, passWord, _, _ := baseParseURL(url)

	if scheme != "cluster" {
		return nil, errors.New("cluster url must starts with `cluster://`")
	}

	hosts := HostIPPortPair{}

	for h := range host {
		hostAndPorts := strings.Split(h, "+")
		for _, hostAndPort := range hostAndPorts {
			hostAndPortArray := strings.Split(hostAndPort, "*")
			_host := hostAndPortArray[0]
			_portStr := hostAndPortArray[1]
			_port, err := strconv.Atoi(_portStr)
			if err == nil {
				hosts[_host] = _port
			} else {
				return nil, errors.New("cluster url port must be int")
			}
		}
		break
	}

	return &Option{Hosts: hosts,
		Password:   passWord,
		DataBase:   0,
		MasterName: "",
	}, nil
}
