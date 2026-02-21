package proxy

import "sync"

type Config struct {
	SelfIP     string `yaml:"self_ip"`
	SelfPort   string `yaml:"self_port"`
	TargetIP   string `yaml:"target_ip"`
	TargetPort string `yaml:"target_port"`
	Direct     bool   `yaml:"direct"`
}

var (
	ProxyConfig Config
	ConfigLock  = new(sync.RWMutex)
)
