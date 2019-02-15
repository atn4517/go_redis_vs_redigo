package main

import (
	"sync"

	goRedis "go_redis_vs_redigo/clients/go-redis"
	redigo "go_redis_vs_redigo/clients/redigo"
	"go_redis_vs_redigo/utils"
)

func main() {
	logger := utils.GetLogger()
	defer utils.NeverMind(logger.Sync())

	conf, err := utils.NewConfig("config.yaml")
	if err != nil {
		logger.Panicf("config fail: +%v", err)
	} else {
		logger.Infof("%+v", conf)
	}

	var wg sync.WaitGroup
	type runFunc func(*utils.ConfigObj)
	funcArray := []runFunc{redigo.Run, goRedis.Run}
	for _, oneFunc := range funcArray {
		wg.Add(1)
		go func(oneFunc runFunc, conf *utils.ConfigObj) {
			defer wg.Done()
			oneFunc(conf)
		}(oneFunc, conf)
	}
	wg.Wait()

	logger.Info("exiting")
}
