FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    build-essential \
    debhelper=13.6ubuntu1 \
    devscripts \
    dh-autoreconf \
    git \
    wget \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# go >= 1.21 
RUN wget https://go.dev/dl/go1.25.7.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.25.7.linux-amd64.tar.gz && \
    rm go1.25.7.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"

RUN useradd -m -s /bin/bash tester && \
    usermod -aG sudo tester && \
    echo "tester:tester" | chpasswd && \
    echo "tester ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
    RUN chown -R tester:tester /home/tester
USER tester
WORKDIR /home/tester/local_proxy
# RUN dpkg-buildpackage -us -uc -b
# dch -v 0.1.8-1 "修正配置文件权限问题"
WORKDIR /build