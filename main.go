package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	protocolICMP = 1
)

func ping(host string) {
	dst, err := net.ResolveIPAddr("ip", os.Args[1])
	if err != nil {
		fmt.Printf("Failed to resolve hostname: %v\n", err)
		os.Exit(1)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		fmt.Printf("Failed to listen for ICMP packets: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: make([]byte, 8),
		},
	}
	packet, err := msg.Marshal(nil)
	if err != nil {
		fmt.Printf("Failed to create ICMP packet: %v\n", err)
		os.Exit(1)
	}

	start := time.Now()
	n, err := conn.WriteTo(packet, dst)
	if err != nil {
		fmt.Printf("Failed to send ICMP packet: %v\n", err)
		os.Exit(1)
	} else if n != len(packet) {
		fmt.Printf("Failed to send the entire ICMP packet\n")
		os.Exit(1)
	}

	reply := make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		fmt.Printf("Failed to set read deadline: %v\n", err)
		os.Exit(1)
	}

	n, peer, err := conn.ReadFrom(reply)
	if err != nil {
		fmt.Printf("Failed to receive ICMP packet: %v\n", err)
		os.Exit(1)
	}
	duration := time.Since(start)

	rm, err := icmp.ParseMessage(protocolICMP, reply[:n])
	if err != nil {
		fmt.Printf("Failed to parse ICMP message: %v\n", err)
		os.Exit(1)
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		replyEcho := rm.Body.(*icmp.Echo)
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n", n, peer, replyEcho.Seq, duration)
	default:
		fmt.Printf("Unexpected ICMP message: %+v\n", rm)
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s <hostname> <option>\n", os.Args[0])
		os.Exit(1)
	}

	if len(os.Args) == 2 {
		ping(os.Args[1])
	}

	if len(os.Args) == 3 && os.Args[2] == "-c" {
		for os.Args[2] == "-c" {
			ping(os.Args[1])
			time.Sleep(1 * time.Second)
		}
	}
}
