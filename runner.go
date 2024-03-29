package lameduck

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/sys/unix"
	"toolman.org/base/log/v2"
)

var (
	defaultPeriod  = 3 * time.Second
	defaultSignals = []os.Signal{unix.SIGINT, unix.SIGTERM}
)

// Runner is the lame-duck coordinator for a type implementing the Server
// interface.
type Runner struct {
	server  Server
	period  time.Duration
	escOK   bool
	signals []os.Signal
	logf    func(string, ...interface{})
	psHook  hookFunction
	state   State
	ready   chan struct{}
	done    chan struct{}

	once sync.Once
}

func newRunner(svr Server, options []Option) (*Runner, error) {
	if svr == nil {
		return nil, errors.New("nil Server")
	}

	r := &Runner{
		server:  svr,
		period:  defaultPeriod,
		signals: defaultSignals,
		logf:    log.Infof,
		state:   NotStarted,
		ready:   make(chan struct{}),
		done:    make(chan struct{}),
	}

	for _, o := range options {
		o.set(r)
	}

	if r.period <= 0 {
		return nil, errors.New("lame-duck period must be greater than zero")
	}

	if len(r.signals) == 0 {
		return nil, errors.New("no lame-duck signals defined")
	}

	return r, nil
}

func (r *Runner) serve(ctx context.Context) error {
	if r == nil {
		return errors.New("bad state: nil receiver")
	}

	err := r.server.Serve(ctx)

	switch {
	case err == nil:
		return nil
	case r.escOK && err == http.ErrServerClosed:
		return nil
	default:
		return err
	}
}

// Ready returns a channel that is closed when the receiver's underlying
// Server is ready to serve reqeuests.
func (r *Runner) Ready() <-chan struct{} {
	return r.ready
}

func (r *Runner) close() {
	if r == nil || r.done == nil {
		if r != nil {
			r.logf("r.done is nil !!!")
		}
		return
	}

	var closed bool

	r.once.Do(func() {
		close(r.done)
		r.logf("runner closed")
		closed = true
	})

	if !closed {
		r.logf("runner *NOT* closed")
	}
}
