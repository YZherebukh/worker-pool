package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/worker-pool/workers/ordered"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		oscall := <-c
		log.Printf("system call: %+v \n", oscall)
		cancel()
	}()

	Do(ctx)
}

// is a struct that implement Response interface of ordered package
type response struct {
	messageCh chan string
	message   string
}

// Send implements Send methd of ordered.Response
func (r *response) Send() {
	r.messageCh <- r.message
}

// Send implements Close methd of ordered.Response
func (r *response) Close() {
	close(r.messageCh)
}

// worker is a worker struct
type worker struct {
	index   int         // worker's index, to keep track of outputed order
	message string      // just some message to print as a response example
	respCh  chan string // channel, into which responses will be published
}

// getResponse is a func that will be executed in each worker
func (w *worker) getResponse() ordered.Response {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(5)                          // n will be between 0 and 10
	time.Sleep(time.Duration(n) * time.Second) // sleep randomly and send a responsse after that
	w.index++                                  // increase index after each iteration
	return &response{
		messageCh: w.respCh,
		message:   w.message + strconv.Itoa(w.index),
	}
}

// Do does something
func Do(ctx context.Context) {
	responseCh := make(chan string)
	myWorkers := []*worker{ // 2 workers are created. to make the track of order easier, method is passed
		{
			message: "worker ",
			index:   100,
			respCh:  responseCh,
		},
		{
			message: "worker ",
			index:   200,
			respCh:  responseCh,
		},
		{
			message: "worker ",
			index:   300,
			respCh:  responseCh,
		},
		{
			message: "worker ",
			index:   400,
			respCh:  responseCh,
		},
		{
			message: "worker ",
			index:   500,
			respCh:  responseCh,
		},
		{
			message: "worker ",
			index:   600,
			respCh:  responseCh,
		},
		{
			message: "worker ",
			index:   700,
			respCh:  responseCh,
		},
		{
			message: "worker ",
			index:   800,
			respCh:  responseCh,
		},
		{
			message: "worker ",
			index:   900,
			respCh:  responseCh,
		},
		{
			message: "worker ",
			index:   1000,
			respCh:  responseCh,
		},
	}

	log.Printf("created %d workers\n", len(myWorkers))

	pool := ordered.New(ctx, len(myWorkers)) // create new pool with wnated number of workers
	ww := make([]func() ordered.Response, 0, len(myWorkers))
	for _, w := range myWorkers {
		ww = append(ww, w.getResponse) // create an array of ordered.Worker to pass to ordered.Pool in a specific order
	}

	pool.AddWorkers(ww...) // add workers to the pool
	go pool.Read()         // start reading responce sreom workers
	go pool.Start()        // start workers

	for r := range responseCh {
		fmt.Println("response ", r)
	}
}
