package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"go-network-mini-project/config"
)

type PacketBuffer struct {
	seqNum    int
	message   string
	timestamp time.Time
}

var (
	packetCache      = make(map[int]PacketBuffer) // cache for retransmission
	cacheMutex       sync.RWMutex
	retransmitChan   = make(chan RetransmitRequest, 100)
	clientsCompleted = make(map[string]bool) // track which clients have finished
	clientsMutex     sync.Mutex
)

type RetransmitRequest struct {
	seqNum     int
	clientAddr *net.UDPAddr
	conn       *net.UDPConn
}

func main() {
	// load config
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("config load failed: %v\n", err)
		return
	}

	proxyConfig := cfg.GetProxyConfig()
	quietMode := len(os.Args) > 1 && os.Args[1] == "-q"

	// create UDP listener (not dial, so we can use WriteToUDP)
	serverAddr := "0.0.0.0:0" // bind to any available port
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		fmt.Printf("resolve server UDP address failed: %v\n", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("create UDP listener failed: %v\n", err)
		return
	}
	defer conn.Close()

	// resolve Proxy addresses
	proxy1Addr := proxyConfig.UDPProxy1IP + ":" + proxyConfig.UDPProxy1ListenPort
	proxy1UDPAddr, err := net.ResolveUDPAddr("udp", proxy1Addr)
	if err != nil {
		fmt.Printf("resolve Proxy 1 UDP address failed: %v\n", err)
		return
	}

	proxy2Addr := proxyConfig.UDPProxy2IP + ":" + proxyConfig.UDPProxy2ListenPort
	proxy2UDPAddr, err := net.ResolveUDPAddr("udp", proxy2Addr)
	if err != nil {
		fmt.Printf("resolve Proxy 2 UDP address failed: %v\n", err)
		return
	}

	// start NACK listener
	go nackListener(conn, quietMode)

	// start retransmit handler
	go retransmitHandler(quietMode)

	if !quietMode {
		fmt.Printf("UDP Server started on %s, sending to Proxy 1: %s and Proxy 2: %s\n",
			conn.LocalAddr().String(), proxy1Addr, proxy2Addr)
	}

	// send 10000 packets
	for i := 1; i <= 10000; i++ {
		// add timestamp (RFC3339Nano format) to packet content
		timestamp := time.Now().Format(time.RFC3339Nano)
		message := fmt.Sprintf("SEQ:%d|%s", i, timestamp)

		// cache packet for potential retransmission
		cacheMutex.Lock()
		packetCache[i] = PacketBuffer{
			seqNum:    i,
			message:   message,
			timestamp: time.Now(),
		}
		cacheMutex.Unlock()

		// send to Proxy 1
		_, err := conn.WriteToUDP([]byte(message), proxy1UDPAddr)
		if err != nil {
			if !quietMode {
				fmt.Printf("send Packet %d to Proxy 1 failed: %v\n", i, err)
			}
		} else if !quietMode && i%1000 == 0 {
			fmt.Printf("sent: Packet %d to Proxy 1\n", i)
		}

		// send to Proxy 2
		_, err = conn.WriteToUDP([]byte(message), proxy2UDPAddr)
		if err != nil {
			if !quietMode {
				fmt.Printf("send Packet %d to Proxy 2 failed: %v\n", i, err)
			}
		} else if !quietMode && i%1000 == 0 {
			fmt.Printf("sent: Packet %d to Proxy 2\n", i)
		}

		// wait 10ms (adjust for 10000 packets)
		time.Sleep(10 * time.Millisecond)
	}

	if !quietMode {
		fmt.Println("all packets sent, waiting for retransmit requests...")
	}

	// wait for both clients to finish or timeout after 60 seconds
	timeout := time.After(60 * time.Second)
	checkTicker := time.NewTicker(2 * time.Second)
	defer checkTicker.Stop()

	for {
		select {
		case <-timeout:
			if !quietMode {
				fmt.Println("timeout reached, shutting down server")
			}
			return
		case <-checkTicker.C:
			clientsMutex.Lock()
			completedCount := len(clientsCompleted)
			clientsMutex.Unlock()

			if !quietMode {
				fmt.Printf("clients completed: %d/2\n", completedCount)
			}

			if completedCount >= 2 {
				if !quietMode {
					fmt.Println("all clients completed, shutting down server")
				}
				time.Sleep(1 * time.Second) // give time for final messages
				return
			}
		}
	}
}

func nackListener(conn *net.UDPConn, quietMode bool) {
	buffer := make([]byte, 1024)
	for {
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// timeout is expected, continue
				continue
			}
			// connection closed or other error, exit gracefully
			return
		}

		message := string(buffer[:n])

		// Check for FIN message: "FIN"
		if strings.TrimSpace(message) == "FIN" {
			clientsMutex.Lock()
			clientKey := addr.String()
			if !clientsCompleted[clientKey] {
				clientsCompleted[clientKey] = true
				if !quietMode {
					fmt.Printf("received FIN from %s\n", addr)
				}
			}
			clientsMutex.Unlock()
			continue
		}

		// NACK format: "NACK:123"
		if strings.HasPrefix(message, "NACK:") {
			var seqNum int
			_, err := fmt.Sscanf(message, "NACK:%d", &seqNum)
			if err != nil {
				if !quietMode {
					fmt.Printf("parse NACK failed: %v\n", err)
				}
				continue
			}

			if !quietMode {
				fmt.Printf("received NACK for packet %d from %s\n", seqNum, addr)
			}

			retransmitChan <- RetransmitRequest{
				seqNum:     seqNum,
				clientAddr: addr,
				conn:       conn,
			}
		}
	}
}

func retransmitHandler(quietMode bool) {
	for req := range retransmitChan {
		cacheMutex.RLock()
		packet, exists := packetCache[req.seqNum]
		cacheMutex.RUnlock()

		if exists {
			_, err := req.conn.WriteToUDP([]byte(packet.message), req.clientAddr)
			if err != nil {
				if !quietMode {
					fmt.Printf("retransmit packet %d failed: %v\n", req.seqNum, err)
				}
			} else if !quietMode {
				fmt.Printf("retransmitted packet %d to %s\n", req.seqNum, req.clientAddr)
			}
		} else {
			if !quietMode {
				fmt.Printf("packet %d not found in cache for retransmission\n", req.seqNum)
			}
		}
	}
}
