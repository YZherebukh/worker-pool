package ordered

import (
	"context"
)

// Response is a responce interface
// any resp struct, expected to be returned by worker, should imlement this interface
type Response interface {
	Send()
	Close()
}

// Worker is a worker interface
type Worker interface {
	Do(ch chan Response)
}

// Pool is a workers struct
type Pool struct {
	ctx            context.Context
	workers        []Worker
	workerNumberCh chan int
	responseChnls  []chan Response
}

// New creates new ordered workers pool
func New(ctx context.Context, workersNumber int) *Pool {
	pool := &Pool{
		ctx:            ctx,
		workerNumberCh: make(chan int, workersNumber), // create a chnanel, that informes which worker should pick up the job
	}

	pool.responseChnls = make([]chan Response, workersNumber) // slice of channels with a responses of each worker, in a specific order
	for i := range pool.responseChnls {
		pool.responseChnls[i] = make(chan Response)
		pool.workerNumberCh <- i // sending worker number to the queue for it to be picked up later
	}

	return pool
}

// AddWorkers is adding workers to run the job
// expected that each specific worker will have his own set of parameters
func (p *Pool) AddWorkers(ff ...func() Response) {
	p.workers = make([]Worker, 0, len(p.workerNumberCh))

	for _, f := range ff {
		p.workers = append(p.workers, newWorker(f))
	}
}

// AddWorker is adding workers to run the job
// expected that each specific worker will have his own set of parameters
func (p *Pool) AddWorker(f func() Response) {
	p.workers = make([]Worker, 0, len(p.workerNumberCh))

	for i := range p.workers {
		p.workers[i] = newWorker(f)
	}
}

// Start starts ordered worker's pool
func (p *Pool) Start() {
	for i := range p.workerNumberCh { // waiting for a worker numer to arrive, so the specific worker can be strted
		go p.workers[i].Do(p.responseChnls[i])
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
				i--
			case resp, ok := <-p.responseChnls[i]:
				if !ok {
					copy(p.responseChnls[i:], p.responseChnls[i+1:])
					p.responseChnls = p.responseChnls[:len(p.responseChnls)-1]
					if i > 0 {
						i--
					}
					continue
				}
				resp.Send()

				p.workerNumberCh <- i
			}
		}
	}
}

func (p *Pool) readAndClose(respCh chan Response) {
	response, ok := <-respCh
	if !ok {
		return
	}

	response.Send()
	if len(p.responseChnls) == 1 {
		response.Close()
	}
	close(respCh)
}
