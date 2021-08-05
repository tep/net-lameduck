
# lameduck [![Mit License][mit-img]][mit] [![GitHub Release][release-img]][release] [![GoDoc][godoc-img]][godoc] [![Go Report Card][reportcard-img]][reportcard]

`import toolman.org/net/lameduck`

Package `lameduck` provides coordinated lame-duck behavior for any service
implementing this package's Server interface.

    type Server interface {
      Serve(context.Context) error
      Shutdown(context.Context) error
      Close() error
    }

By default, lame-duck mode is triggered by receipt of SIGINT or SIGTERM
and the default lame-duck period is 3 seconds.  Options are provided to
alter these (and other) values.

This package is written assuming behavior similar to the standard library's
`http.Server` -- in that its `Shutdown` and `Close` methods exhibit behavior
matching the `lameduck.Server` interface -- however, in order to allow other
types to be used, a Serve method that returns nil is also needed.

Here's an example wrapper around `http.Server` that leverages this package:

    package mypkg

    import (
      "net/http"
      "toolman.org/net/lameduck"
    )

    type MyLameDuckServer struct {
       // This embedded http.Server provides Shutdown and Close methods
       // with behavior expected by the lameduck.Server interface.
      *http.Server
    }
    
    // Serve executes the embedded http.Server's ListenAndServe method in
    // a manner compliant with the lameduck package's Server interface.
    func (s *MyLameDuckServer) Serve(context.Contxt) error {
      err := s.ListenAndServe()
    
      if err == http.ErrServerClosed {
        err = nil
      }
    
      return err
    }
    
    // Run will run the receiver's embedded http.Server and provide
    // lame-duck coverage on receipt of SIGTERM or SIGINT.
    func (s *MyLameDuckServer) Run(ctx context.Context) error {
      return lameduck.Run(ctx, s)
    }

...and a simple `main` packages that uses it:

    package main

    import (
      "fmt"
      "net/http"

      "mypkg"
    )

    func main() {
      mlds := &mypkg.MyLameDuckServer{&http.Server{Addr: ":8080"}}

      if err := mlds.Run(context.Background()); err != nil {
        fmt.Printf("Server error: %v", err)
      }
    }

The above illustrates a simple wrapper around http.Server that may be started
using the provided Run method. This server will continue to run until receiving
a SIGINT or SIGTERM. On receipt of one of these signals, lameduck logic will
call its Server's Shutdown method which, in turn, will cause ListenAndServe
to return immediately -- as specified in the net/http documentation:

> When Shutdown is called, Serve, ListenAndServe, and ListenAndServeTLS
> immediately return ErrServerClosed. Make sure the program doesn't exit and
> waits instead for Shutdown to return.

Note however that the call to the underlying Server's Shutdown method indicates
that start of lame-duck processing which will continue until all in-flight
requests have completed or the provided Context is canceled:

> Shutdown gracefully shuts down the server without interrupting any active
> connections. Shutdown works by first closing all open listeners, then
> closing all idle connections, and then waiting indefinitely for connections
> to return to idle and then shut down. If the provided context expires before
> the shutdown is complete, Shutdown returns the context's error, otherwise it
> returns any error returned from closing the Server's underlying Listener(s).

This package ensures that a proper Context is provided to Shutdown having
a Deadline indicating the desired lame-duck grace period -- by default,
3 seconds -- then reacts accordingly based on the error returned by Shutdown.

If a non-nil error is returned from Run it will always be of type
LameDuckError:

    type LameDuckError struct {
      Expired bool
      Failed  bool
      Err     error
    }

If `Serve` returns an error, Run returns a `*LameDuckError` with `Failed` set to
true and Err set to that returned by Serve.

If the call to `Shutdown` returns nil, this indicates the the underlying Server
has stopped completely within the desired lame-duck grace period. In this case,
Run will return nil.  Otherwise, the error returned by `Shutdown` will follow
the below rules:

* If `Shutdown` returns `context.DeadlineExceeded` then Run's `*LameDuckError`
  will have a `true` value for `Expired` and its `Err` field will contain the
  error returned from the Server's `Close` method - if any.  Note also that the
  underlying Server's `Close` method is only called when `Shutdown` returns
  `context.DeadlineExceeded`.

* On the other hand, if `Shutdown` returns some other error, `Run` will return
  a `*LameDuckError` wrapping that error but both of its boolean fields will be
  false.


[mit-img]: http://img.shields.io/badge/License-MIT-c41e3a.svg
[mit]: https://github.com/tep/net-lameduck/blob/master/LICENSE

[release-img]: https://img.shields.io/github/release/tep/net-lameduck/all.svg
[release]: https://github.com/tep/net-lameduck/releases

[godoc-img]: https://pkg.go.dev/badge/toolman.org/net/lameduck.svg
[godoc]: https://pkg.go.dev/toolman.org/net/lameduck

[reportcard-img]: https://goreportcard.com/badge/toolman.org/net/lameduck
[reportcard]: https://goreportcard.com/report/toolman.org/net/lameduck

