package app

import (
	"sync"

	"github.com/hashicorp/go-multierror"
)

type Runner struct {
	stop    func()
	stopped <-chan struct{}
	eg      multierror.Group
}

func NewRunner() *Runner {
	o := &sync.Once{}
	stopCh := make(chan struct{})
	return &Runner{
		stop:    func() { o.Do(func() { close(stopCh) }) },
		stopped: stopCh,
		eg:      multierror.Group{},
	}
}

func (r *Runner) Go(f func() error) {
	r.eg.Go(func() error {
		defer r.stop()
		return f()
	})
}

func (r *Runner) Cleanup(f func() error) {
	r.eg.Go(func() error {
		<-r.stopped
		return f()
	})
}

func (r *Runner) Wait() error {
	return r.eg.Wait().ErrorOrNil()
}
