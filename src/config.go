package main

import (
	"log/slog"
	"os"
	"path/filepath"

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

		// 获取配置文件所在的绝对路径及其目录
		absPath, _ := filepath.Abs(path)
		dir := filepath.Dir(absPath)

		// 监视目录是 Linux 下处理配置文件更新最稳妥的方式
		err = watcher.Add(dir)
		if err != nil {
			slog.Error("无法监视配置目录", "dir", dir, "error", err)
			return
		}

		slog.Info("开始监视配置目录", "dir", dir)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// 只处理我们关心的那个配置文件
				eventPath, _ := filepath.Abs(event.Name)
				if eventPath != absPath {
					continue
				}

				// 逻辑判断：
				// Write: 直接覆盖写入
				// Create: 很多编辑器（如 vim）先删再建，或 rename 到此处
				// Chmod: 某些情况下权限变更也意味着内容同步完成
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Chmod) != 0 {
					slog.Warn("检测到配置文件变动", "file", event.Name, "op", event.Op.String())
					LoadConfig(absPath)
				}

				// 如果文件被删除了，或者发生了导致 Inode 丢失的操作
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					slog.Warn("配置文件被删除或移动，尝试重新加载", "file", event.Name)
					// 这里可以根据需求决定是否继续 LoadConfig
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("文件监视器运行中出错", "error", err)
			}
		}
	}()
}
