package lameduck

// State represents the lame-duck runtime state for a Server.
type State int

const (
	Unknown    State = iota // Zero value; state is unknown
	NotStarted              // The Server has not been started
	Running                 // The Server is currently running without incident
	Failed                  // The Server failed to start
	Stopping                // The Server is in the process of stopping
	Stopped                 // The Server has been stopped.
)

func (s State) String() string {
	switch s {
	case NotStarted:
		return "NOT_STARTED"
	case Running:
		return "RUNNING"
	case Failed:
		return "FAILED"
	case Stopped:
		return "STOPPED"
	case Stopping:
		return "STOPPING"
	default:
		return "UNKNOWN"
	}
}

// State returns the current runtime State for the receiver.
func (r *Runner) State() State {
	if r == nil || r.done == nil {
		return Unknown
	}

	return r.state
}
