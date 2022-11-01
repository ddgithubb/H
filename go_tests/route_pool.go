package main

import (
	"fmt"

	cmap "github.com/streamrail/concurrent-map"
)

const WORKER_AMOUNT = 10000

type routerPool struct {

	connections cmap.ConcurrentMap

	router chan string

	stop_chan chan struct{}

	test_chan chan string

	//rw sync.RWMutex
}

func (rp *routerPool) Initialize() {
	rp.router = make(chan string, 100000000)
	rp.stop_chan = make(chan struct{})
	rp.test_chan = make(chan string, 100000000)
	rp.connections = cmap.New()

	for i := 0; i < WORKER_AMOUNT; i++ {
		go routerWorker(rp)
	}
}

func (rp *routerPool) Close() {
	rp.stop_chan <- struct{}{}
}

func routerWorker(rp *routerPool) {
	for  { 
		select {
		case msg := <-rp.router:
			//fmt.Println("Router worker", msg)
			_, ok := rp.connections.Get(msg)
			if ok {
				
			}
			rp.test_chan <- "done: " + fmt.Sprint(ok)
			//Parse
			//Handle
			// msgChan, ok := connections.Get(connectionID)
			// if ok {
			// 	msgChan.(chan string) <- payload
			// }
		case <-rp.stop_chan:
			return
		}
	}
}