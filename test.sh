#!/bin/bash

function show_usage() {
    echo "usage:"
    echo "  ./test.sh server     - Start UDP Server"
    echo "  ./test.sh client1    - Start UDP Client 1"
    echo "  ./test.sh client2    - Start UDP Client 2"
    echo "  ./test.sh proxy1     - Start UDP Proxy 1"
    echo "  ./test.sh proxy2     - Start UDP Proxy 2"
    echo "  ./test.sh all        - Start all components"
    echo "  ./test.sh all -q     - Start all components (quiet mode, only show Client results)"
}

function run_server() {
    echo "Start UDP Server..."
    go run server/server.go
}

function run_client1() {
    echo "Start UDP Client 1..."
    go run client1/client1.go
}

function run_client2() {
    echo "Start UDP Client 2..."
    go run client2/client2.go
}

function run_proxy1() {
    echo "Start UDP Proxy 1..."
    go run proxy1/proxy1.go
}

function run_proxy2() {
    echo "Start UDP Proxy 2..."
    go run proxy2/proxy2.go
}

function run_all() {
    local quiet_mode=$1
    
    if [ "$quiet_mode" != "-q" ]; then
        echo "start all components..."
    fi
    
    # Start Client 1
    if [ "$quiet_mode" != "-q" ]; then
        echo "Start UDP Client 1..."
    fi
    if [ "$quiet_mode" = "-q" ]; then
        go run client1/client1.go &
    else
        go run client1/client1.go &
    fi
    
    # Start Client 2
    if [ "$quiet_mode" != "-q" ]; then
        echo "Start UDP Client 2..."
    fi
    if [ "$quiet_mode" = "-q" ]; then
        go run client2/client2.go &
    else
        go run client2/client2.go &
    fi
    
    # Wait a moment to let Clients start
    sleep 1
    
    # Start Proxy 1
    if [ "$quiet_mode" != "-q" ]; then
        echo "Start UDP Proxy 1..."
    fi
    if [ "$quiet_mode" = "-q" ]; then
        go run proxy1/proxy1.go -q > /dev/null 2>&1 &
    else
        go run proxy1/proxy1.go &
    fi
    
    # Start Proxy 2
    if [ "$quiet_mode" != "-q" ]; then
        echo "Start UDP Proxy 2..."
    fi
    if [ "$quiet_mode" = "-q" ]; then
        go run proxy2/proxy2.go -q > /dev/null 2>&1 &
    else
        go run proxy2/proxy2.go &
    fi
    
    # Wait a moment to let Proxies start
    sleep 1

    # Start Server
    if [ "$quiet_mode" != "-q" ]; then
        echo "Start UDP Server..."
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
        echo "Unknown command: $command"
        show_usage
        exit 1
        ;;
esac
