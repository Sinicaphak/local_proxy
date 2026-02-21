// test_watch.go
package test

import (
	"local_proxy/internal/proxy"
	"testing"
)

var testFile string = "/etc/local-proxy/config.yaml"

func TestWatchConfig(t *testing.T) {
	proxy.WatchConfig(testFile)
	select {}
}
