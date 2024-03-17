package app

import (
	"sync"

	"github.com/hashicorp/go-multierror"
)

type Runnable interface {
	Run() error
	Stop() error
}

type Runner struct {
	runnables []Runnable
	stop      func()
	stopped   <-chan struct{}
	eg        multierror.Group
}

func NewRunner(runnables ...Runnable) *Runner {
	o := &sync.Once{}
	stopCh := make(chan struct{})
	return &Runner{
		runnables: runnables,
		stop:      func() { o.Do(func() { close(stopCh) }) },
		stopped:   stopCh,
		eg:        multierror.Group{},
	}
}

func (r *Runner) Run() error {
	for _, runnable := range r.runnables {
		runnable := runnable
		r.eg.Go(func() error {
			defer r.stop()
			return runnable.Run()
		})
		r.eg.Go(func() error {
			<-r.stopped
			return runnable.Stop()
		})
	}
	return r.eg.Wait().ErrorOrNil()
}

// Stop stops all runnables. Usually this method is called only in tests.
func (r *Runner) Stop() {
	r.stop()
}
