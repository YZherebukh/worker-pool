package ordered

type worker struct {
	work func() Response
}

// newWorker creates new worker
func newWorker(w func() Response) *worker {
	return &worker{
		work: w,
	}
}

func (w *worker) Do(ch chan Response) {
	ch <- w.work()
}
