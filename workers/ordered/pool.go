package ordered

import (
	"context"
)

// Response is a responce interface
type Response interface {
	Send()
	// Err(errCh chan error)
}

// Worker is a worker interface
type Worker interface {
	Do(ch chan Response)
}

// Pool is a workers struct
type Pool struct {
	ctx context.Context
	// mu             *sync.Mutex
	workerNumberCh chan int
	workers        []Worker
	responseChnls  []chan Response
	onErrorStop    bool
}

// New creates new ordered workers pool
func New(ctx context.Context, workersNumber int) *Pool {
	pool := &Pool{
		ctx: ctx,
		// mu:             &sync.Mutex{},
		workerNumberCh: make(chan int, workersNumber),
		// workers:        make([]Worker, workersNumber),
	}

	pool.responseChnls = make([]chan Response, workersNumber)
	for i := range pool.responseChnls {
		pool.responseChnls[i] = make(chan Response)
	}

	for i := 0; i < workersNumber; i++ {
		pool.workerNumberCh <- i
	}

	return pool
}

// AddWorkers is adding workers to run the job
func (p *Pool) AddWorkers(workers ...Worker) {
	p.workers = workers
}

// Start starts ordered worker's pool
func (p *Pool) Start() {
	for {
		//log.Println("waiting for workers")
		for i := range p.workerNumberCh {
			//	log.Printf("worker %d start\n", i)

			go p.workers[i].Do(p.responseChnls[i])
		}
	}
}

// Read reads responce in order from workers
func (p *Pool) Read() {
	for {
		if len(p.responseChnls) == 0 {
			return
		}

		for i := 0; i < len(p.responseChnls); i++ {
			select {
			case <-p.ctx.Done():
				p.readAndClose(p.responseChnls[i])

			case resp, ok := <-p.responseChnls[i]:
				//		log.Printf("worker %d arrived\n", i)

				if !ok {
					p.responseChnls[i] = nil
					copy(p.responseChnls[i:], p.responseChnls[i+1:])
					p.responseChnls = p.responseChnls[:len(p.responseChnls)-1]

					if i > 0 {
						i--
					}

					continue
				}

				resp.Send()
				p.workerNumberCh <- i
				//	log.Printf("worker %d has been send\n", i)
			}
		}
	}
}

func (p *Pool) readAndClose(respCh chan Response) {
	response := <-respCh
	response.Send()

	close(respCh)
}
