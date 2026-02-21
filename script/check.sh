#!/bin/sh
# 检查 terminal/docker/apt/git/rosdep/pip 的代理配置

# 函数：打印文件内容（如果文件存在）
print_config() {
    local title="$1"
    local file_path="$2"

    printf "\n==================== %s ====================\n" "$title"
    printf "File: %s\n\n" "$file_path"
    if [ -f "$file_path" ]; then
        # 使用 cat 打印整个文件内容
        cat "$file_path"
    else
        printf "文件不存在。\n"
    fi
    printf "==================== %s END ====================\n" "$title"
}

# 1. 终端代理
# .bashrc 比较特殊，只看我们添加的代理部分
printf "\n==================== TERMINAL PROXY ====================\n"
printf "File: %s (代理相关行)\n\n" "$HOME/.bashrc"
if [ -f "$HOME/.bashrc" ]; then
    # 使用 grep 查找包含 proxy 的 export 行
    grep "export .*_proxy=" "$HOME/.bashrc" || printf "未找到终端代理设置。\n"
else
    printf "文件不存在。\n"
fi
printf "==================== TERMINAL PROXY END ====================\n"


# 2. Docker 代理
print_config "DOCKER PROXY" "/etc/docker/daemon.json"

# 3. APT 代理
print_config "APT PROXY" "/etc/apt/apt.conf"

# 4. Git 代理
printf "\n==================== GIT PROXY ====================\n"
printf "Command: git config --global http.proxy\n"
git config --global --get http.proxy || printf "未设置 http.proxy\n"
printf "\nCommand: git config --global https.proxy\n"
git config --global --get https.proxy || printf "未设置 https.proxy\n"
printf "==================== GIT PROXY END ====================\n"

# 5. rosdep 代理 (源)
print_config "ROSDEP SOURCE" "/etc/ros/rosdep/sources.list.d/20-default.list"

# 6. pip 代理
print_config "PIP PROXY" "$HOME/.pip/pip.conf"