package main

import (
	"fmt"
	"net"
	"os"

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
	clientConfig := cfg.GetClientConfig()
	quietMode := len(os.Args) > 1 && os.Args[1] == "-q"

	// create UDP listener (receive packets from Server)
	proxy2Addr := proxyConfig.UDPProxy2IP + ":" + proxyConfig.UDPProxy2ListenPort
	addr, err := net.ResolveUDPAddr("udp", proxy2Addr)
	if err != nil {
		fmt.Printf("resolve UDP address failed: %v\n", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("create UDP listener failed: %v\n", err)
		return
	}
	defer conn.Close()

	clientAddr := clientConfig.ClientIP + ":" + clientConfig.Client2ListenPort
	if !quietMode {
		fmt.Printf("UDP Proxy 2 started, listening on: %s\n", proxy2Addr)
		fmt.Printf("target Client 2 address: %s\n", clientAddr)
	}

	// resolve Client 2 address
	clientUDPAddr, err := net.ResolveUDPAddr("udp", clientAddr)
	if err != nil {
		fmt.Printf("resolve Client 2 UDP address failed: %v\n", err)
		return
	}

	// receive and forward packets (normal listening)
	buffer := make([]byte, 1024)
	packetCount := 0

	if !quietMode {
		fmt.Println("Proxy 2 started normal listening and forwarding...")
	}

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("read UDP data failed: %v\n", err)
			continue
		}

		packetCount++
		message := string(buffer[:n])
		
		if !quietMode {
			fmt.Printf("Proxy 2 received: %s (%d packet)\n", 
				message, packetCount)
		}

		// 轉發到Client 2
		_, err = conn.WriteToUDP(buffer[:n], clientUDPAddr)
		if err != nil {
			if !quietMode {
				fmt.Printf("forward packet %d to Client 2 failed: %v\n", packetCount, err)
			}
		} else if !quietMode {
			fmt.Printf("Proxy 2 forwarded packet %d to Client 2\n", packetCount)
		}
	}
}
