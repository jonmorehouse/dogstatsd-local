package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
)

type msgHandler func([]byte) error

type asyncMsgHandler interface {
	handler([]byte) error
	stop()
}

type handler struct {
	fn    msgHandler
	msgCh chan ([]byte)
	wg    sync.WaitGroup
}

func newAsyncMsgHandler(fn msgHandler, poolSize int, bufferSize int) asyncMsgHandler {
	h := &handler{
		fn:    fn,
		msgCh: make(chan []byte, bufferSize),
	}

	h.wg.Add(poolSize)

	// build a pool of goroutines to listen for messages to process
	for i := 0; i < poolSize; i++ {
		go func() {
			defer h.wg.Done()
			for msg := range h.msgCh {
				h.fn(msg)
			}
		}()
	}

	return h
}

// submit to the pool
func (a handler) handler(msg []byte) error {
	select {
	case a.msgCh <- msg:
		return nil
	default:
	}

	return errors.New("POOL_CAPACITY_EXCEEDED")
}

func (a *handler) stop() {
	close(a.msgCh)
	a.wg.Wait()
}

func main() {
	host := flag.String("host", "0.0.0.0", "bind address")
	port := flag.Int("port", 8125, "listen port")
	format := flag.String("format", "stdout", "output format: json|std|raw")
	rawTags := flag.String("tags", "", "extra tags: comma delimited")
	flag.Parse()

	extraTags := strings.Split(*rawTags, ",")
	var handler msgHandler

	if *format == "json" {
		handler = newJsonDogstatsdMsgHandler(extraTags)
	} else if *format == "human" {
		handler = newHumanDogstatsdMsgHandler(extraTags)
	} else {
		handler = newRawDogstatsdMsgHandler()
	}

	asyncHandler := newAsyncMsgHandler(handler, 1000, 10000)

	var wg sync.WaitGroup

	// create a new server and listen on a background goroutine
	addr := fmt.Sprintf("%s:%d", *host, *port)
	srv := newServer(addr, asyncHandler.handler)
	wg.Add(1)
	go func(srv server) {
		defer wg.Done()
		if err := srv.listen(); err != nil {
			log.Fatalf(err.Error())
		}
	}(srv)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	if err := srv.stop(); err != nil {
		log.Println(err.Error())
	}
	wg.Wait()
	asyncHandler.stop()
}
