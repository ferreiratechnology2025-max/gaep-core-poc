// Architecture E — Hierarchical State Machine — rival of the E1 battery.
//
// Born from the state-machine paradigm (frozen spec, experiment-plan.md,
// block E): the state of a capability is the position of its own machine;
// the system state is the collection of per-capability positions; zero
// shared mutable cells/registers. A single dispatcher serializes every
// transition attempt; the canonical order IS the dispatch position (never
// a timestamp, goroutine arrival or scheduler order). Policy is applied
// as transition guards (admission policy; effect requires an authorized,
// non-failed, non-timed-out position): a blocked transition does not
// fire, does not enter the target state, and emits NO observable event.
// Observable events are outputs of fired transitions — internal reads of
// machine position are never observable. Rollback is a forced transition
// to a recovery position emitting exactly ONE RollbackExecuted, never
// re-emitting prior transition outputs. The transition function is pure:
// transition(position, input, guard) → (newPosition, events); same input
// sequence ⇒ same trajectory ⇒ same events (no stochastic clause).
//
// Provenance: E is implemented from scratch — no kernel/, models_comp/ or
// C (c_eventsourcing_test.go) code is reused or adapted.
package e1_architectures_test

import (
	"sync"

	"gaep-core-poc/harness"
)

// eState is a position in the per-capability machine. The model supports
// superstate/substate hierarchy (a superstate can aggregate positions); the
// battery exercises the flat lifecycle, so the exercised positions are the
// flat ones below.
type eState int

const (
	eNone eState = iota
	eAdmitted
	eAuthorized
	eRunning
	eSucceeded
	eFailed
	eRecovered
	eTimedOut
)

// eInput is one transition trigger (scenario command, same content as C).
type eInput int

const (
	eInAdmit eInput = iota
	eInAuthorize
	eInStart
	eInSucceed
	eInFail
	eInRollback
	eInEffect
	eInTimeout
)

type eCommand struct {
	in  eInput
	cap string
}

// eTransition is the pure transition function output: fired=false means
// the guard blocked the transition — no state change, no observable event.
type eTransition struct {
	fired  bool
	to     eState
	events []harness.EventKind
}

// eTransitionFunc is pure: transition(position, input, guardOK) →
// (newPosition, events). Same inputs ⇒ same result.
func eTransitionFunc(s eState, in eInput, guardOK bool) eTransition {
	switch in {
	case eInAdmit:
		if !guardOK {
			return eTransition{fired: false, to: s}
		}
		return eTransition{fired: true, to: eAdmitted,
			events: []harness.EventKind{harness.EvAdmissionRequested, harness.EvAdmissionAccepted, harness.EvAuditWritten}}
	case eInAuthorize:
		return eTransition{fired: true, to: eAuthorized,
			events: []harness.EventKind{harness.EvAuthorizationGranted, harness.EvAuditWritten}}
	case eInStart:
		return eTransition{fired: true, to: eRunning,
			events: []harness.EventKind{harness.EvCapabilityStarted, harness.EvAuditWritten}}
	case eInSucceed:
		return eTransition{fired: true, to: eSucceeded,
			events: []harness.EventKind{harness.EvCapabilitySucceeded, harness.EvAuditWritten}}
	case eInFail:
		return eTransition{fired: true, to: eFailed,
			events: []harness.EventKind{harness.EvCapabilityFailed, harness.EvAuditWritten}}
	case eInRollback:
		// Forced transition to the recovery position: exactly ONE
		// RollbackExecuted, never re-emitting prior transition outputs.
		return eTransition{fired: true, to: eRecovered,
			events: []harness.EventKind{harness.EvRollbackExecuted, harness.EvAuditWritten}}
	case eInEffect:
		if !guardOK {
			return eTransition{fired: false, to: s}
		}
		// Effect is a self-loop: position unchanged, output emitted.
		return eTransition{fired: true, to: s,
			events: []harness.EventKind{harness.EvExternalEffect, harness.EvAuditWritten}}
	case eInTimeout:
		return eTransition{fired: true, to: eTimedOut,
			events: []harness.EventKind{harness.EvTimedOut, harness.EvAuditWritten}}
	}
	return eTransition{fired: false, to: s}
}

// eCapFor maps a transition output to the capability it refers to.
// Admission and audit are GLOBAL events (Cap "") per the observable
// contract (events.go: "global events such as admission or audit").
func eCapFor(kind harness.EventKind, cap string) string {
	switch kind {
	case harness.EvAdmissionRequested, harness.EvAdmissionAccepted,
		harness.EvAdmissionRejected, harness.EvAuditWritten:
		return ""
	}
	return cap
}

