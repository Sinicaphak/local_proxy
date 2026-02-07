package main

import (
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
)

func watchConfig(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("创建文件监视器失败: %v", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("检测到配置文件修改: %s", event.Name)
					LoadConfig(path)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("文件监视器错误: %v", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		log.Fatalf("添加文件到监视器失败: %v", err)
	}
}

func main() {
	configPath := "../config/config.yaml"
	LoadConfig(configPath)
	watchConfig(configPath) // 启动文件监视

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
	log.Printf("代理服务已启动在 %s，上游代理: %s", addr, GetUpstreamProxy())
	log.Fatal(server.ListenAndServe())
}
