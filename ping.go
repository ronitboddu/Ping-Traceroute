package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	arg := os.Args[1:]
	destIP := arg[0]
	var err error
	domain, err := net.LookupIP(destIP)
	if err != nil {
		log.Fatal(err)
	}
	destIP = domain[len(domain)-1].String()
	count := math.MaxInt
	wait := 1
	size := 56
	timeout := math.MaxInt
	timeout_flag := false
	if flag, i := contains(arg, "-c"); flag {
		count, err = strconv.Atoi(arg[i+1])
		if err != nil {
			log.Fatal("Cannot convert string to int")
		}
	}
	if flag, i := contains(arg, "-i"); flag {
		wait, err = strconv.Atoi(arg[i+1])
		if err != nil {
			log.Fatal("Cannot convert string to int")
		}
	}
	if flag, i := contains(arg, "-s"); flag {
		size, err = strconv.Atoi(arg[i+1])
		if err != nil {
			log.Fatal("Cannot convert string to int")
		}
	}
	if flag, i := contains(arg, "-t"); flag {
		timeout, err = strconv.Atoi(arg[i+1])
		if err != nil {
			log.Fatal("Cannot convert string to int")
		}
		timeout_flag = true
	}
	start := time.Now().Second()
	seq := 0
	for i := 0; i < count; i++ {
		time.Sleep(time.Duration(wait) * time.Second)
		timeout := timeout - (time.Now().Second() - start)
		// fmt.Println(timeout)
		if timeout <= 0 {
			fmt.Println("Timeout")
			os.Exit(0)
		}
		Ping(destIP, size, timeout, seq, timeout_flag)
		seq += 1
	}
}

func Ping(destIP string, size int, timeout int, seq int, timeout_flag bool) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatalf("Packet listen err, %s", err)
	}
	if timeout_flag {
		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}
	defer conn.Close()

	start_ms := time.Now().UnixMicro()

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 2,
			Data: make([]byte, size),
		},
	}

	wb, err := wm.Marshal(nil)
	//fmt.Println(len(wb))
	if err != nil {
		log.Fatal(err)
	}
	if _, err := conn.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(destIP)}); err != nil {
		log.Fatalf("%s", err)
	}

	rb := make([]byte, 10000)
	n, peer, err := conn.ReadFrom(rb)
	if err != nil {
		log.Fatal(err)
	}
	rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), rb[:n])

	if err != nil {
		log.Fatal(err)
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		length := rm.Body.Len(0)
		ttl, err := conn.IPv4PacketConn().TTL()
		time := float64(time.Now().UnixMicro()-start_ms) / 1000
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%d bytes from %v: icmp_seq=%d ttl=%d time=%.3f ms", length, peer.String(), seq, ttl, time)
	default:
		log.Printf("got %+v; want echo reply", rm)
	}
}

func contains(s []string, e string) (bool, int) {
	for i := 0; i < len(s); i++ {
		if s[i] == e {
			return true, i
		}
	}
	return false, -1
}
