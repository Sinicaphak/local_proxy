package main

import (
	"log"
	"net/http"
)

func main() {
	LoadConfig("../config/config.yaml")
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
		log.Println("运行在直连模式")
	} else {
		log.Printf("代理服务已启动在 %s，上游代理: %s", addr, GetUpstreamProxy())
	}
	log.Fatal(server.ListenAndServe())
}
