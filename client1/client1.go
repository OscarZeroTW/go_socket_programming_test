package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"go-network-mini-project/config"
)

type ReorderBuffer struct {
	mu               sync.Mutex
	buffer           map[int]PacketData
	expectedSeqNum   int
	receivedCount    int
	processedCount   int
	lostPackets      map[int]bool
	nackSent         map[int]bool
	nackLastSentTime map[int]time.Time
	totalPackets     int
	completed        bool
}

type PacketData struct {
	seqNum    int
	message   string
	timestamp time.Time
	recvTime  time.Time
}

func NewReorderBuffer() *ReorderBuffer {
	return &ReorderBuffer{
		buffer:           make(map[int]PacketData),
		expectedSeqNum:   1,
		lostPackets:      make(map[int]bool),
		nackSent:         make(map[int]bool),
		nackLastSentTime: make(map[int]time.Time),
		totalPackets:     10000,
		completed:        false,
	}
}

func (rb *ReorderBuffer) processPacket(seqNum int, message string, timestamp time.Time, recvTime time.Time, conn *net.UDPConn, senderAddr *net.UDPAddr) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.receivedCount++

	if seqNum == rb.expectedSeqNum {
		// received expected packet, process it
		rb.processAndPrint(seqNum, message, timestamp, recvTime)
		rb.expectedSeqNum++
		rb.processedCount++

		// try to process buffered packets
		for {
			if pkt, exists := rb.buffer[rb.expectedSeqNum]; exists {
				rb.processAndPrint(pkt.seqNum, pkt.message, pkt.timestamp, pkt.recvTime)
				delete(rb.buffer, rb.expectedSeqNum)
				rb.expectedSeqNum++
				rb.processedCount++
			} else {
				break
			}
		}

		// check if all packets received
		rb.checkCompletion(conn, senderAddr)
	} else if seqNum > rb.expectedSeqNum {
		// received out-of-order packet, buffer it
		rb.buffer[seqNum] = PacketData{
			seqNum:    seqNum,
			message:   message,
			timestamp: timestamp,
			recvTime:  recvTime,
		}
		fmt.Printf("[Client 1] Out-of-order: received SEQ %d, expected %d (buffered)\n", seqNum, rb.expectedSeqNum)

		// send NACK for missing packets (only if not already buffered and not already sent NACK)
		for i := rb.expectedSeqNum; i < seqNum; i++ {
			// Check if packet is not in buffer and NACK not sent yet
			_, alreadyBuffered := rb.buffer[i]
			if !alreadyBuffered && !rb.nackSent[i] {
				rb.sendNACK(i, conn, senderAddr)
				rb.nackSent[i] = true
				rb.lostPackets[i] = true
			}
		}
	} else {
		// received duplicate or old packet
		fmt.Printf("[Client 1] Duplicate/Old packet: received SEQ %d, expected %d (ignored)\n", seqNum, rb.expectedSeqNum)
	}
}

func (rb *ReorderBuffer) processAndPrint(seqNum int, message string, timestamp time.Time, recvTime time.Time) {
	latency := recvTime.Sub(timestamp)
	if seqNum%1000 == 0 || seqNum <= 10 {
		fmt.Printf("[Client 1] Processed SEQ %d\n", seqNum)
		fmt.Printf("  Transmit Time: %s\n", timestamp.Format(time.RFC3339Nano))
		fmt.Printf("  Receive Time: %s\n", recvTime.Format(time.RFC3339Nano))
		fmt.Printf("  Latency: %v\n", latency.Round(time.Microsecond))
	}
}

func (rb *ReorderBuffer) sendNACK(seqNum int, conn *net.UDPConn, senderAddr *net.UDPAddr) {
	nackMsg := fmt.Sprintf("NACK:%d", seqNum)
	_, err := conn.WriteToUDP([]byte(nackMsg), senderAddr)
	if err != nil {
		fmt.Printf("[Client 1] send NACK for SEQ %d failed: %v\n", seqNum, err)
	} else {
		rb.nackLastSentTime[seqNum] = time.Now()
		fmt.Printf("[Client 1] sent NACK for missing SEQ %d\n", seqNum)
	}
}

