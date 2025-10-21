package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"go-network-mini-project/config"
)

func main() {
	// load config
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("config load failed: %v\n", err)
		return
	}

	proxyConfig := cfg.GetProxyConfig()
	quietMode := len(os.Args) > 1 && os.Args[1] == "-q"

	// create TCP connection to Proxy 1
	proxy1Addr := proxyConfig.UDPProxy1IP + ":" + proxyConfig.UDPProxy1ListenPort
	conn1, err := net.Dial("tcp", proxy1Addr)
	if err != nil {
		fmt.Printf("create TCP connection to Proxy 1 failed: %v\n", err)
		return
	}
	defer conn1.Close()

	// create TCP connection to Proxy 2
	proxy2Addr := proxyConfig.UDPProxy2IP + ":" + proxyConfig.UDPProxy2ListenPort
	conn2, err := net.Dial("tcp", proxy2Addr)
	if err != nil {
		fmt.Printf("create TCP connection to Proxy 2 failed: %v\n", err)
		return
	}
	defer conn2.Close()

	if !quietMode {
		fmt.Printf("TCP Server started, sending to Proxy 1: %s and Proxy 2: %s\n", proxy1Addr, proxy2Addr)
	}

	// send 100 packets, every 100ms
	for i := 1; i <= 100; i++ {
		// add timestamp (RFC3339Nano format) to packet content
		timestamp := time.Now().Format(time.RFC3339Nano)
		message := fmt.Sprintf("Packet %d|%s", i, timestamp)

		// send to Proxy 1
		_, err := conn1.Write([]byte(message))
		if err != nil {
			if !quietMode {
				fmt.Printf("send Packet %d to Proxy 1 failed: %v\n", i, err)
			}
		} else if !quietMode {
			fmt.Printf("sent: Packet %d to Proxy 1\n", i)
		}

		// send to Proxy 2
		_, err = conn2.Write([]byte(message))
		if err != nil {
			if !quietMode {
				fmt.Printf("send Packet %d to Proxy 2 failed: %v\n", i, err)
			}
		} else if !quietMode {
			fmt.Printf("sent: Packet %d to Proxy 2\n", i)
		}

		// wait 10ms
		time.Sleep(10 * time.Millisecond)
	}

	if !quietMode {
		fmt.Println("all packets sent")
	}
}
