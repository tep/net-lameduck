package lameduck

import (
	"os"
	"time"
)

// NOTE: This file contains no tests of its own.
//
//       Instead it provides the definitions for testEvent and evtType. These
//       are used by the mkCase function to generate test cases from a series
//       of ordered events.
//

type evtType int

const (
	evtNone evtType = iota
	evtCancelContext
	evtCloseReturn
	evtLameDuckExpires
	evtSendSignal
	evtServeReturn
	evtShutdownReturn
	evtWantError
)

type testEvent struct {
	etype  evtType
	signal os.Signal
	err    error
}

var (
	lameDuckExpires = testEvent{etype: evtLameDuckExpires}
	cancelContext   = testEvent{etype: evtCancelContext}
)

func closeReturn(err error) testEvent    { return testEvent{etype: evtCloseReturn, err: err} }
func sendSignal(sig os.Signal) testEvent { return testEvent{etype: evtSendSignal, signal: sig} }
func serveReturn(err error) testEvent    { return testEvent{etype: evtServeReturn, err: err} }
func shutdownReturn(err error) testEvent { return testEvent{etype: evtShutdownReturn, err: err} }
func wantError(err error) testEvent      { return testEvent{etype: evtWantError, err: err} }

// mkCase converts a series of ordered testEvents into a testcase.
func mkCase(evts ...testEvent) *testcase {
	tc := new(testcase)

	var ld time.Duration

	gap := 10 * time.Millisecond
	after := gap

	for _, e := range evts {
		switch e.etype {
		case evtCancelContext:
			tc.cancelAfter = after
			after += gap

		case evtCloseReturn:
			tc.closeError = e.err

		case evtLameDuckExpires:
			ld = after
			after += gap

		case evtSendSignal:
			tc.signal = e.signal
			tc.signalAfter = after
			after += gap

		case evtServeReturn:
			tc.serveError = e.err

		case evtShutdownReturn:
			tc.shutdownError = e.err
			tc.shutdownAfter = after
			after += gap
			if ld != 0 {
				after += gap / 2
			}

		case evtWantError:
			tc.want = e.err
		}
	}

	if ld < tc.signalAfter {
		panic("cannot expire lame-duck before sending signal")
	}

	if ld > 0 {
		tc.runOptions = append(tc.runOptions, Period(ld-tc.signalAfter))
	}

	return tc
}
