#!/bin/sh
# 这会移除#!/bin/sh
# 这会移除你的terminal/docker/apt/git/rosdep/pip的代理，并从备份中恢复。

NOW_USER=${SUDO_USER:-$(id -un 1000)} 
NOW_HOME=$(getent passwd "$NOW_USER" | cut -d: -f6) 

# 函数：从备份恢复文件
# 用法: restore_from_backup <文件路径>
restore_from_backup() {
    local conf_file="$1"
    
    # 删除由代理脚本创建的文件
    if [ -f "$conf_file" ]; then
        echo "删除代理配置文件: $conf_file"
        sudo rm "$conf_file"
    fi

    # 如果备份文件存在，则恢复它
    if [ -f "$conf_file.bak" ]; then
        echo "从备份恢复: $conf_file.bak -> $conf_file"
        sudo mv "$conf_file.bak" "$conf_file"
    else
        echo "未找到备份文件: $conf_file.bak，跳过恢复。"
    fi
}

echo "--- 开始清除代理配置 ---"

# 1. 终端代理
echo "清除终端代理..."
TERMINAL_CONF="$NOW_HOME/.bashrc"
# 删除由 set_proxy.sh 添加的代理块
sudo sed -i '/# Local Proxy Settings Start/,/# Local Proxy Settings End/d' "$TERMINAL_CONF"
# 取消对原始代理设置的注释
sudo sed -i -E 's/^#(export (http_proxy|https_proxy)=.*)/\1/g' "$TERMINAL_CONF"
echo "终端代理设置已恢复。请运行 'source ~/.bashrc' 或重开终端生效。"

# 2. Docker 代理
echo "清除 Docker 代理..."
DOCKER_CONF="/etc/docker/daemon.json"
# set_proxy.sh 直接覆盖了该文件，所以我们直接删除它。
# 如果用户之前有这个文件，他们需要手动恢复。
# 一个更安全的 set_proxy.sh 应该也为这个文件创建备份。
if [ -f "$DOCKER_CONF" ]; then
    echo "删除 Docker 代理配置文件: $DOCKER_CONF"
    sudo rm "$DOCKER_CONF"
    echo "重启 Docker 服务..."
    sudo systemctl daemon-reload
    sudo systemctl restart docker
else
    echo "未找到 Docker 代理配置文件，无需操作。"
fi


# 3. APT 代理
echo "恢复 APT 代理配置..."
APT_CONF="/etc/apt/apt.conf"
restore_from_backup "$APT_CONF"

# 4. Git 代理
echo "清除 Git 代理..."
# 检查代理是否存在，存在则取消
if git config --global --get http.proxy > /dev/null; then
    git config --global --unset http.proxy
    echo "已取消 http.proxy"
fi
if git config --global --get https.proxy > /dev/null; then
    git config --global --unset https.proxy
    echo "已取消 https.proxy"
fi

# 5. rosdep 代理 (源)
echo "恢复 rosdep 源..."
ROSDEP_CONF="/etc/ros/rosdep/sources.list.d/20-default.list"
restore_from_backup "$ROSDEP_CONF"

# 6. pip 代理
echo "恢复 pip 代理配置..."
PIP_CONF="$NOW_HOME/.pip/pip.conf"
restore_from_backup "$PIP_CONF"

echo -e "\n所有代理配置已清除或从备份中恢复！"你的terminal/docker/apt/git/rosdep/pip的代理
