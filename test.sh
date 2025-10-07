#!/bin/bash

function show_usage() {
    echo "使用方法:"
    echo "  ./test.sh server     - 啟動UDP Server"
    echo "  ./test.sh client1    - 啟動UDP Client 1"
    echo "  ./test.sh client2    - 啟動UDP Client 2"
    echo "  ./test.sh proxy1     - 啟動UDP Proxy 1"
    echo "  ./test.sh proxy2     - 啟動UDP Proxy 2"
    echo "  ./test.sh all        - 啟動所有組件"
    echo "  ./test.sh all -q     - 啟動所有組件（靜默模式，只顯示Client結果）"
}

function run_server() {
    echo "啟動UDP Server..."
    go run server/server.go
}

function run_client1() {
    echo "啟動UDP Client 1..."
    go run client1/client1.go
}

function run_client2() {
    echo "啟動UDP Client 2..."
    go run client2/client2.go
}

function run_proxy1() {
    echo "啟動UDP Proxy 1..."
    go run proxy1/proxy1.go
}

function run_proxy2() {
    echo "啟動UDP Proxy 2..."
    go run proxy2/proxy2.go
}

function run_all() {
    local quiet_mode=$1
    
    if [ "$quiet_mode" != "-q" ]; then
        echo "啟動所有組件..."
    fi
    
    # 啟動Client 1
    if [ "$quiet_mode" != "-q" ]; then
        echo "啟動UDP Client 1..."
    fi
    if [ "$quiet_mode" = "-q" ]; then
        go run client1/client1.go &
    else
        go run client1/client1.go &
    fi
    
    # 啟動Client 2
    if [ "$quiet_mode" != "-q" ]; then
        echo "啟動UDP Client 2..."
    fi
    if [ "$quiet_mode" = "-q" ]; then
        go run client2/client2.go &
    else
        go run client2/client2.go &
    fi
    
    # 等待一下讓Clients啟動
    sleep 1
    
    # 啟動Proxy 1
    if [ "$quiet_mode" != "-q" ]; then
        echo "啟動UDP Proxy 1..."
    fi
    if [ "$quiet_mode" = "-q" ]; then
        go run proxy1/proxy1.go -q > /dev/null 2>&1 &
    else
        go run proxy1/proxy1.go &
    fi
    
    # 啟動Proxy 2
    if [ "$quiet_mode" != "-q" ]; then
        echo "啟動UDP Proxy 2..."
    fi
    if [ "$quiet_mode" = "-q" ]; then
        go run proxy2/proxy2.go -q > /dev/null 2>&1 &
    else
        go run proxy2/proxy2.go &
    fi
    
    # 等待一下讓Proxies啟動
    sleep 1
    
    # 啟動Server
    if [ "$quiet_mode" != "-q" ]; then
        echo "啟動UDP Server..."
    fi
    if [ "$quiet_mode" = "-q" ]; then
        go run server/server.go -q > /dev/null 2>&1
    else
        go run server/server.go
    fi
}

# Main script logic
if [ $# -lt 1 ]; then
    show_usage
    exit 0
fi

command=$1
quiet_flag=${2:-""}

case $command in
    server)
        run_server
        ;;
    client1)
        run_client1
        ;;
    client2)
        run_client2
        ;;
    proxy1)
        run_proxy1
        ;;
    proxy2)
        run_proxy2
        ;;
    all)
        run_all "$quiet_flag"
        ;;
    *)
        echo "未知命令: $command"
        show_usage
        exit 1
        ;;
esac
