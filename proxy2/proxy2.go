package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"sync"
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

	// create TCP listener (receive packets from Server)
	proxy2Addr := proxyConfig.UDPProxy2IP + ":" + proxyConfig.UDPProxy2ListenPort
	listener, err := net.Listen("tcp", proxy2Addr)
	if err != nil {
		fmt.Printf("create TCP listener failed: %v\n", err)
		return
	}
	defer listener.Close()

	clientAddr := clientConfig.ClientIP + ":" + clientConfig.Client2ListenPort
	if !quietMode {
		fmt.Printf("TCP Proxy 2 started, listening on: %s\n", proxy2Addr)
		fmt.Printf("target Client 2 address: %s\n", clientAddr)
		fmt.Println("Proxy 2 started listening and forwarding (5% async delay 20ms simulation)...")
	}

	for {
		// accept connection from Server
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("accept connection failed: %v\n", err)
			continue
		}

		// handle connection in goroutine
		go handleConnection(conn, clientAddr, quietMode)
	}
}

func handleConnection(serverConn net.Conn, clientAddr string, quietMode bool) {
	defer serverConn.Close()

	// connect to Client 2
	clientConn, err := net.Dial("tcp", clientAddr)
	if err != nil {
		if !quietMode {
			fmt.Printf("connect to Client 2 failed: %v\n", err)
		}
		return
	}
	defer clientConn.Close()

	buffer := make([]byte, 1024)
	packetCount := 0
	delayedCount := 0
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var wg sync.WaitGroup

	for {
		n, err := serverConn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("read TCP data failed: %v\n", err)
			}
			break
		}

		packetCount++
		message := string(buffer[:n])

		if !quietMode {
			fmt.Printf("Proxy 2 received: %s (packet #%d)\n", message, packetCount)
		}

		// Copy data to avoid buffer reuse issues
		data := make([]byte, n)
		copy(data, buffer[:n])

		// 5% delay simulation (20ms) - non-blocking
		if rng.Float64() < 0.05 {
			delayedCount++
			if !quietMode {
				fmt.Printf("Proxy 2 will DELAY packet #%d by 20ms (5%% delay simulation) - Total delayed: %d\n",
					packetCount, delayedCount)
			}
			// Use goroutine to delay and send packet asynchronously
			wg.Add(1)
			go func(data []byte, pktNum int) {
				defer wg.Done()
				time.Sleep(20 * time.Millisecond)
				_, err := clientConn.Write(data)
				if err != nil {
					if !quietMode {
						fmt.Printf("delayed packet %d to Client 2 failed: %v\n", pktNum, err)
					}
				} else if !quietMode {
					fmt.Printf("Proxy 2 forwarded DELAYED packet %d to Client 2 (after 20ms)\n", pktNum)
				}
			}(data, packetCount)
		} else {
			// forward to Client 2
			_, err = clientConn.Write(data)
			if err != nil {
				if !quietMode {
					fmt.Printf("forward packet %d to Client 2 failed: %v\n", packetCount, err)
				}
				break
			} else if !quietMode {
				fmt.Printf("Proxy 2 forwarded packet %d to Client 2 (immediate)\n", packetCount)
			}
		}
	}

	// Wait for all delayed packets to be sent
	wg.Wait()
}
