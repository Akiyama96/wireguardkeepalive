package main

import (
	"context"
	"flag"
	"log"
	"sync"
)

func main() {
	interfaceName := flag.String("i", "wg0", "wireguard interface name")
	hostAddress := flag.String("h", "10.0.0.1", "host ip address")
	flag.Parse()

	var (
		wg  = new(sync.WaitGroup)
		ctx = context.Background()
	)
	c := NewWireGuardConnection(*hostAddress, *interfaceName)

	wg.Add(1)
	go c.statusCheckServer(ctx, wg)
	wg.Add(1)
	go c.onStatusDisconnected(ctx, wg)
	log.Printf("wireguard keepalive server start, host=%s, interface=%s", *hostAddress, *interfaceName)

	wg.Wait()
}
