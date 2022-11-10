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

const destIP = "8.8.8.8"

func main() {
	arg := os.Args[1:]
	var err error
	count := math.MaxInt
	wait := 1
	size := 56
	timeout := math.MaxInt
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
	}
	start := time.Now()
	for i := 0; i < count; i++ {
		time.Sleep(time.Duration(wait) * time.Second)
		duration := time.Since(start)
		if duration.Seconds() >= float64(timeout) {
			fmt.Println("Timeout")
			os.Exit(0)
		}
		Ping(destIP, size)
	}
}

func Ping(destIP string, size int) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatalf("Packet listen err, %s", err)
	}
	defer conn.Close()

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
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

	rb := make([]byte, 1500)
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
		log.Printf("got reflection from %v", peer)
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
