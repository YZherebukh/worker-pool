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
		ctx: ctx,
	}

	return pool
}

// AddWorkers is adding workers to run the job
// expected that each specific worker will have his own set of parameters
func (p *Pool) AddWorkers(ff ...func() Response) {
	workers := make([]Worker, 0, len(ff))

	for _, f := range ff {
		workers = append(workers, newWorker(f))
	}

	p.addWorkers(workers)
}

func (p *Pool) addWorkers(workers []Worker) {
	p.workers = workers

	p.workerNumberCh = make(chan int, len(workers)) // create a chnanel, that informes which worker should pick up the job

	p.responseChnls = make([]chan Response, len(workers)) // slice of channels with a responses of each worker, in a specific order

	for i := range p.responseChnls {
		p.responseChnls[i] = make(chan Response) // creating response channel for each worker
		p.workerNumberCh <- i                    // sending worker number to the queue for it to be picked up later
	}
}

// AddWorker is adding workers to run the job
// expected that each specific worker will have his own set of parameters
func (p *Pool) AddWorker(workersNumber int, f func() Response) {
	workers := make([]Worker, 0, workersNumber)

	for i := range p.workers {
		workers[i] = newWorker(f)
	}

	p.addWorkers(workers)
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
	close(respCh)

	response.Send()
	if len(p.responseChnls) == 1 {
		response.Close()
	}
}
