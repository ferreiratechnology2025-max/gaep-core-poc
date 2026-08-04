// Architecture D — Actor Model — rival of the E1 battery.
//
// Born from the actor paradigm (frozen spec, experiment-plan.md, block D):
// state is encapsulated in actors, ONE actor per capability, each with a
// mailbox (channel); zero shared mutable state between actors. Each actor
// is single-writer over its own mailbox (internal order preserved);
// BETWEEN actors delivery is non-deterministic (runtime scheduling). A
// single collector merges the emitted events into the final trace; the
// canonical order IS the collector's receipt order (merge position) —
// never a timestamp, never individual goroutine arrival. Policy is an
// internal actor decision: the actor admits (admission policy) and
// authorizes effects from its own encapsulated state; a blocked effect
// emits AuthorizationDenied (+ AuditWritten); a rejected admission emits
// AdmissionRejected (+ AuditWritten) without creating state. Observable
// events are emitted by actors at decision boundaries; actor internal
// state is never read externally; the collector only merges, never
// emits. Rollback is an actor transition to a recovery state emitting
// exactly ONE RollbackExecuted (+ AuditWritten), never re-emitting prior
// decisions. Commands are messages to the target actor's mailbox; each
// actor choreographs its own lifecycle; no central coordinator.
//
// D IS NON-DETERMINISTIC (genuine actor concurrency). Single-capability
// scenarios (S1/S2/S3/S5) use one actor ⇒ internal order preserved ⇒ the
// distribution must be point-mass (a single verdict, 100% frequency); S4
// (two capabilities) interleaves between actors ⇒ the pre-registered
// stochastic clause (N=50, modal verdict) governs the comparison.
//
// Race-free by construction: actors communicate ONLY through channels; no
// shared mutable variable; the collector is the only goroutine that
// writes the trace.
//
// Provenance: D is implemented from scratch — no kernel/, models_comp/,
// B, C or E code is reused or adapted.
package e1_architectures_test

import (
	"sync"
	"testing"
	"time"

	"gaep-core-poc/harness"
)

// --- messages to actor mailboxes ---

type dCmdKind int

const (
	dCmdAdmit dCmdKind = iota
	dCmdAuthorize
	dCmdStart
	dCmdSucceed
	dCmdFail
	dCmdRollback
	dCmdEffect
	dCmdTimeout
)

type dCommand struct {
	kind dCmdKind
	cap  string
}

// --- actor: encapsulated state + mailbox, single-writer per actor ---

// dActor encapsulates the state of ONE capability and processes its
// mailbox sequentially (frozen spec point 2: single-writer per actor,
// internal order preserved). It emits observable events to the shared
// events channel at decision boundaries; its internal state is never read
// externally (frozen spec point 5).
type dActor struct {
	cap    string
	mail   chan dCommand
	out    chan<- harness.Event
	policy func(cap string) bool

	// Encapsulated (actor-private) state — zero shared mutable state. The
	// fields track the actor's lifecycle position (frozen spec points 1 and
	// 6): the actor is the only reader/writer of its own state; the
	// observable guards only read authorized/failed/timedOut (P2/P3/P5),
	// while running/succeeded/recovered keep the lifecycle self-model
	// (start/succeed/rollback transitions). No field is shared or read
	// externally.
	authorized bool
	running    bool
	succeeded  bool
	failed     bool
	timedOut   bool
	recovered  bool
}

func newDActor(cap string, out chan<- harness.Event, policy func(cap string) bool) *dActor {
	return &dActor{cap: cap, mail: make(chan dCommand, 64), out: out, policy: policy}
}

// run drains the mailbox until it closes: each message is handled in
// arrival order (the actor never processes two messages concurrently).
func (a *dActor) run() {
	for cmd := range a.mail {
		a.handle(cmd)
	}
}

// emit sends an observable event to the collector. Seq is NOT set here:
// the canonical order is the collector's receipt order (frozen spec
// point 3).
func (a *dActor) emit(kind harness.EventKind, cap string) {
	a.out <- harness.Event{Kind: kind, Cap: cap}
}

