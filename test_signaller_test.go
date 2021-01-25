package lameduck

import (
	"os"
)

// NOTE: This file contains no tests of its own.
//
//       Instead, it contains the testSignaller implemenation for emitting fake
//       signals into the main code base without the need for actually using
//       real (OS level) signaling.
//

type testSignaller struct {
	orig signaler
	sigs map[os.Signal]bool
	ch   chan<- os.Signal
}

// injectSignaller injects and returns a new testSignaller into the main codebase.
func injectSignaller() *testSignaller {
	ts := &testSignaller{orig: sig}
	sig = ts
	return ts
}

// emit sends a Signal through the testSignaller.
func (ts *testSignaller) emit(s os.Signal) {
	if ts == nil || ts.ch == nil || ts.sigs == nil || !ts.sigs[s] {
		return
	}

	go func() { ts.ch <- s }()
}

// revert replaces an injected testSignaller with the signaler in place at the
// time of injection.
func (ts *testSignaller) revert() {
	if ts != nil && ts.orig != nil {
		sig = ts.orig
	}
}

// notify contributes to the signaler interface
func (ts *testSignaller) notify(c chan<- os.Signal, sig ...os.Signal) {
	ts.ch = c

	if ts.sigs == nil {
		ts.sigs = make(map[os.Signal]bool)
	}

	for _, s := range sig {
		ts.sigs[s] = true
	}
}

// stop contributes to the signaler interface
func (ts *testSignaller) stop(c chan<- os.Signal) {
	if ts != nil && ts.ch == c {
		ts.sigs = nil
	}
}
