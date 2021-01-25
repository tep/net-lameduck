package lameduck

import (
	"errors"
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

type runner struct {
	server  Server
	period  time.Duration
	signals []os.Signal
	logf    func(string, ...interface{})
	done    chan struct{}

	once sync.Once
}

func newRunner(svr Server, options []Option) (*runner, error) {
	if svr == nil {
		return nil, errors.New("nil Server")
	}

	r := &runner{
		server:  svr,
		period:  defaultPeriod,
		signals: defaultSignals,
		logf:    log.Infof,
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

func (r *runner) close() {
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