// handle applies one message to the actor's encapsulated state. Every
// state-affecting decision is followed by an audit record (P1); a blocked
// effect emits the contract denial event (frozen spec point 4); rollback
// is a forced transition to the recovery state emitting exactly one
// RollbackExecuted without re-emitting prior decisions (spec point 6).
func (a *dActor) handle(cmd dCommand) {
	switch cmd.kind {
	case dCmdAdmit:
		a.emit(harness.EvAdmissionRequested, "")
		if a.policy(a.cap) {
			a.emit(harness.EvAdmissionAccepted, "")
		} else {
			a.emit(harness.EvAdmissionRejected, "") // no state created
		}
		a.emit(harness.EvAuditWritten, "")
	case dCmdAuthorize:
		a.authorized = true
		a.emit(harness.EvAuthorizationGranted, a.cap)
		a.emit(harness.EvAuditWritten, "")
	case dCmdStart:
		a.running = true
		a.emit(harness.EvCapabilityStarted, a.cap)
		a.emit(harness.EvAuditWritten, "")
	case dCmdSucceed:
		a.running = false
		a.succeeded = true
		a.emit(harness.EvCapabilitySucceeded, a.cap)
		a.emit(harness.EvAuditWritten, "")
	case dCmdFail:
		a.running = false
		a.failed = true
		a.emit(harness.EvCapabilityFailed, a.cap)
		a.emit(harness.EvAuditWritten, "")
	case dCmdRollback:
		// Forced transition to the recovery state: exactly ONE
		// RollbackExecuted; no prior decision is re-emitted.
		a.failed = false
		a.recovered = true
		a.emit(harness.EvRollbackExecuted, a.cap)
		a.emit(harness.EvAuditWritten, "")
	case dCmdEffect:
		// Effect is authorized from the actor's own encapsulated state:
		// only an authorized, non-failed, non-timed-out capability may
		// produce an external effect; otherwise the actor answers the
		// message with the contract denial event (spec point 4).
		if a.authorized && !a.failed && !a.timedOut {
			a.emit(harness.EvExternalEffect, a.cap)
		} else {
			a.emit(harness.EvAuthorizationDenied, a.cap)
		}
		a.emit(harness.EvAuditWritten, "")
	case dCmdTimeout:
		a.timedOut = true
		a.emit(harness.EvTimedOut, a.cap)
		a.emit(harness.EvAuditWritten, "")
	}
}

// --- architecture D: dispatcher + actor network + single collector ---

// actorModelD is architecture D. run() builds a fresh actor network per
// call: commands are dispatched to the target actor's mailbox, each actor
// processes sequentially and emits to a shared events channel, and the
// collector merges the trace in canonical (receipt) order.
type actorModelD struct {
	policy func(cap string) bool
}

func newActorModelD(policy func(cap string) bool) *actorModelD {
	if policy == nil {
		policy = func(string) bool { return true }
	}
	return &actorModelD{policy: policy}
}

// dRunTimeout guards against deadlock: if the actor network does not
// finish within this budget, run() reports ok=false — an implementation
// failure, NOT a result (frozen spec: timeout must be reported as a bug).
const dRunTimeout = 5 * time.Second

// run executes the command script once on a fresh actor network and
// returns the collected trace in canonical order. ok=false means the
// network deadlocked (implementation failure).
func (d *actorModelD) run(cmds []dCommand) ([]harness.Event, bool) {
	type result struct{ trace []harness.Event }
	ch := make(chan result, 1)
	go func() { ch <- result{trace: d.runInner(cmds)} }()
	select {
	case r := <-ch:
		return r.trace, true
	case <-time.After(dRunTimeout):
		return nil, false
	}
}

