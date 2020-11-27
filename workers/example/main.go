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

type myStruct struct {
}

type response struct {
	messageCh chan string
	message   string
	err       error
}

func (r response) Send() {
	r.messageCh <- r.message
}

type worker struct {
	message string
	index   int
	respCh  chan string
}

func (w *worker) getResponse() ordered.Response {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(5) // n will be between 0 and 10
	time.Sleep(time.Duration(n) * time.Second)
	w.index++
	return response{
		messageCh: w.respCh,
		message:   w.message + strconv.Itoa(w.index),
	}
}

// Do does something
func Do(ctx context.Context) {
	responseCh := make(chan string)
	myWorkers := []*worker{
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

	workers := make([]ordered.Worker, 0, len(myWorkers))
	for _, w := range myWorkers {
		workers = append(workers, ordered.NewWorker(w.getResponse))
	}

	pool := ordered.New(ctx, len(myWorkers))
	pool.AddWorkers(workers...)
	go pool.Start()
	go pool.Read()

	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-responseCh:
			if !ok {
				return
			}
			fmt.Println("response ", r)
		}
	}
}
