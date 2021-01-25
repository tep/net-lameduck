package lameduck

import (
	"time"

	"github.com/spf13/pflag"
)

// NOTE: This file contains no tests of its own.
//
//       Instead, it provides the definition for type testLogger -- used to
//       inject T.Logf (from package testing) as a lameduck.Logger.

func init() {
	// To stifle warnings from logger.
	pflag.Parse()
}

type testLogger struct {
	logf func(string, ...interface{})
}

func (tl *testLogger) Infof(msg string, args ...interface{}) {
	ts := time.Now().Format("[15:04:05.000000] ")
	tl.logf(ts+msg, args...)
}
