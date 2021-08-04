// Package lameduck provides coordinated lame-duck behavior for any service
// implementing this package's Server interface.
//
// By default, lame-duck mode is triggered by receipt of SIGINT or SIGTERM
// and the default lame-duck period is 3 seconds.  Options are provided to
// alter these (an other) values.
//
// This package is written assuming behavior similar to the standard library's
// http.Server -- in that its Shutdown and Close methods exhibit behavior
// matching the lameduck.Server interface -- however, in order to allow other
// types to be used, a Serve method that returns nil is also needed.
//
//
//     type LameDuckServer struct {
//     	 // This embedded http.Server provides Shutdown and Close methods
//     	 // with behavior expected by the lameduck.Server interface.
//       *http.Server
//     }
//
//     // Serve executes ListenAndServe in a manner compliant with the
//     // lameduck.Server interface.
//     func (s *LameDuckServer) Serve(context.Contxt) error {
//       err := s.Server.ListenAndServe()
//
//       if err == http.ErrServerClosed {
//         err = nil
//       }
//
//       return err
//     }
//
//     // Run will run the receiver's embedded http.Server and provide
//     // lame-duck coverage on receipt of SIGTERM or SIGINT.
//     func (s *LameDuckServer) Run(ctx context.Context) error {
//       return lameduck.Run(ctx, s)
//     }
package lameduck

import (
	"context"
	"strings"

	"golang.org/x/sync/errgroup"
)

// Server defines the interface that should be implemented by types intended
// for lame-duck support. It is expected that these methods exhibit behavior
// similar to http.Server -- in that a call to Shutdown or Close should cause
// Serve to return immediately.
//
// However, unlike http.Server's Serve, ListenAndServe, and ListenAndServeTLS
// methods (which return http.ErrServerClosed in this situation), this Serve
// method should return a nil error when lame-duck mode is desired.
//
type Server interface {
	// Serve executes the Server. If Serve returns an error, that error will be
	// returned immediately by Run and no lame-duck coverage will be provided.
	//
	Serve(context.Context) error

	// Shutdown is called by Run (after catching one of the configured signals)
	// to initiate a graceful shutdown of the Server; this marks the beginning
	// of lame-duck mode. If Shutdown returns a nil error before the configured
	// lame-duck period has elapsed, Run will immediately return nil as well.
	//
	// The Context provided to Shutdown will have a timeout set to the configured
	// lame-duck Period. If Shutdown returns context.DeadlineExceeded, Run will
	// return a LameDuckError with its Expired field set to true and Err set to
	// the return value from calling Close.
	//
	// Any other error returned by Shutdown will be wrapped by a LameDuckError
	// with its Expired field set to false.
	Shutdown(context.Context) error

	// Close is called by Run when Shutdown returns context.DeadlineExceeded and
	// its return value will be assigned to the Err field of the LameDuckError
	// returned by Run.
	Close() error
}

// Run executes the given Server providing coordinated lame-duck behavior on
// reciept of one or more configurable signals. By default, the lame-duck
// period is 3s and is triggered by SIGINT or SIGTERM. Options are available
// to alter these values.
//
// Note that:
//
//    r, err := lameduck.Run(ctx, svr)
//
// ...is equivalent to calling:
//
// 		if r, err := lameduck.NewRunner(svr); err == nil {
// 			r.Run(ctx)
// 		}
//
func Run(ctx context.Context, svr Server, options ...Option) error {
	r, err := newRunner(svr, options)
	if err != nil {
		return err
	}

	return r.Run(ctx)
}

// Runner returns a lame-duck Runner that providing coordinated lame-duck
// behavior for the given svr.
//
// See the Run func for details.
func NewRunner(svr Server, options ...Option) (*Runner, error) {
	return newRunner(svr, options)
}

// Run executes the receiver's Server while providing coordinated lame-duck
// behavior on reciept of one or more configurable signals.
//
// See the Run func for details.
func (r *Runner) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Goroutine #1
	//
	//   - Waits for one of the configured signals
	//   - Calls Shutdown using a Context with a deadline for the configure period
	//   - If deadline is exceeded, returns the result of calling Close
	//   - Otherwise, returns the result from the call to Shutdown
	//   - On return, calls r.close()
	//
	eg.Go(func() error {
		defer r.close()

		r.logf("Waiting for signals: %v", r.signals)

		sig, err := r.waitForSignal(ctx)
		if err != nil {
			return err
		}

		r.logf("Received signal [%s]; entering lame-duck mode for %v", sig, r.period)

		ctx, cancel2 := context.WithTimeout(ctx, r.period)
		defer cancel2()

		err = r.server.Shutdown(ctx)
		switch err {
		case nil:
			r.logf("Completed lame-duck mode")
			return nil

		case context.DeadlineExceeded:
			r.logf("Lame-duck period has expired")
			return &LameDuckError{
				Expired: true,
				Err:     r.server.Close(),
			}

		default:
			r.logf("error shutting down server: %v", err)
			cancel()
			return &LameDuckError{Err: err}
		}
	})

	// Goroutine #2
	//
	//   - Calls Serve
	//   - If Server returns a non-nil error, return it immediately
	//   - Otherwise, wait for the Context or receiver to be "done".
	//
	eg.Go(func() error {
		r.logf("Starting server")
		close(r.ready)
		if err := r.server.Serve(ctx); err != nil {
			r.logf("Server failed: %v", err)
			return err
		}

		r.logf("Stopping server")

		select {
		case <-ctx.Done():
			r.logf("Context canceled wait for server shutdown")

		case <-r.done:
			r.logf("Server stopped")
		}

		return nil
	})

	return eg.Wait()
}

// LameDuckError is the error type returned by Run for errors related to
// lame-duck mode.
type LameDuckError struct {
	Expired bool
	Err     error
}

func (lde *LameDuckError) Error() string {
	if lde == nil {
		return ""
	}

	var msgs []string

	if lde.Expired {
		msgs = append(msgs, "Lame-duck period has expired")
	}

	if lde.Err != nil {
		if msg := lde.Err.Error(); msg != "" {
			msgs = append(msgs, msg)
		}
	}

	if len(msgs) == 0 {
		return ""
	}

	return strings.Join(msgs, " + ")
}

func (lde *LameDuckError) Unwrap() error {
	if lde == nil {
		return nil
	}
	return lde.Err
}
