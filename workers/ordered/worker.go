package ordered

type worker struct {
	work func() Response
}

// NewWorker creates new worker
func NewWorker(w func() Response) Worker {
	return &worker{
		work: w,
	}
}

func (w *worker) Do(ch chan Response) {
	ch <- w.work()
}
