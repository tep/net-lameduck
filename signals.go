package lameduck

import (
	"context"
	"os"
	"os/signal"
)

type osSignals struct{}

func (*osSignals) notify(c chan<- os.Signal, sig ...os.Signal) { signal.Notify(c, sig...) }
func (*osSignals) stop(c chan<- os.Signal)                     { signal.Stop(c) }

var sig signaler = new(osSignals)

type signaler interface {
	notify(chan<- os.Signal, ...os.Signal)
	stop(chan<- os.Signal)
}

func (r *runner) waitForSignal(ctx context.Context) (os.Signal, error) {
	ch := make(chan os.Signal, 1)
	defer close(ch)

	sig.notify(ch, r.signals...)
	defer sig.stop(ch)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case sig := <-ch:
		return sig, nil
	}
}
