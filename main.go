package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/ipv4"
)

type multiCastsAddress struct {
	ip   net.IP
	port int
}

func main() {
	candidates := []multiCastsAddress{
		{
			port: 8025,
			ip:   net.ParseIP("224.0.0.1"),
		}, {
			port: 8012,
			ip:   net.ParseIP("224.0.0.2"),
		}, {
			port: 1111,
			ip:   net.ParseIP("224.0.0.3"),
		}, {
			port: 1345,
			ip:   net.ParseIP("224.0.0.4"),
		}, {
			port: 1235,
			ip:   net.ParseIP("224.0.0.1"),
		},
	}

	if err := interfaceAdd("Ethernet", candidates, msgHandler); err != nil {
		log.Printf("main: error: %v", err)
	}
	sendMessage("224.0.0.1:11049")
	log.Printf("main: waiting forever")
	<-make(chan int)
}

func msgHandler(ipaddrs string, n int, b []byte) {
	log.Println(n, "bytes read from", ipaddrs)
	log.Println(hex.Dump(b[:n]))
}

func interfaceAdd(s string, candidates []multiCastsAddress, h func(string, int, []byte)) error {

	iface, err1 := net.InterfaceByName(s)
	if err1 != nil {
		return err1
	}

	addrList, err2 := iface.Addrs()
	if err2 != nil {
		return err2
	}

	for _, a := range addrList {
		addr, _, err3 := net.ParseCIDR(a.String())
		if err3 != nil {
			log.Printf("interfaceAdd: parse CIDR error for '%s' on '%s': %v", addr, s, err3)
			continue
		}

		if err := join(iface, candidates, h); err != nil {
			log.Printf("interfaceAdd: join error for '%s' on '%s': %v", addr, s, err)
		}
	}

	return nil
}

func join(iface *net.Interface, candidates []multiCastsAddress, h func(string, int, []byte)) error {
	proto := "udp"
	for x := range candidates {
		var a string
		if candidates[x].ip.To4() == nil {
			// IPv6
			a = fmt.Sprintf("[%s]", candidates[x].ip.String())
		} else {
			// IPv4
			a = candidates[x].ip.String()
		}

		hostPort := fmt.Sprintf("%s:%d", a, candidates[x].port) // rip multicast port

		// open socket (connection)
		conn, err2 := net.ListenPacket(proto, hostPort)
		if err2 != nil {
			return fmt.Errorf("join: %s/%s listen error: %v", proto, hostPort, err2)
		}

		// join multicast address
		pc := ipv4.NewPacketConn(conn)
		if err := pc.JoinGroup(iface, &net.UDPAddr{IP: candidates[x].ip}); err != nil {
			conn.Close()
			return fmt.Errorf("join: join error: %v", err)
		}

		// request control messages
		/*
		   if err := pc.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true); err != nil {
		       // warning only
		       log.Printf("join: control message flags error: %v", err)
		   }
		*/

		go udpReader(pc, candidates[x].ip.String(), candidates[x].ip.String(), h)
	}
	return nil
}

func udpReader(c *ipv4.PacketConn, ifname, ifaddr string, h func(string, int, []byte)) {

	log.Printf("udpReader: reading from '%s' on '%s'", ifaddr, ifname)

	defer c.Close()

	buf := make([]byte, 1000)

	for {
		n, _, _, err := c.ReadFrom(buf)
		if err != nil {
			log.Printf("udpReader: ReadFrom: error %v", err)
			break
		}

		// make a copy because we will overwrite buf
		b := make([]byte, n)
		copy(b, buf)

		h(ifname, n, b)
	}

	log.Printf("udpReader: exiting '%s'", ifname)
}

func sendMessage(addr string) {
	conn, err := newBroadcaster(addr)
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte("hello, world"))
}

func newBroadcaster(address string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return nil, err
	}

	return conn, nil

}
