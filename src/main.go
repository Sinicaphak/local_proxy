package main

import (
	"flag"
	"log/slog"
	"net/http"
	"sync"
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

func main() {
	configPath := flag.String("config", "/etc/local-proxy/config.yaml", "配置文件路径")
	flag.Parse()
	LoadConfig(*configPath)
	watchConfig(*configPath) // 启动文件监视

	addr := proxyConfig.SelfIP + ":" + proxyConfig.SelfPort
	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				HandleTunneling(w, r)
			} else {
				HandleHTTP(w, r)
			}
		}),
	}
	if proxyConfig.Direct {
		slog.Info("运行在直连模式")
	} else {
		slog.Info("代理服务已启动在 %s，上游代理: %s", "addr", addr, "upstream", GetUpstreamProxy())
	}
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("代理服务启动失败", "error", err)
	}
}
