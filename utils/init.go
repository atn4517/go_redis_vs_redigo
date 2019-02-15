package utils

import (
	"sync"
	
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger
var onceForLogger sync.Once

// GetLogger 获取一个 日志 对象
func GetLogger() *zap.SugaredLogger {
	onceForLogger.Do(func() {
		_logger, _ := zap.NewDevelopment()
		logger = _logger.Sugar()
	})
	return logger
}

// NeverMind 忽略错误
func NeverMind(_ error){
	return
}