package lameduck

type State int

const (
	Unknown State = iota
	NotStarted
	Running
	Failed
	Stopping
	Stopped
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

func (r *Runner) State() State {
	if r == nil || r.done == nil {
		return Unknown
	}

	return r.state
}
