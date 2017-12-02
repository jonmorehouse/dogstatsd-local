package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
)

type msg struct {
	addr string
	data []byte
}

func server(host string, port int, msgCh chan msg) error {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		return err
	}
	defer serverConn.Close()

	buf := make([]byte, 1024)
	for {
		n, addr, err := serverConn.ReadFromUDP(buf)
		msgCh <- msg{
			addr: addr.String(),
			data: buf[:n],
		}

		log.Printf("Received %d", n)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func main() {
	port := flag.Int("listen-port", 8126, "listen port; default 8126")
	host := flag.String("listen-host", "0.0.0.0", "listen host; default 0.0.0.0")
	flag.Parse()

	handlerCh := make(chan msg, 10)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for _ = range handlerCh {
			log.Println("connection received...")
		}

		wg.Done()
	}()

	if err := server(*host, *port, handlerCh); err != nil {
		log.Fatalf(err.Error())
	}

	close(handlerCh)
	wg.Wait()
}
