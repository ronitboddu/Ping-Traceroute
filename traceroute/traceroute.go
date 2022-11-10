package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	arg := os.Args[1:]
	destIP := arg[0]
	var n_flag bool
	// size := 56
	// timeout := math.MaxInt
	if flag, _ := contains(arg, "-n"); flag {
		n_flag = true
	}
	ttl := 1

	for {
		time.Sleep(1 * time.Second)
		Ping(destIP, ttl, n_flag)
		ttl += 1
	}
}

func Ping(destIP string, ttl int, n_flag bool) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	conn.IPv4PacketConn().SetTTL(ttl)
	if err != nil {
		log.Fatalf("Packet listen err, %s", err)
	}
	defer conn.Close()

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: make([]byte, 56),
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
	var domain []string
	if !n_flag {
		domain, err = net.LookupAddr(peer.String())
		if err != nil {
			domain = []string{}
		}
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		// log.Printf("got reflection from %v", peer)
		if len(domain) > 0 {
			fmt.Print(domain[0] + " ")
		}
		fmt.Println(peer.String())
		if peer.String() == destIP {
			os.Exit(0)
		}
	default:
		if len(domain) > 0 {
			fmt.Print(domain[0] + " ")
		}
		fmt.Println(peer.String())
		//log.Printf("got %+v; want echo reply", rm)
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