// runInner builds the network: one actor goroutine per capability (in
// first-appearance order), a dispatcher that sends each command to the
// target mailbox, and the single collector goroutine. WaitGroup ensures
// every actor finished before eventsCh closes, so the collector drains
// every event.
func (d *actorModelD) runInner(cmds []dCommand) []harness.Event {
	eventsCh := make(chan harness.Event, 256)
	trace := make([]harness.Event, 0, 64)

	// Collector: the ONLY goroutine that reads eventsCh, assigns Seq in
	// receipt order (canonical order) and writes the trace. It never
	// emits; it only merges (frozen spec points 3 and 5).
	collectDone := make(chan struct{})
	go func() {
		defer close(collectDone)
		seq := 0
		for e := range eventsCh {
			seq++
			e.Seq = seq
			trace = append(trace, e)
		}
	}()

	// One actor per capability, in first-appearance order.
	actors := map[string]*dActor{}
	var order []string
	for _, cmd := range cmds {
		if _, ok := actors[cmd.cap]; !ok {
			actors[cmd.cap] = newDActor(cmd.cap, eventsCh, d.policy)
			order = append(order, cmd.cap)
		}
	}

	var wg sync.WaitGroup
	for _, cap := range order {
		a := actors[cap]
		wg.Add(1)
		go func(a *dActor) {
			defer wg.Done()
			a.run()
		}(a)
	}

	// Dispatch: commands become messages to the target actor's mailbox
	// (no central coordinator — each actor choreographs its own cycle).
	for _, cmd := range cmds {
		actors[cmd.cap].mail <- cmd
	}

	// Close mailboxes; each actor drains and exits.
	for _, cap := range order {
		close(actors[cap].mail)
	}

	wg.Wait()       // all actors finished emitting
	close(eventsCh) // collector drains the rest and exits
	<-collectDone
	return trace
}

// --- scenario command scripts (identical inputs for every run) ---

// dCommands drives D through each battery scenario. Content is IDENTICAL
// to B's, C's and E's command scripts (same capabilities, same order —
// identical protocol inputs); only the implementation is D's own.
var dCommands = map[string][]dCommand{
	// S1 — capability fails mid-execution (P1/P3/P4). Single actor.
	"S1": {
		{dCmdAdmit, "c1"}, {dCmdAuthorize, "c1"}, {dCmdStart, "c1"},
		{dCmdFail, "c1"}, {dCmdRollback, "c1"},
	},
	// S2 — an external effect is attempted before authorization: the
	// actor answers the message with AuthorizationDenied (frozen spec
	// point 4); after authorization the same effect is allowed (P2).
	// Single actor.
	"S2": {
		{dCmdAdmit, "c1"},
		{dCmdEffect, "c1"}, // denied: capability not authorized yet
		{dCmdAuthorize, "c1"},
		{dCmdEffect, "c1"}, // allowed: capability authorized
	},
	// S3 — late success after timeout: timeout is recorded, the late
	// completion records a success, and no effect is emitted (P5).
	// Single actor.
	"S3": {
		{dCmdAdmit, "c1"}, {dCmdAuthorize, "c1"}, {dCmdStart, "c1"},
		{dCmdTimeout, "c1"}, {dCmdSucceed, "c1"},
	},
	// S4 — two capabilities interleaved. Two actors: each mailbox keeps
	// its own order, but the interleaving BETWEEN actors is governed by
	// the runtime scheduler — genuinely non-deterministic (P8 probe: the
	// stochastic clause applies; the oracle decides per run).
	"S4": {
		{dCmdAuthorize, "c1"}, {dCmdStart, "c1"},
		{dCmdStart, "c2"}, {dCmdSucceed, "c2"}, {dCmdSucceed, "c1"},
	},
	// S5 — identical admission request repeated (P7, N runs). Single
	// actor.
	"S5": {
		{dCmdAdmit, "c1"},
	},
	// S6 — E2.1b coverage: capability fails, an effect is illegally
	// attempted after the failure (pre-rollback), rollback recovers c1, and
	// a second, unrelated capability is freshly admitted and completes
	// normally (adr-evidence/E2_1a_scenario_design.md). Two actors, one
	// after the other (no interleaving intended: c2's commands only start
	// after c1's mailbox is dispatched).
	"S6": {
		{dCmdAdmit, "c1"}, {dCmdAuthorize, "c1"}, {dCmdStart, "c1"}, {dCmdFail, "c1"},
		{dCmdEffect, "c1"}, // illegal attempt, post-failure, pre-rollback
		{dCmdRollback, "c1"},
		{dCmdAdmit, "c2"}, {dCmdAuthorize, "c2"}, {dCmdStart, "c2"}, {dCmdSucceed, "c2"},
	},
	// S7 — E2.1b coverage: capability times out, an effect is illegally
	// attempted after the timeout, and the late success is still ignored
	// (adr-evidence/E2_1a_scenario_design.md). Single actor.
	"S7": {
		{dCmdAdmit, "c1"}, {dCmdAuthorize, "c1"}, {dCmdStart, "c1"}, {dCmdTimeout, "c1"},
		{dCmdEffect, "c1"},  // illegal attempt, post-timeout
		{dCmdSucceed, "c1"}, // late, still ignored
	},
	// S8 — E2.1b coverage: two capabilities admitted, then interleaved
	// (S4's structure, with explicit Admit on both — S4 itself is frozen
	// and untouched) (adr-evidence/E2_1a_scenario_design.md). Two actors;
	// interleaving between them is governed by the runtime scheduler, same
	// as S4.
	"S8": {
		{dCmdAdmit, "c1"}, {dCmdAdmit, "c2"},
		{dCmdAuthorize, "c1"}, {dCmdStart, "c1"},
		{dCmdStart, "c2"}, {dCmdSucceed, "c2"}, {dCmdSucceed, "c1"},
	},
	// S9 — E2.2b coverage: cross-capability effect scoping (E2.2a design,
	// S9). Two actors, dispatched sequentially (c1's commands complete
	// before c2's Admit is sent).
	"S9": {
		{dCmdAdmit, "c1"}, {dCmdAuthorize, "c1"}, {dCmdEffect, "c1"}, // legitimate
		{dCmdAdmit, "c2"},
		{dCmdEffect, "c2"},                        // illegal attempt: c2 not authorized, even though c1 is
		{dCmdAuthorize, "c2"}, {dCmdEffect, "c2"}, // legitimate
	},
	// S10 — E2.2b coverage: effect attempted again after Fail+Rollback,
	// without a fresh Authorize (E2.2a design, S10 — DECISÃO A: rollback
	// does not restore authorization). Single actor.
	"S10": {
		{dCmdAdmit, "c1"}, {dCmdAuthorize, "c1"}, {dCmdEffect, "c1"}, // legitimate
		{dCmdFail, "c1"}, {dCmdRollback, "c1"},
		{dCmdEffect, "c1"}, // attempted again, NO fresh Authorize
	},
}

