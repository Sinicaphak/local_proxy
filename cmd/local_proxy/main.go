package main

import (
	"flag"
	"log/slog"
	"net/http"

	proxy "local_proxy/internal/proxy"
)

func main() {
	configPath := flag.String("config", "/etc/local-proxy/config.yaml", "配置文件路径")
	flag.Parse()
	proxy.LoadConfig(*configPath)
	proxy.WatchConfig(*configPath) // 启动文件监视

	addr := proxy.ProxyConfig.SelfIP + ":" + proxy.ProxyConfig.SelfPort
	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				proxy.HandleTunneling(w, r)
			} else {
				proxy.HandleHTTP(w, r)
			}
		}),
	}
	if proxy.ProxyConfig.Direct {
		slog.Info("运行在直连模式")
	} else {
		slog.Info("代理服务已启动在 %s，上游代理: %s", "addr", addr, "upstream", proxy.GetUpstreamProxy())
	}
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("代理服务启动失败", "error", err)
	}
}
