package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"go-network-mini-project/config"
)

var (
	serverAddr    *net.UDPAddr
	serverAddrMux sync.RWMutex
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

	// resolve Client 1 address
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
		n, senderAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("read UDP data failed: %v\n", err)
			continue
		}

		message := string(buffer[:n])

		// Check if message is from Server or Client
		isFromClient := senderAddr.String() == clientUDPAddr.String()

		if isFromClient {
			// Message from Client (NACK or FIN) - forward to Server
			serverAddrMux.RLock()
			currentServerAddr := serverAddr
			serverAddrMux.RUnlock()

			if currentServerAddr != nil {
				if strings.HasPrefix(message, "NACK:") || strings.TrimSpace(message) == "FIN" {
					_, err = conn.WriteToUDP(buffer[:n], currentServerAddr)
					if err != nil && !quietMode {
						fmt.Printf("[Proxy1] forward %s to Server failed: %v\n", message, err)
					} else if !quietMode {
						fmt.Printf("[Proxy1] forwarded %s from Client to Server\n", message)
					}
				}
			}
		} else {
			// Message from Server - store server address and forward to Client
			serverAddrMux.Lock()
			serverAddr = senderAddr
			serverAddrMux.Unlock()

			packetCount++

			if !quietMode {
				fmt.Printf("Proxy 1 received: %s from Server (packet #%d)\n", message, packetCount)
			}

			// 10% packet loss simulation (only for data packets, not retransmissions)
			if rng.Float64() < 0.10 && !strings.HasPrefix(message, "NACK:") {
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
}
