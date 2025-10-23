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
	mu             sync.Mutex
	buffer         map[int]PacketData
	expectedSeqNum int
	receivedCount  int
	processedCount int
	lostPackets    map[int]bool
	nackSent       map[int]bool
}

type PacketData struct {
	seqNum    int
	message   string
	timestamp time.Time
	recvTime  time.Time
}

func NewReorderBuffer() *ReorderBuffer {
	return &ReorderBuffer{
		buffer:         make(map[int]PacketData),
		expectedSeqNum: 1,
		lostPackets:    make(map[int]bool),
		nackSent:       make(map[int]bool),
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
	} else if seqNum > rb.expectedSeqNum {
		// received out-of-order packet, buffer it
		rb.buffer[seqNum] = PacketData{
			seqNum:    seqNum,
			message:   message,
			timestamp: timestamp,
			recvTime:  recvTime,
		}
		fmt.Printf("[Client 2] Out-of-order: received SEQ %d, expected %d (buffered)\n", seqNum, rb.expectedSeqNum)

		// send NACK for missing packets
		for i := rb.expectedSeqNum; i < seqNum; i++ {
			if !rb.nackSent[i] {
				rb.sendNACK(i, conn, senderAddr)
				rb.nackSent[i] = true
				rb.lostPackets[i] = true
			}
		}
	} else {
		// received duplicate or old packet
		fmt.Printf("[Client 2] Duplicate/Old packet: received SEQ %d, expected %d (ignored)\n", seqNum, rb.expectedSeqNum)
	}
}

func (rb *ReorderBuffer) processAndPrint(seqNum int, message string, timestamp time.Time, recvTime time.Time) {
	latency := recvTime.Sub(timestamp)
	if seqNum%1000 == 0 || seqNum <= 10 {
		fmt.Printf("[Client 2] Processed SEQ %d\n", seqNum)
		fmt.Printf("  Transmit Time: %s\n", timestamp.Format(time.RFC3339Nano))
		fmt.Printf("  Receive Time: %s\n", recvTime.Format(time.RFC3339Nano))
		fmt.Printf("  Latency: %v\n", latency.Round(time.Microsecond))
	}
}

func (rb *ReorderBuffer) sendNACK(seqNum int, conn *net.UDPConn, senderAddr *net.UDPAddr) {
	nackMsg := fmt.Sprintf("NACK:%d", seqNum)
	_, err := conn.WriteToUDP([]byte(nackMsg), senderAddr)
	if err != nil {
		fmt.Printf("[Client 2] send NACK for SEQ %d failed: %v\n", seqNum, err)
	} else {
		fmt.Printf("[Client 2] sent NACK for missing SEQ %d\n", seqNum)
	}
}

func (rb *ReorderBuffer) printStats() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	fmt.Printf("\n[Client 2] === Statistics ===\n")
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
	fmt.Println("waiting for packets with reordering and loss recovery...")

	reorderBuf := NewReorderBuffer()
	buffer := make([]byte, 1024)

	// periodically print stats
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			reorderBuf.printStats()
		}
	}()

	var lastSenderAddr *net.UDPAddr

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
				fmt.Printf("[Client 2] parse sequence number failed: %v\n", err)
				continue
			}

			timestampStr := parts[1]
			sendTime, err := time.Parse(time.RFC3339Nano, timestampStr)
			if err != nil {
				fmt.Printf("[Client 2] parse timestamp failed: %v\n", err)
				continue
			}

			reorderBuf.processPacket(seqNum, message, sendTime, recvTime, conn, lastSenderAddr)
		} else {
			fmt.Printf("[Client 2] unknown packet format: %s\n", message)
		}
	}
}
