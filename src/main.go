package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
)

// 指定上游代理地址，如果为空字符串 "" 则直接放行（直连）
const upstreamProxy = "http://127.0.0.1:7890"

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	var destConn net.Conn
	var err error

	if upstreamProxy != "" {
		// --- 转发到上游代理 ---
		u, _ := url.Parse(upstreamProxy)
		// 连接上游代理服务器
		destConn, err = net.Dial("tcp", u.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		// 向代理发送 CONNECT 指令，告诉它我们要去哪里
		// 格式: CONNECT google.com:443 HTTP/1.1
		connectReq := "CONNECT " + r.Host + " HTTP/1.1\r\nHost: " + r.Host + "\r\n\r\n"
		destConn.Write([]byte(connectReq))
	} else {
		// --- 直接放行（直连） ---
		destConn, err = net.Dial("tcp", r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, _ := hijacker.Hijack()
	defer destConn.Close()
	defer clientConn.Close()

	// 建立双向拷贝
	done := make(chan bool, 2)
	go func() { io.Copy(destConn, clientConn); done <- true }()
	go func() { io.Copy(clientConn, destConn); done <- true }()
	<-done
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	transport := http.DefaultTransport.(*http.Transport)

	if upstreamProxy != "" {
		// 修改 Transport，让它把请求发往上游代理
		proxyURL, _ := url.Parse(upstreamProxy)
		transport.Proxy = http.ProxyURL(proxyURL)
	} else {
		// 直连
		transport.Proxy = nil
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	server := &http.Server{
		Addr: ":7897",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
	}
	log.Printf("代理服务已启动在 :7897，上游代理: %s", upstreamProxy)
	log.Fatal(server.ListenAndServe())
}