// eStateMachine is architecture E. Single-writer: run() holds the mutex for
// the whole command batch, so every transition attempt is dispatched at one
// point; machines are per-capability and no mutable transition state is
// shared.
type eStateMachine struct {
	mu     sync.Mutex
	mach   map[string]eState // per-capability machine position
	policy func(cap string) bool
	events []harness.Event
	seq    int
}

func newEStateMachine(policy func(cap string) bool) *eStateMachine {
	if policy == nil {
		policy = func(string) bool { return true }
	}
	return &eStateMachine{policy: policy, mach: map[string]eState{}}
}

// run dispatches cmds in canonical (dispatch) order and returns the emitted
// observable events (a fresh copy). Deterministic: same cmds + same policy
// ⇒ same trajectory ⇒ same events.
func (e *eStateMachine) run(cmds []eCommand) []harness.Event {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.mach = map[string]eState{}
	e.events = nil
	e.seq = 0
	for _, cmd := range cmds {
		e.dispatch(cmd)
	}
	out := make([]harness.Event, len(e.events))
	copy(out, e.events)
	return out
}

func (e *eStateMachine) emit(kind harness.EventKind, cap string) {
	e.seq++
	e.events = append(e.events, harness.Event{Seq: e.seq, Kind: kind, Cap: cap})
}

// guard returns the policy guard for the input on the given position:
// admission policy on eInAdmit; effect requires an authorized, non-failed,
// non-timed-out position (P2/P3/P5 enforcement).
func (e *eStateMachine) guard(in eInput, s eState, cap string) bool {
	switch in {
	case eInAdmit:
		return e.policy(cap)
	case eInEffect:
		return s == eAuthorized || s == eRunning
	}
	return true
}

func (e *eStateMachine) dispatch(cmd eCommand) {
	s := e.mach[cmd.cap]
	tr := eTransitionFunc(s, cmd.in, e.guard(cmd.in, s, cmd.cap))
	if !tr.fired {
		return // blocked: no event, no state change (frozen spec point 4)
	}
	e.mach[cmd.cap] = tr.to
	for _, k := range tr.events {
		e.emit(k, eCapFor(k, cmd.cap))
	}
}

// eCommands drives E through each battery scenario. Content is IDENTICAL to
// C's command scripts (same capabilities, same order — identical protocol
// inputs); only the implementation is E's own.
var eCommands = map[string][]eCommand{
	// S1 — capability fails mid-execution (P1/P3/P4).
	"S1": {
		{eInAdmit, "c1"}, {eInAuthorize, "c1"}, {eInStart, "c1"},
		{eInFail, "c1"}, {eInRollback, "c1"},
	},
	// S2 — an external effect is attempted before authorization: the
	// effect guard blocks it (position eAdmitted, not authorized); after
	// authorization the same effect fires (P2).
	"S2": {
		{eInAdmit, "c1"},
		{eInEffect, "c1"}, // blocked: guard (position not authorized)
		{eInAuthorize, "c1"},
		{eInEffect, "c1"}, // fires: position authorized
	},
	// S3 — late success after timeout: timeout fires (eTimedOut), the late
	// completion records a transition to eSucceeded, and no effect is
	// emitted (P5).
	"S3": {
		{eInAdmit, "c1"}, {eInAuthorize, "c1"}, {eInStart, "c1"},
		{eInTimeout, "c1"}, {eInSucceed, "c1"},
	},
	// S4 — two capabilities interleaved: E has no mechanism in its frozen
	// spec to reorder or suppress the interleaving, so the dispatch order
	// stays observable (P8 probe: the oracle decides).
	"S4": {
		{eInAuthorize, "c1"}, {eInStart, "c1"},
		{eInStart, "c2"}, {eInSucceed, "c2"}, {eInSucceed, "c1"},
	},
	// S5 — identical admission request repeated (P7, two runs).
	"S5": {
		{eInAdmit, "c1"},
	},
}

// archE adapts E to the scaffold's archColumn convention (S1 wired; the
// full battery runs in TestE1_ArchitectureE_Battery).
func archE() archColumn {
	return archColumn{
		id:   "E",
		name: "Hierarchical State Machine",
		emit: func(h ScenarioHandle) []harness.Event {
			if h == scnFailMid {
				return newEStateMachine(nil).run(eCommands["S1"])
			}
			return nil
		},
	}
}
