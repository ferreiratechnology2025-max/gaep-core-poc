package models_comp

import "fmt"

// LifecycleState enumerates the engine control-flow states.
type LifecycleState int

const (
	StateIDLE LifecycleState = iota
	StateRunning
	StateWaitOutcome
	StateRelease
)

func (s LifecycleState) String() string {
	switch s {
	case StateIDLE:
		return "IDLE"
	case StateRunning:
		return "RUNNING"
	case StateWaitOutcome:
		return "WAIT_OUTCOME"
	case StateRelease:
		return "RELEASE"
	}
	return "?"
}

// LifecycleEvent is one control-flow input to the automaton.
type LifecycleEvent int

const (
	EvSession LifecycleEvent = iota // start a session (requires session id)
	EvWait                          // RUNNING → WAIT_OUTCOME
	EvOutcome                       // WAIT_OUTCOME → RUNNING
	EvRelease                       // RUNNING/WAIT_OUTCOME → RELEASE
	EvFail                          // RUNNING → RELEASE
	EvTimeout                       // WAIT_OUTCOME → RELEASE (watchdog fired)
)

// LifecycleSample is a point-in-time projection used by invariant checks.
type LifecycleSample struct {
	State    LifecycleState
	Session  uint64
	Watchdog uint64
}

// Lifecycle is a minimal governance state machine used by experiment E0:
//
//	IDLE --session--> RUNNING --wait--> WAIT_OUTCOME --outcome--> RUNNING
//	RUNNING --release|fail--> RELEASE
//	WAIT_OUTCOME --timeout(watchdog)|release--> RELEASE
//
// The No* fields remove individual components so that E0 can observe which
// lifecycle invariants stop being guaranteed (or become unverifiable).
type Lifecycle struct {
	State    LifecycleState
	Session  uint64 // 0 = no active session
	Watchdog uint64 // deadline as unix nanos; 0 = disarmed

	NoSessionGate bool // SESSION removed: session checks disabled
	NoWatchdog    bool // WATCHDOG removed: timeout path disabled
	NoWaitOutcome bool // WAIT_OUTCOME removed: wait/outcome disabled
	NoRelease     bool // RELEASE removed: release/fail/timeout delivery disabled
}

// NewLifecycle returns an IDLE automaton.
func NewLifecycle() *Lifecycle { return &Lifecycle{} }

// Sample projects the current state.
func (l *Lifecycle) Sample() LifecycleSample {
	return LifecycleSample{State: l.State, Session: l.Session, Watchdog: l.Watchdog}
}

// Step applies one event; it returns an error when the transition is not
// permitted in the current state (or its component was removed).
func (l *Lifecycle) Step(ev LifecycleEvent) error {
	switch ev {
	case EvSession:
		if l.State != StateIDLE {
			return fmt.Errorf("session only from IDLE")
		}
		if !l.NoSessionGate && l.Session == 0 {
			return fmt.Errorf("session requires a non-zero session id")
		}
		l.State = StateRunning
		return nil
	case EvWait:
		if l.NoWaitOutcome {
			return fmt.Errorf("WAIT_OUTCOME removed")
		}
		if l.State != StateRunning {
			return fmt.Errorf("wait only from RUNNING")
		}
		if !l.NoSessionGate && l.Session == 0 {
			return fmt.Errorf("wait requires an active session")
		}
		l.State = StateWaitOutcome
		return nil
	case EvOutcome:
		if l.NoWaitOutcome {
			return fmt.Errorf("WAIT_OUTCOME removed")
		}
		if l.State != StateWaitOutcome {
			return fmt.Errorf("outcome only from WAIT_OUTCOME")
		}
		l.State = StateRunning
		return nil
	case EvTimeout:
		if l.NoWatchdog {
			return fmt.Errorf("WATCHDOG removed")
		}
		if l.NoRelease {
			return fmt.Errorf("RELEASE removed")
		}
		if l.State != StateWaitOutcome {
			return fmt.Errorf("timeout only from WAIT_OUTCOME")
		}
		if l.Watchdog == 0 {
			return fmt.Errorf("watchdog disarmed")
		}
		l.State = StateRelease
		return nil
	case EvRelease, EvFail:
		if l.NoRelease {
			return fmt.Errorf("RELEASE removed")
		}
		if l.State == StateRelease {
			return fmt.Errorf("RELEASE is absorbing")
		}
		if l.State != StateRunning && l.State != StateWaitOutcome {
			return fmt.Errorf("release only from RUNNING or WAIT_OUTCOME")
		}
		l.State = StateRelease
		return nil
	}
	return fmt.Errorf("unknown event %d", ev)
}
