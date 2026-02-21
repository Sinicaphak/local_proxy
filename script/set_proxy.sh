#!/bin/sh
# 这会更改你的terminal/docker/apt/git/rosdep/pip的代理
NOW_USER=${SUDO_USER:-$(id -un 1000)} 
NOW_HOME=$(getent passwd "$NOW_USER" | cut -d: -f6) 
echo $NOW_USER
echo $NOW_HOME

ensure_config_file() {
    local conf_file="$1"
    local dir
    dir=$(dirname "$conf_file")
    local cmd_prefix=""

    # 检查文件路径是否在当前用户的 HOME 目录下
    # 注意：使用 case 语句比 grep 更安全，可以避免 $HOME 中有特殊字符的问题
    case "$conf_file" in
        "$NOW_HOME"/*)
            # 在用户目录下，不需要 sudo
            cmd_prefix=""
            ;;
        *)
            # 不在用户目录下，需要 sudo
            cmd_prefix="sudo"
            ;;
    esac

    # 如果目录不存在，则创建
    if [ ! -d "$dir" ]; then
        echo "创建目录: $dir"
        $cmd_prefix mkdir -p "$dir"
        # 如果是用户目录，修正权限
        if [ -z "$cmd_prefix" ]; then
            # 如果是用户目录，mkdir 已经以用户身份创建，通常不需要 chown
            # 但为确保，可以加上 chown
            chown -R "$NOW_USER:$NOW_USER" "$dir"
        fi
    fi

    # 如果文件存在，则创建备份
    if [ -f "$conf_file" ]; then
        # 如果备份已存在，先删除旧备份
        if [ -f "$conf_file.bak" ]; then
            $cmd_prefix rm "$conf_file.bak"
        fi
        echo "备份旧文件: $conf_file -> $conf_file.bak"
        $cmd_prefix mv "$conf_file" "$conf_file.bak"
    fi
    
    # 创建一个空的新文件，以便后续写入
    echo "创建新文件: $conf_file"
    $cmd_prefix touch "$conf_file"
    # 如果是用户目录，确保文件属于当前用户
    if [ -z "$cmd_prefix" ]; then
        chown "$NOW_USER:$NOW_USER" "$conf_file" 2>/dev/null
    fi
}

# 读取本地代理配置
CONFIG_FILE="/etc/local-proxy/config.yaml"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件未找到于 $CONFIG_FILE"
    exit 1
fi

SELF_IP=$(grep "^self_ip:" "$CONFIG_FILE" | sed 's/self_ip:[[:space:]]*//; s/"//g; s/'"'"'//g')
SELF_PORT=$(grep "^self_port:" "$CONFIG_FILE" | sed 's/self_port:[[:space:]]*//; s/"//g; s/'"'"'//g')
if [ -z "$SELF_IP" ] || [ -z "$SELF_PORT" ]; then
    echo "错误: 无法从 $CONFIG_FILE 解析 IP 或端口。"
    exit 1
fi
PROXY_HTTP_URL="http://$SELF_IP:$SELF_PORT"
PROXY_HTTPS_URL="https://$SELF_IP:$SELF_PORT"
NO_PROXY="localhost,127.0.0.1,localaddress,.localdomain.com"


## 终端代理
TERMINAL_CONF="$NOW_HOME/.bashrc"
# 注释掉任何已存在的 http_proxy 和 https_proxy 设置
sed -i -E 's/^(export (http_proxy|https_proxy)=.*)/#\1/g' "$TERMINAL_CONF"
# 删除之前由本脚本添加的代理块，以防重复
sed -i '/# Local Proxy Settings Start/,/# Local Proxy Settings End/d' "$TERMINAL_CONF"

# 在文件末尾添加新的代理配置块
tee -a "$TERMINAL_CONF" > /dev/null <<EOF

# Local Proxy Settings Start
export http_proxy="$PROXY_HTTP_URL"
export https_proxy="$PROXY_HTTP_URL"
# Local Proxy Settings End
EOF

## docker 代理

DOCKER_CONF="/etc/docker/daemon.json"
sudo mkdir -p "/etc/docker"
# 检查文件是否存在,不存在则创建
if [ ! -f "$DOCKER_CONF" ]; then
    sudo touch "$DOCKER_CONF"
fi

sudo tee "$DOCKER_CONF"  > /dev/null <<EOF
{
  "proxies": {
      "http-proxy": "http://127.0.0.1:7897",
      "https-proxy": "http://127.0.0.1:7897",
      "no-proxy": "localhost,127.0.0.1"
  }
}

EOF

sudo systemctl daemon-reload
sudo systemctl restart docker
# 确保 docker 组存在
if ! grep -q "^docker:" /etc/group; then
    echo "创建 docker 用户组..."
    sudo groupadd docker
    sudo usermod -aG docker "$NOW_USER"
fi


## 修改apt代理
APT_CONF="/etc/apt/apt.conf"

ensure_config_file "$APT_CONF"


# 在文件尾部添加新的代理配置
sudo tee -a "$APT_CONF" > /dev/null <<EOF
Acquire::http::proxy "$PROXY_HTTP_URL";
Acquire::ftp::proxy "ftp://$SELF_IP:$SELF_PORT";
Acquire::https::proxy "$PROXY_HTTPS_URL";
EOF

## git 代理
git config --global http.proxy "$PROXY_HTTP_URL"
git config --global https.proxy "$PROXY_HTTP_URL"

## rosdep 代理
ROSDEP_CONF="/etc/ros/rosdep/sources.list.d/20-default.list"

ensure_config_file "$ROSDEP_CONF"

sudo tee -a "$ROSDEP_CONF" > /dev/null <<EOF
# os-specific listings first
yaml https://mirrors.tuna.tsinghua.edu.cn/github-raw/ros/rosdistro/master/rosdep/osx-homebrew.yaml osx

# generic
yaml https://mirrors.tuna.tsinghua.edu.cn/github-raw/ros/rosdistro/master/rosdep/base.yaml
yaml https://mirrors.tuna.tsinghua.edu.cn/github-raw/ros/rosdistro/master/rosdep/python.yaml
yaml https://mirrors.tuna.tsinghua.edu.cn/github-raw/ros/rosdistro/master/rosdep/ruby.yaml

# newer distributions (Groovy, Hydro, ...) must not be listed anymore, they are being fetched from the rosdistro index.yaml instead
EOF

## pip 代理
PIP_CONF="$NOW_HOME/.pip/pip.conf"

ensure_config_file "$PIP_CONF"

# 在文件尾部添加新的代理配置
tee -a "$PIP_CONF" > /dev/null <<EOF
[global]
proxy = $PROXY_HTTP_URL
EOF

