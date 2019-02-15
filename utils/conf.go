package utils

import (
	"fmt"
	"io/ioutil"
	
	"gopkg.in/yaml.v2"
)

type configObj struct {
	RedisURL     string `yaml:"redis_url,omitempty"`
	SentinelURL  string `yaml:"sentinel_url,omitempty"`
	ClusterURL   string `yaml:"cluster_url,omitempty"`
	SecondToStop uint   `yaml:"run_for_seconds,omitempty"`
}

// ConfigObj 配置参数对象
type ConfigObj struct {
	conf           *configObj
	configFileName string
}

// NewConfig 生成新的配置参数
func NewConfig(configFileName string) (*ConfigObj, error) {
	confObj, err := newConfig(configFileName)
	conf := &ConfigObj{
		conf:           confObj,
		configFileName: configFileName,
	}
	if err != nil {
		return conf, err
	}
	return conf, nil
}

// GetSecond 获取 SecondToStop
func (c *ConfigObj) GetSecond() uint {
	return c.conf.SecondToStop
}

// GetRedisURL 获取 GetRedisURL
func (c *ConfigObj) GetRedisURL() string {
	return c.conf.RedisURL
}

// GetSentinelURL 获取 GetSentinelURL
func (c *ConfigObj) GetSentinelURL() string {
	return c.conf.SentinelURL
}

// GetClusterURL 获取 GetClusterURL
func (c *ConfigObj) GetClusterURL() string {
	return c.conf.ClusterURL
}

func newConfig(configFileName string) (*configObj, error) {
	conf := new(configObj)
	yamlFile, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return conf, err
	}
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

// String 定制打印输出
func (c *ConfigObj) String() string {
	return fmt.Sprintf("CONFIG FILE NAME: %s, CONFIG INFO: %+v", c.configFileName, c.conf)
}