func (rb *ReorderBuffer) checkCompletion(conn *net.UDPConn, senderAddr *net.UDPAddr) {
	if !rb.completed && rb.processedCount >= rb.totalPackets {
		rb.completed = true
		fmt.Printf("\n[Client 1] === ALL PACKETS RECEIVED ===\n")
		rb.printStats()

		// send FIN to server
		finMsg := "FIN"
		_, err := conn.WriteToUDP([]byte(finMsg), senderAddr)
		if err != nil {
			fmt.Printf("[Client 1] send FIN failed: %v\n", err)
		} else {
			fmt.Printf("[Client 1] sent FIN to server\n")
		}
	}
}

func (rb *ReorderBuffer) retryNACKs(conn *net.UDPConn, senderAddr *net.UDPAddr) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.completed {
		return
	}

	now := time.Now()
	retryInterval := 500 * time.Millisecond

	// Retry NACK for missing packets between expectedSeqNum and first buffered packet
	for i := rb.expectedSeqNum; i < rb.expectedSeqNum+100; i++ {
		_, inBuffer := rb.buffer[i]
		if !inBuffer && rb.nackSent[i] {
			lastSent, exists := rb.nackLastSentTime[i]
			if !exists || now.Sub(lastSent) > retryInterval {
				// Retry NACK
				nackMsg := fmt.Sprintf("NACK:%d", i)
				conn.WriteToUDP([]byte(nackMsg), senderAddr)
				rb.nackLastSentTime[i] = now
			}
		}
		// Stop if we find a buffered packet (packets beyond might not be lost yet)
		if i > rb.expectedSeqNum+10 && len(rb.buffer) > 0 {
			break
		}
	}
}

func (rb *ReorderBuffer) printStats() {
	fmt.Printf("\n[Client 1] === Statistics ===\n")
	fmt.Printf("  Total Received: %d\n", rb.receivedCount)
	fmt.Printf("  Total Processed: %d\n", rb.processedCount)
	fmt.Printf("  Buffered Packets: %d\n", len(rb.buffer))
	fmt.Printf("  Lost Packets Detected: %d\n", len(rb.lostPackets))
	fmt.Printf("  Expected Next: %d\n", rb.expectedSeqNum)
}

func main() {
	// load config
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("config load failed: %v\n", err)
		return
	}

	clientConfig := cfg.GetClientConfig()

	// create UDP listener
	clientAddr := clientConfig.ClientIP + ":" + clientConfig.Client1ListenPort
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

	fmt.Printf("UDP Client 1 started, listening on: %s\n", clientAddr)
	fmt.Println("waiting for packets with reordering and loss recovery...")

	reorderBuf := NewReorderBuffer()
	buffer := make([]byte, 1024)
	var lastSenderAddr *net.UDPAddr

	// periodically print stats
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			reorderBuf.mu.Lock()
			reorderBuf.printStats()
			reorderBuf.mu.Unlock()
		}
	}()

	// periodically retry NACKs for missing packets
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			if lastSenderAddr != nil {
				reorderBuf.retryNACKs(conn, lastSenderAddr)
			}
		}
	}()

	for {
		n, senderAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("read UDP data failed: %v\n", err)
			continue
		}

		lastSenderAddr = senderAddr
		recvTime := time.Now()
		message := string(buffer[:n])

		// parse packet: format is "SEQ:<number>|<timestamp>"
		parts := strings.SplitN(message, "|", 2)
		if len(parts) == 2 && strings.HasPrefix(parts[0], "SEQ:") {
			var seqNum int
			_, err := fmt.Sscanf(parts[0], "SEQ:%d", &seqNum)
			if err != nil {
				fmt.Printf("[Client 1] parse sequence number failed: %v\n", err)
				continue
			}

			timestampStr := parts[1]
			sendTime, err := time.Parse(time.RFC3339Nano, timestampStr)
			if err != nil {
				fmt.Printf("[Client 1] parse timestamp failed: %v\n", err)
				continue
			}

			reorderBuf.processPacket(seqNum, message, sendTime, recvTime, conn, lastSenderAddr)
		} else {
			fmt.Printf("[Client 1] unknown packet format: %s\n", message)
		}
	}
}
