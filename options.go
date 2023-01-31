package lameduck

import (
	"context"
	"os"
	"time"
)

// Option is the interface implemented by types that offer optional behavior
// while running a Server with lame-duck support.
type Option interface {
	set(*Runner)
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

// Period returns an Option that alters the lame-duck period to the given
// Duration.
func Period(p time.Duration) Option {
	return period(p)
}

type period time.Duration

func (p period) set(r *Runner) {
	r.period = time.Duration(p)
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

// Signals returns an Options that changes the list of Signals that trigger the
// beginning of lame-duck mode. Using this Option fully replaces the previous
// list of triggering signals.
func Signals(s ...os.Signal) Option {
	return signals(s)
}

type signals []os.Signal

func (s signals) set(r *Runner) {
	r.signals = ([]os.Signal)(s)
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

// Logger is the interface needed for the WithLogger Option.
type Logger interface {
	Infof(string, ...interface{})
}

type loggerOption struct {
	logger Logger
}

// WithLogger returns an Option that alters this package's logging facility
// to the provided Logger. Note, the default Logger is one derived from
// 'github.com/golang/glog'. To prevent all logging, use WithoutLogger.
func WithLogger(l Logger) Option {
	return &loggerOption{l}
}

// WithoutLogger returns an option the disables all logging from this package.
func WithoutLogger() Option {
	return &loggerOption{}
}

func (o *loggerOption) set(r *Runner) {
	if r.logf = o.logger.Infof; r.logf == nil {
		// a "silent" logger
		r.logf = func(string, ...interface{}) {}
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

// HookFunction is a function that may be registered using the Option provided
// by WithPreShutdownHook.
type HookFunction func(ctx context.Context) error

// WithPreShutdownHook registers a function to be executed just prior to server
// Shutdown. The Context passed to the HookFunction is the same one passed to
// Run and, if the HookFunction returns an error it is merely logged (if
// logging is enabled).  Otherwise, it will be ignored.
//
// Note: Only one HookFunction may be registered. If this Option is given
// multiple times, all but the final one will be ignored.
func WithPreShutdownHook(f HookFunction) Option {
	return hookFunction(f)
}

type hookFunction HookFunction

func (f hookFunction) set(r *Runner) {
	r.psHook = f
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func ErrServerClosedOK() Option {
	return new(escOK)
}

type escOK struct{}

func (e *escOK) set(r *Runner) {
	r.escOK = true
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
