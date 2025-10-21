package main

import (
	"fmt"
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

	// create UDP listener
	clientAddr := clientConfig.ClientIP + ":" + clientConfig.Client2ListenPort
	addr, err := net.ResolveUDPAddr("udp", clientAddr)
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

	fmt.Printf("UDP Client 2 started, listening on: %s\n", clientAddr)
	fmt.Println("waiting for packets (normal listening)...")

	// receive packets (normal listening)
	buffer := make([]byte, 1024)
	packetCount := 0

	for {
		n, senderAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("read UDP data failed: %v\n", err)
			continue
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
				fmt.Printf("[Client 2] parse timestamp failed: %v\n", err)
				fmt.Printf("[Client 2] received from %s: %s (%d packet)\n",
					senderAddr, packetInfo, packetCount)
				fmt.Printf("  Receive Time: %s\n", receiveTime.Format(time.RFC3339Nano))
			} else {
				// calculate latency
				latency := receiveTime.Sub(sendTime)
				fmt.Printf("[Client 2] received from %s: %s (%d packet)\n",
					senderAddr, packetInfo, packetCount)
				fmt.Printf("  Transmit Time: %s\n", sendTime.Format(time.RFC3339Nano))
				fmt.Printf("  Receive Time: %s\n", receiveTime.Format(time.RFC3339Nano))
				fmt.Printf("  Latency: %v\n", latency.Round(time.Microsecond))
			}
		} else {
			// old format or format error
			fmt.Printf("[Client 2] received from %s: %s (%d packet)\n",
				senderAddr, message, packetCount)
			fmt.Printf("  Receive Time: %s\n", receiveTime.Format(time.RFC3339Nano))
		}
	}
}
