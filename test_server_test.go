package lameduck

import (
	"context"
	"sync"
)

// NOTE: This file contains no tests of its own.
//
// 			 Instead, it contains the definition for type testServer - a test
// 			 object implementing this package's Server interface.
//
type testServer struct {
	logger   Logger
	serve    *gate
	shutdown *gate
	close    *gate
}

func newTestServer(logger Logger, serveErr, shutdownErr, closeErr error) *testServer {
	serveBlocking := true
	if serveErr != nil {
		serveBlocking = false
	}

	ts := &testServer{
		logger:   logger,
		serve:    newGate(serveErr, serveBlocking),
		shutdown: newGate(shutdownErr, true),
		close:    newGate(closeErr, false),
	}

	return ts
}

func (ts *testServer) Serve(ctx context.Context) error {
	err := ts.serve.wait(ctx)
	ts.logger.Infof("Serve returned: %v", err)
	return err
}

func (ts *testServer) Shutdown(ctx context.Context) error {
	ts.serve.finish()
	return ts.shutdown.wait(ctx)
}

func (ts *testServer) Close() error {
	ts.serve.finish()
	return ts.close.wait(context.Background())
}

type gate struct {
	err  error
	done chan struct{}
	once sync.Once
}

func newGate(err error, blocking bool) *gate {
	g := &gate{err: err}
	if blocking {
		g.done = make(chan struct{})
	}

	return g
}

func (g *gate) wait(ctx context.Context) error {
	if g.done == nil {
		return g.err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-g.done:
		return g.err
	}
}

func (g *gate) finish() {
	if g != nil && g.done != nil {
		g.once.Do(func() { close(g.done) })
	}
}
