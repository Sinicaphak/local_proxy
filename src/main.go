package main

import (
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

type Config struct {
	SelfIP     string `yaml:"self_ip"`
	SelfPort   string `yaml:"self_port"`
	TargetIP   string `yaml:"target_ip"`
	TargetPort string `yaml:"target_port"`
	Direct     bool   `yaml:"direct"`
}

var (
	proxyConfig Config
	configLock  = new(sync.RWMutex)
)

func LoadConfig(path string) {
	configLock.Lock()
	defer configLock.Unlock()
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("无法打开配置文件: %v", err)
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&proxyConfig); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
	log.Println("配置已重新加载")
}

func GetUpstreamProxy() string {
	configLock.RLock()
	defer configLock.RUnlock()
	if proxyConfig.Direct {
		return ""
	}
	return "http://" + proxyConfig.TargetIP + ":" + proxyConfig.TargetPort
}
