

# toolman.org/net/lameduck
`import "toolman.org/net/lameduck"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a id="pkg-overview">Overview</a>

Package lameduck provides coordinated lame-duck behavior for any service
implementing this package's Server interface.

By default, lame-duck mode is triggered by receipt of SIGINT or SIGTERM
and the default lame-duck period is 3 seconds.  Options are provided to
alter these (an other) values.

This package is written assuming behavior similar to the standard library's
http.Server -- in that its Shutdown and Close methods exhibit behavior
matching the lameduck.Server interface -- however, in order to allow other
types to be used, a Serve method that returns nil is also needed.

For example:

	type LameDuckServer struct {
		 // This embedded http.Server provides Shutdown and Close methods
		 // with behavior expected by the lameduck.Server interface.
	  *http.Server
	}
	
	// Serve executes ListenAndServe in a manner compliant with the
	// lameduck.Server interface.
	func (s *LameDuckServer) Serve(context.Contxt) error {
	  err := s.Server.ListenAndServe()
	
	  if err == http.ErrServerClosed {
	    err = nil
	  }
	
	  return err
	}
	
	// Run will run the receiver's embedded http.Server and provide
	// lame-duck coverage on receipt of SIGTERM or SIGINT.
	func (s *LameDuckServer) Run(ctx context.Context) error {
	  return lameduck.Run(ctx, s)
	}




## <a id="pkg-index">Index</a>
* [func Run(ctx context.Context, svr Server, options ...Option) error](#Run)
* [type LameDuckError](#LameDuckError)
  * [func (lde *LameDuckError) Error() string](#LameDuckError.Error)
  * [func (lde *LameDuckError) Unwrap() error](#LameDuckError.Unwrap)
* [type Logger](#Logger)
* [type Option](#Option)
  * [func Period(p time.Duration) Option](#Period)
  * [func Signals(s ...os.Signal) Option](#Signals)
  * [func WithLogger(l Logger) Option](#WithLogger)
  * [func WithoutLogger() Option](#WithoutLogger)
* [type Server](#Server)


#### <a id="pkg-files">Package files</a>
[options.go](https://golang.org/src//target/options.go) [runner.go](https://golang.org/src//target/runner.go) [server.go](https://golang.org/src//target/server.go) [signals.go](https://golang.org/src//target/signals.go)

## <a id="Run">func</a> [Run](https://golang.org/src/./server.go?s=3375:3441#L85)
``` go
func Run(ctx context.Context, svr Server, options ...Option) error
```
Run executes the given Server providing coordinated lame-duck behavior on
reciept of one or more configurable signals. By default, the lame-duck
period is 3s and is triggered by SIGINT or SIGTERM. Options are available
to alter these values.


## <a id="LameDuckError">type</a> [LameDuckError](https://golang.org/src/./server.go?s=5324:5382#L173)
``` go
type LameDuckError struct {
    Expired bool
    Err     error
}

```
LameDuckError is the error type returned by Run for errors related to
lame-duck mode.


### <a id="LameDuckError.Error">func</a> (\*LameDuckError) [Error](https://golang.org/src/./server.go?s=5384:5424#L178)
``` go
func (lde *LameDuckError) Error() string
```

### <a id="LameDuckError.Unwrap">func</a> (\*LameDuckError) [Unwrap](https://golang.org/src/./server.go?s=5728:5768#L202)
``` go
func (lde *LameDuckError) Unwrap() error
```

## <a id="Logger">type</a> [Logger](https://golang.org/src/./options.go?s=853:909#L40)
``` go
type Logger interface {
    Infof(string, ...interface{})
}
```
Logger is the interface needed for the WithLogger Option.


## <a id="Option">type</a> [Option](https://golang.org/src/./options.go?s=171:210#L10)
``` go
type Option interface {
    // contains filtered or unexported methods
}
```
Option is the interface implemented by types that offer optional behavior
while running a Server with lame-duck support.


### <a id="Period">func</a> [Period](https://golang.org/src/./options.go?s=299:334#L16)
``` go
func Period(p time.Duration) Option
```
Period returns an Option that alters the lame-duck period to the given
Duration.


### <a id="Signals">func</a> [Signals](https://golang.org/src/./options.go?s=639:674#L29)
``` go
func Signals(s ...os.Signal) Option
```
Signals returns an Options that changes the list of Signals that trigger the
beginning of lame-duck mode. Using this Option fully replaces the previous
list of triggering signals.


### <a id="WithLogger">func</a> [WithLogger](https://golang.org/src/./options.go?s=1176:1208#L51)
``` go
func WithLogger(l Logger) Option
```
WithLogger returns an Option that alters this package's logging facility
to the provided Logger. Note, the default Logger is one derived from
'github.com/golang/glog'. To prevent all logging, use WithoutLogger.


### <a id="WithoutLogger">func</a> [WithoutLogger](https://golang.org/src/./options.go?s=1318:1345#L56)
``` go
func WithoutLogger() Option
```
WithoutLogger returns an option the disables all logging from this package.


## <a id="Server">type</a> [Server](https://golang.org/src/./server.go?s=1978:3119#L55)
``` go
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
```
Server defines the interface that should be implemented by types intended
for lame-duck support. It is expected that these methods exhibit behavior
similar to http.Server -- in that a call to Shutdown or Close should cause
Serve to return immediately.

However, unlike http.Server's Serve, ListenAndServe, and ListenAndServeTLS
methods (which return http.ErrServerClosed in this situation), this Serve
method should return a nil error when lame-duck mode is desired.


