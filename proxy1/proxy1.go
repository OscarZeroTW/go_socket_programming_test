package main

import (
	"fmt"
	"math/rand"
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
	clientConfig := cfg.GetClientConfig()
	quietMode := len(os.Args) > 1 && os.Args[1] == "-q"

	// create UDP listener (receive packets from Server)
	proxy1Addr := proxyConfig.UDPProxy1IP + ":" + proxyConfig.UDPProxy1ListenPort
	addr, err := net.ResolveUDPAddr("udp", proxy1Addr)
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

	clientAddr := clientConfig.ClientIP + ":" + clientConfig.Client1ListenPort
	if !quietMode {
		fmt.Printf("UDP Proxy 1 started, listening on: %s\n", proxy1Addr)
		fmt.Printf("target Client 1 address: %s\n", clientAddr)
	}

	// resolve Client 2 address
	clientUDPAddr, err := net.ResolveUDPAddr("udp", clientAddr)
	if err != nil {
		fmt.Printf("resolve Client 1 UDP address failed: %v\n", err)
		return
	}

	// receive and forward packets (with 10% packet loss simulation)
	buffer := make([]byte, 1024)
	packetCount := 0
	droppedCount := 0
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	if !quietMode {
		fmt.Println("Proxy 1 started listening and forwarding (10% packet loss simulation)...")
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
			fmt.Printf("Proxy 1 received: %s (packet #%d)\n",
				message, packetCount)
		}

		// 10% packet loss simulation
		if rng.Float64() < 0.10 {
			droppedCount++
			if !quietMode {
				fmt.Printf("Proxy 1 DROPPED packet #%d (10%% loss simulation) - Total dropped: %d\n",
					packetCount, droppedCount)
			}
			continue
		}

		// forward to Client 1
		_, err = conn.WriteToUDP(buffer[:n], clientUDPAddr)
		if err != nil {
			if !quietMode {
				fmt.Printf("forward packet %d to Client 1 failed: %v\n", packetCount, err)
			}
		} else if !quietMode {
			fmt.Printf("Proxy 1 forwarded packet %d to Client 1\n", packetCount)
		}
	}
}
