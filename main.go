package main

import (
	"flag"
	"log"
)

func main() {
	interfaceName := flag.String("i", "wg0", "wireguard interface name")
	hostAddress := flag.String("h", "10.0.0.1", "host ip address")
	flag.Parse()

	c := NewWireGuardConnection(*hostAddress, *interfaceName)
	go c.statusCheckServer()
	go c.onStatusDisconnected()
	log.Printf("wireguard keepalive server start, host=%s, interface=%s", *hostAddress, *interfaceName)
}
