// test_watch.go
package test

import (
	proxy "local_proxy/src"
	"testing"
)

var testFile string = "/etc/local-proxy/config.yaml"

func TestWatchConfig(t *testing.T) {
	proxy.WatchConfig(testFile)
	select {}
}
