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
	packetCache    = make(map[int]PacketBuffer) // cache for retransmission
	cacheMutex     sync.RWMutex
	retransmitChan = make(chan RetransmitRequest, 100)
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

	// create UDP connection to Proxy 1
	proxy1Addr := proxyConfig.UDPProxy1IP + ":" + proxyConfig.UDPProxy1ListenPort
	proxy1UDPAddr, err := net.ResolveUDPAddr("udp", proxy1Addr)
	if err != nil {
		fmt.Printf("resolve Proxy 1 UDP address failed: %v\n", err)
		return
	}
	conn1, err := net.DialUDP("udp", nil, proxy1UDPAddr)
	if err != nil {
		fmt.Printf("create UDP connection to Proxy 1 failed: %v\n", err)
		return
	}
	defer conn1.Close()

	// create UDP connection to Proxy 2
	proxy2Addr := proxyConfig.UDPProxy2IP + ":" + proxyConfig.UDPProxy2ListenPort
	proxy2UDPAddr, err := net.ResolveUDPAddr("udp", proxy2Addr)
	if err != nil {
		fmt.Printf("resolve Proxy 2 UDP address failed: %v\n", err)
		return
	}
	conn2, err := net.DialUDP("udp", nil, proxy2UDPAddr)
	if err != nil {
		fmt.Printf("create UDP connection to Proxy 2 failed: %v\n", err)
		return
	}
	defer conn2.Close()

	// start NACK listener for conn1 and conn2
	go nackListener(conn1, "Proxy1", quietMode)
	go nackListener(conn2, "Proxy2", quietMode)

	// start retransmit handler
	go retransmitHandler(quietMode)

	if !quietMode {
		fmt.Printf("UDP Server started, sending to Proxy 1: %s and Proxy 2: %s\n", proxy1Addr, proxy2Addr)
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
		_, err := conn1.Write([]byte(message))
		if err != nil {
			if !quietMode {
				fmt.Printf("send Packet %d to Proxy 1 failed: %v\n", i, err)
			}
		} else if !quietMode && i%1000 == 0 {
			fmt.Printf("sent: Packet %d to Proxy 1\n", i)
		}

		// send to Proxy 2
		_, err = conn2.Write([]byte(message))
		if err != nil {
			if !quietMode {
				fmt.Printf("send Packet %d to Proxy 2 failed: %v\n", i, err)
			}
		} else if !quietMode && i%1000 == 0 {
			fmt.Printf("sent: Packet %d to Proxy 2\n", i)
		}

		// wait 1ms (adjust for 10000 packets)
		time.Sleep(1 * time.Millisecond)
	}

	if !quietMode {
		fmt.Println("all packets sent, waiting for retransmit requests...")
	}

	// keep server running to handle retransmissions
	time.Sleep(30 * time.Second)

	if !quietMode {
		fmt.Println("server shutting down")
	}
}

func nackListener(conn *net.UDPConn, proxyName string, quietMode bool) {
	buffer := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if !quietMode {
				fmt.Printf("[%s] read NACK failed: %v\n", proxyName, err)
			}
			continue
		}

		message := string(buffer[:n])
		// NACK format: "NACK:123"
		if strings.HasPrefix(message, "NACK:") {
			var seqNum int
			_, err := fmt.Sscanf(message, "NACK:%d", &seqNum)
			if err != nil {
				if !quietMode {
					fmt.Printf("[%s] parse NACK failed: %v\n", proxyName, err)
				}
				continue
			}

			if !quietMode {
				fmt.Printf("[%s] received NACK for packet %d from %s\n", proxyName, seqNum, addr)
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
