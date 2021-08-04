package lameduck

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

type testcase struct {
	signal        os.Signal
	signalAfter   time.Duration
	shutdownAfter time.Duration
	cancelAfter   time.Duration
	serveError    error
	shutdownError error
	closeError    error
	runOptions    []Option

	want error
}

var (
	errServeFailed    = errors.New("server failed to start")
	errShutdownFailed = errors.New("server failed to shutdown")
)

func TestRun(t *testing.T) {
	// Test cases are generated from a series of ordered events that should
	// result in a specific outcome (as declared by 'wantError'). If no error
	// is declared, a nil error is expected.
	cases := map[string]*testcase{
		"normal":   mkCase(sendSignal(unix.SIGTERM), shutdownReturn(nil), lameDuckExpires),
		"expired":  mkCase(sendSignal(unix.SIGTERM), lameDuckExpires, shutdownReturn(nil), wantError(&LameDuckError{Expired: true})),
		"canceled": mkCase(cancelContext, wantError(context.Canceled)),
		"nostart":  mkCase(serveReturn(errServeFailed), wantError(errServeFailed)),
		"badstop":  mkCase(sendSignal(unix.SIGTERM), shutdownReturn(errShutdownFailed), lameDuckExpires, wantError(&LameDuckError{Err: errShutdownFailed})),
	}

	for label, tc := range cases {
		t.Run(label, tc.test)
	}
}

func (tc *testcase) test(t *testing.T) {
	if tc.cancelAfter == 0 && tc.signalAfter == 0 && tc.serveError == nil {
		t.Fatal("Invalid testcase: must set one of 'cancelAfter', 'signalAfter', or 'serveError'")
	}

	ts := injectSignaller()
	defer ts.revert()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tl := &testLogger{t.Logf}

	tc.runOptions = append(tc.runOptions, WithLogger(tl))

	svr := newTestServer(tl, tc.serveError, tc.shutdownError, tc.closeError)

	errs := make(chan error)

	var wg sync.WaitGroup

	r, err := NewRunner(svr, tc.runOptions...)
	if err != nil {
		t.Fatalf("cannot create Runner: %v", err)
	}

	tc.chkState(t, r, NotStarted)

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(errs)

		if err := r.Run(ctx); err != nil {
			tl.Infof("Run error: %v", err)
			errs <- err
		} else {
			tl.Infof("Server Run Successful")
		}
	}()

	if tc.cancelAfter != 0 {
		wg.Add(1)
		tl.Infof("Will cancel context after %v", tc.cancelAfter)
		time.AfterFunc(tc.cancelAfter, func() {
			defer wg.Done()
			tl.Infof("Cancelling Context")
			cancel()
		})
	}

	if tc.signal != nil && tc.signalAfter != 0 {
		wg.Add(1)
		tl.Infof("Will emit signal %q after %v", tc.signal, tc.signalAfter)
		time.AfterFunc(tc.signalAfter, func() {
			defer wg.Done()
			tl.Infof("Emitting signal: %v", tc.signal)
			ts.emit(tc.signal)
		})
	}

	if tc.shutdownAfter != 0 {
		wg.Add(1)
		tl.Infof("Will finish Shutdown after %v", tc.shutdownAfter)
		time.AfterFunc(tc.shutdownAfter, func() {
			defer wg.Done()
			tl.Infof("Finishing server Shutdown method")
			svr.shutdown.finish()
		})
	}

	r.Ready()

	tc.chkState(t, r, Running)

	if got := <-errs; !tc.isWanted(got) {
		t.Errorf("Run(ctx, svr) == (%v); wanted (%v)", got, tc.want)
	}

	wg.Wait()

	tc.chkState(t, r, Stopped)
}

func (tc *testcase) isWanted(got error) bool {
	if lde, ok := tc.want.(*LameDuckError); ok {
		return lde.isEqual(got)
	}

	return got == tc.want
}

func (tc *testcase) chkState(t *testing.T, r *Runner, want State) {
	t.Helper()

	t.Logf("want:%v tc:%#v", want, tc)

	switch want {
	case Stopped:
		if tc.want == context.Canceled {
			want = Failed
		}
		fallthrough

	case Running:
		time.Sleep(100 * time.Microsecond) // to compensate for test flakiness
		if tc.serveError == errServeFailed {
			want = Failed
		}
	}

	if got := r.State(); got != want {
		t.Errorf("r.State() == %v; wanted %v", got, want)
	}
}

// A convenience method added to type LameDuckError (only during testing).
func (lde *LameDuckError) isEqual(err error) bool {
	switch {
	case lde == nil && err == nil:
		return true
	case lde == nil || err == nil:
		return false
	}

	if olde, ok := err.(*LameDuckError); ok {
		return lde.Expired == olde.Expired && lde.Err == olde.Err
	}

	return false
}
