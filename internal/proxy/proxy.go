package proxy

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/url"
)

func GetUpstreamProxy() string {
	ConfigLock.RLock()
	defer ConfigLock.RUnlock()
	if ProxyConfig.Direct {
		return ""
	}
	return "http://" + ProxyConfig.TargetIP + ":" + ProxyConfig.TargetPort
}

func HandleTunneling(w http.ResponseWriter, r *http.Request) {
	var destConn net.Conn
	var err error
	upstreamProxy := GetUpstreamProxy()

	if upstreamProxy != "" {
		u, _ := url.Parse(upstreamProxy)
		destConn, err = net.Dial("tcp", u.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		connectReq := "CONNECT " + r.Host + " HTTP/1.1\r\nHost: " + r.Host + "\r\n\r\n"
		destConn.Write([]byte(connectReq))

		br := bufio.NewReader(destConn)
		for {
			line, err := br.ReadString('\n')
			if err != nil || line == "\r\n" {
				break
			}
		}
	} else {
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

	done := make(chan bool, 2)
	go func() { io.Copy(destConn, clientConn); done <- true }()
	go func() { io.Copy(clientConn, destConn); done <- true }()
	<-done
}

func HandleHTTP(w http.ResponseWriter, req *http.Request) {
	transport := http.DefaultTransport.(*http.Transport)
	upstreamProxy := GetUpstreamProxy()

	if upstreamProxy != "" {
		proxyURL, _ := url.Parse(upstreamProxy)
		transport.Proxy = http.ProxyURL(proxyURL)
	} else {
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
