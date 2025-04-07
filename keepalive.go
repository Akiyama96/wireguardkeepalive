package main

import (
	"context"
	"fmt"
	"github.com/go-ping/ping"
	"log"
	"os/exec"
	"sync"
	"time"
)

type WireGuardConnection struct {
	status        bool
	address       string
	interfaceName string
	errorCount    int
	statusChannel chan bool
}

func NewWireGuardConnection(address, interfaceName string) *WireGuardConnection {
	return &WireGuardConnection{
		address:       address,
		interfaceName: interfaceName,
		statusChannel: make(chan bool),
	}
}

// 是否连接到了WireGuard服务
func (w *WireGuardConnection) isConnected() bool {
	return w.status
}

// 连接状态从连接转换到非连接时需要重启一下WireGuard服务
func (w *WireGuardConnection) onStatusDisconnected(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("context done")
			return
		case status, ok := <-w.statusChannel:
			if !ok {
				log.Println("status channel closed")
				return
			}

			if !status && w.errorCount > 3 {
				err := w.restartWireGuard(w.interfaceName)
				if err != nil {
					log.Println("restarting wireguard interface error: ", err)
					continue
				}

				log.Println("restarted wireguard interface")
			}
		}
	}
}

// 循环检查WireGuard服务的连接性
func (w *WireGuardConnection) statusCheckServer(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("context done")
			return
		case <-ticker.C:
			err := w.ping(w.address)
			if err != nil {
				w.status = false
				w.errorCount++
				w.statusChannel <- w.status
				log.Println("status check error: ", err)
				continue
			}

			w.status = true
			w.errorCount = 0
			w.statusChannel <- w.status
			log.Println("status check passed")
		}
	}
}

// 检查服务连通性
func (w *WireGuardConnection) ping(address string) error {
	pinger, err := ping.NewPinger(address)
	if err != nil {
		return err
	}

	pinger.Count = 3
	pinger.Timeout = time.Second * 5
	pinger.SetPrivileged(true)

	err = pinger.Run()
	if err != nil {
		return err
	}

	stats := pinger.Statistics()
	if stats.PacketLoss < 100 {
		return nil
	}

	return fmt.Errorf("ping failed: packet send=%d, packet recv=%d, packet loss=%f ", stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
}

// 重启WireGuard
func (w *WireGuardConnection) restartWireGuard(interfaceName string) error {
	stopCmd := exec.Command("wg-quick", "down", interfaceName)
	err := stopCmd.Run()
	if err != nil {
		log.Println("stopping wireguard interface: ", err)
	}
	log.Println("stop wireguard")

	startCmd := exec.Command("wg-quick", "up", interfaceName)
	err = startCmd.Run()
	if err != nil {
		return err
	}
	log.Println("start wireguard")

	return nil
}
