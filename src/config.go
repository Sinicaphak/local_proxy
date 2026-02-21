package main

import (
	"log/slog"
	"os"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

func LoadConfig(path string) {
	configLock.Lock()
	defer configLock.Unlock()
	f, err := os.Open(path)
	if err != nil {
		slog.Error("无法打开配置文件", "error", err)
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&proxyConfig); err != nil {
		slog.Error("解析配置文件失败", "error", err)
	}

	mode := "代理"
	if proxyConfig.Direct {
		mode = "直连"
	}
	slog.Info("配置已重新加载, 运行在" + mode + "模式")
}

func watchConfig(path string) {
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			slog.Error("创建文件监视器失败", "error", err)
			return
		}
		defer watcher.Close()

		err = watcher.Add(path)
		if err != nil {
			slog.Error("添加文件到监视器失败", "path", path, "error", err)
			return
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					slog.Warn("检测到配置文件修改", "file", event.Name)
					LoadConfig(path)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("文件监视器错误", "error", err)
			}
		}
	}()
}
