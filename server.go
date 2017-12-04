package main

import (
	"log"
	"net"
	"sync"
	"time"
)

type server interface {
	listen() error
	stop() error
}

func newServer(addr string, fn msgHandler) server {
	return &udpServer{
		msgHandler:    fn,
		rawAddr:       addr,
		readDeadline:  time.Second / 4,
		writeDeadline: time.Second / 4,
		errCh:         make(chan error, 1),
		stopCh:        make(chan struct{}),
		wg:            sync.WaitGroup{},
	}
}

type udpServer struct {
	msgHandler msgHandler
	rawAddr    string

	readDeadline  time.Duration
	writeDeadline time.Duration

	stopCh chan struct{}
	errCh  chan error

	wg sync.WaitGroup
}

func (u *udpServer) listen() error {
	addr, err := net.ResolveUDPAddr("udp", u.rawAddr)
	if err != nil {
		return err
	}

	serverConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	u.wg.Add(1)
	go u.errHandler()

	buf := make([]byte, 1024)
	respMsg := []byte{}

	for {
		select {
		case <-u.stopCh:
			goto stop
		default:
		}

		serverConn.SetDeadline(time.Now().Add(u.readDeadline))

		n, clientAddr, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			// check if the error is a timeout error
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			u.errCh <- err
			continue
		}

		// copy the message and pass it to the handler function
		msg := make([]byte, n)
		copy(msg, buf[:n])
		u.msgHandler(msg)

		// respond to the origin connection
		serverConn.SetWriteDeadline(time.Now().Add(u.writeDeadline))
		_, err = serverConn.WriteToUDP(respMsg, clientAddr)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			u.errCh <- err
		}
	}

stop:
	serverConn.Close()
	close(u.stopCh)
	return nil
}

func (u *udpServer) errHandler() {
	for err := range u.errCh {
		log.Println(err.Error())
	}

	u.wg.Done()
}

func (u *udpServer) stop() error {
	// stop the server and wait for it to finish
	u.stopCh <- struct{}{}
	<-u.stopCh

	// close the err channel and wait for any in progress errors to complete
	close(u.errCh)
	u.wg.Wait()
	return nil
}
