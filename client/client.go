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
	clientAddr := clientConfig.ClientIP + ":" + clientConfig.ClientListenPort
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

	fmt.Printf("UDP Client started, listening on: %s\n", clientAddr)
	fmt.Println("waiting for packets (normal listening)...")

	// receive packets
	buffer := make([]byte, 1024)
	packetCount := 0

	for {
		n, senderAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("read UDP data failed: %v\n", err)
			continue
		}

		packetCount++
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
					senderAddr, packetInfo, packetCount)
			} else {
				// calculate latency
				latency := time.Since(sendTime)
				fmt.Printf("received from %s: %s (%d packet, latency: %v)\n", 
					senderAddr, packetInfo, packetCount, latency.Round(time.Microsecond))
			}
		} else {
			// old format or format error
			fmt.Printf("received from %s: %s (%d packet)\n", 
				senderAddr, message, packetCount)
		}
	}
}