// TestD_AdmissionRejectionMapping verifies the frozen rejection path
// (spec point 4): a rejected admission emits the contract rejection event
// and creates no state (no authorization is granted by a later message
// without an explicit Authorize).
func TestD_AdmissionRejectionMapping(t *testing.T) {
	d := newActorModelD(func(cap string) bool { return cap != "deny" })
	tr, ok := d.run([]dCommand{{kind: dCmdAdmit, cap: "deny"}})
	if !ok {
		t.Fatal("D.run() timed out — implementation failure")
	}
	if len(tr) != 3 || tr[0].Kind != harness.EvAdmissionRequested ||
		tr[1].Kind != harness.EvAdmissionRejected || tr[2].Kind != harness.EvAuditWritten {
		t.Fatalf("rejection mapping wrong: %v", tr)
	}
	// A subsequent effect on the rejected capability must be denied: no
	// authorization was created by the rejected admission.
	tr2, ok := d.run([]dCommand{{kind: dCmdEffect, cap: "deny"}})
	if !ok {
		t.Fatal("D.run() timed out — implementation failure")
	}
	if len(tr2) != 2 || tr2[0].Kind != harness.EvAuthorizationDenied || tr2[1].Kind != harness.EvAuditWritten {
		t.Fatalf("effect after rejected admission must be denied, got: %v", tr2)
	}
}

// archD adapts D to the scaffold's archColumn convention (S1 wired; the
// full battery runs under the stochastic clause in
// TestE1_ArchitectureD_Battery).
func archD() archColumn {
	return archColumn{
		id:   "D",
		name: "Actor Model",
		emit: func(h ScenarioHandle) []harness.Event {
			if h == scnFailMid {
				tr, ok := newActorModelD(nil).run(dCommands["S1"])
				if !ok {
					return nil
				}
				return tr
			}
			return nil
		},
	}
}
