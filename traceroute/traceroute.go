package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const WAIT_TIME = 3000

func main() {
	arg := os.Args[1:]
	destIP := arg[0]
	sum := false
	var err error
	domain, err := net.LookupIP(destIP)
	if err != nil {
		log.Fatal(err)
	}
	destIP = domain[len(domain)-1].String()
	prob := 3

	var n_flag bool
	if flag, _ := contains(arg, "-n"); flag {
		n_flag = true
	}
	if flag, i := contains(arg, "-q"); flag {
		prob, err = strconv.Atoi(arg[i+1])
		if err != nil {
			log.Fatal("Cannot convert string to int")
		}
	}
	if flag, _ := contains(arg, "-S"); flag {
		sum = true
	}
	ttl := 1
	temp := 0
	dest_reached := false
	for {
		first_flag := false
		prob_not_ans := 0
		for i := prob; i > 0; i-- {
			time.Sleep(1 * time.Second)
			temp, dest_reached = Ping(destIP, ttl, n_flag, first_flag, i)
			prob_not_ans += temp
			first_flag = true
		}
		if sum {
			fmt.Printf("\n%d probes not answered\n", prob_not_ans)
		}
		fmt.Println()

		if dest_reached {
			os.Exit(0)
		}
		ttl += 1
	}
}

func Ping(destIP string, ttl int, n_flag bool, first_flag bool, prob int) (int, bool) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	conn.IPv4PacketConn().SetTTL(ttl)
	if err != nil {
		log.Fatalf("Packet listen err, %s", err)
	}
	defer conn.Close()

	start_ms := time.Now().UnixMicro()

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

	time := float64(time.Now().UnixMicro()-start_ms) / 1000

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		// log.Printf("got reflection from %v", peer)
		if !first_flag {
			if len(domain) > 0 {
				fmt.Print(domain[0] + " ")
			}
			fmt.Printf("(%v)", peer.String())
		}
		if time <= 3000 {
			fmt.Printf(" %.3f ms", time)
		} else {
			fmt.Printf(" *")
			return 1, true
		}
		return 0, true
	default:
		if !first_flag {
			if len(domain) > 0 {
				fmt.Print(domain[0] + " ")
			}
			fmt.Printf("(%v)", peer.String())
		}
		if time <= 30 {
			fmt.Printf(" %.3f ms", time)
		} else {
			fmt.Printf(" *")
			return 1, false
		}
	}
	return 0, false
}

func contains(s []string, e string) (bool, int) {
	for i := 0; i < len(s); i++ {
		if s[i] == e {
			return true, i
		}
	}
	return false, -1
}
