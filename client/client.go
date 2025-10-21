package main

import (
	"fmt"
	"io"
	"net"
	"strings"
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

	clientConfig := cfg.GetClientConfig()

	// create TCP listener
	clientAddr := clientConfig.ClientIP + ":" + clientConfig.ClientListenPort
	listener, err := net.Listen("tcp", clientAddr)
	if err != nil {
		fmt.Printf("create TCP listener failed: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("TCP Client started, listening on: %s\n", clientAddr)
	fmt.Println("waiting for packets (normal listening)...")

	for {
		// accept connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("accept connection failed: %v\n", err)
			continue
		}

		// handle connection in goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Get remote address
	remoteAddr := conn.RemoteAddr().String()

	// receive packets (normal listening)
	buffer := make([]byte, 1024)
	packetCount := 0

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("read TCP data failed: %v\n", err)
			}
			break
		}

		packetCount++
		receiveTime := time.Now() // Record receive time
		message := string(buffer[:n])

		// parse packet: format is "Packet <number>|<timestamp>"
		parts := strings.SplitN(message, "|", 2)
		if len(parts) == 2 {
			packetInfo := parts[0]
			timestampStr := parts[1]

			// parse timestamp
			sendTime, err := time.Parse(time.RFC3339Nano, timestampStr)
			if err != nil {
				fmt.Printf("parse timestamp failed: %v\n", err)
				fmt.Printf("received from %s: %s (%d packet)\n",
					remoteAddr, packetInfo, packetCount)
			} else {
				// calculate latency
				latency := receiveTime.Sub(sendTime)
				fmt.Printf("received from %s: %s (%d packet)\n",
					remoteAddr, packetInfo, packetCount)
				fmt.Printf("  Transmit Time: %s\n", sendTime.Format(time.RFC3339Nano))
				fmt.Printf("  Receive Time: %s\n", receiveTime.Format(time.RFC3339Nano))
				fmt.Printf("  Latency: %v\n", latency.Round(time.Microsecond))
			}
		} else {
			// old format or format error
			fmt.Printf("received from %s: %s (%d packet)\n",
				remoteAddr, message, packetCount)
		}
	}
}
