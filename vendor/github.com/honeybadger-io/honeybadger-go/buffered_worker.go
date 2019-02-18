package honeybadger

import "fmt"

var (
	errWorkerOverflow = fmt.Errorf("The worker is full; this envelope will be dropped.")
)

func newBufferedWorker(config *Configuration) *bufferedWorker {
	worker := &bufferedWorker{ch: make(chan envelope, 100)}
	go func() {
		for w := range worker.ch {
			work := func() error {
				defer func() {
					if err := recover(); err != nil {
						config.Logger.Printf("worker recovered from panic: %v\n", err)
					}
				}()
				return w()
			}
			if err := work(); err != nil {
				config.Logger.Printf("worker processing error: %v\n", err)
			}
		}
	}()
	return worker
}

type bufferedWorker struct {
	ch chan envelope
}

func (w *bufferedWorker) Push(work envelope) error {
	select {
	case w.ch <- work:
		return nil
	default:
		return errWorkerOverflow
	}
}

func (w *bufferedWorker) Flush() {
	ch := make(chan bool)
	w.ch <- func() error {
		ch <- true
		return nil
	}
	<-ch
}
