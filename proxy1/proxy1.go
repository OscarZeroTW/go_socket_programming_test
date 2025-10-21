package main

import (
	"fmt"
	"io"
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

	// create TCP listener (receive packets from Server)
	proxy1Addr := proxyConfig.UDPProxy1IP + ":" + proxyConfig.UDPProxy1ListenPort
	listener, err := net.Listen("tcp", proxy1Addr)
	if err != nil {
		fmt.Printf("create TCP listener failed: %v\n", err)
		return
	}
	defer listener.Close()

	clientAddr := clientConfig.ClientIP + ":" + clientConfig.Client1ListenPort
	if !quietMode {
		fmt.Printf("TCP Proxy 1 started, listening on: %s\n", proxy1Addr)
		fmt.Printf("target Client 1 address: %s\n", clientAddr)
		fmt.Println("Proxy 1 started listening and forwarding (10% packet loss simulation)...")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		// accept connection from Server
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("accept connection failed: %v\n", err)
			continue
		}

		// handle connection in goroutine
		go handleConnection(conn, clientAddr, quietMode, rng)
	}
}

func handleConnection(serverConn net.Conn, clientAddr string, quietMode bool, rng *rand.Rand) {
	defer serverConn.Close()

	// connect to Client 1
	clientConn, err := net.Dial("tcp", clientAddr)
	if err != nil {
		if !quietMode {
			fmt.Printf("connect to Client 1 failed: %v\n", err)
		}
		return
	}
	defer clientConn.Close()

	buffer := make([]byte, 1024)
	packetCount := 0
	droppedCount := 0

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
			fmt.Printf("Proxy 1 received: %s (packet #%d)\n", message, packetCount)
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
		_, err = clientConn.Write(buffer[:n])
		if err != nil {
			if !quietMode {
				fmt.Printf("forward packet %d to Client 1 failed: %v\n", packetCount, err)
			}
			break
		} else if !quietMode {
			fmt.Printf("Proxy 1 forwarded packet %d to Client 1\n", packetCount)
		}
	}
}
