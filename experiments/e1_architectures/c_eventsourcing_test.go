// Architecture C — Event Sourcing — rival of the E1 battery.
//
// Born from the event-sourcing paradigm (frozen spec, experiment-plan.md,
// block C): state is a pure projection of an append-only log; the log is
// the source of truth; there are zero mutable cells/registers. A single
// writer serializes every append; the canonical order IS the log index
// (never a timestamp, never goroutine arrival, never the scheduler).
// Policy runs only at admission: admitted ⇒ append + observable event;
// rejected ⇒ contract rejection event, no append. Observable events
// represent persisted facts / admission results — the projection NEVER
// emits. Rollback = replay of the log prefix: reconstructs only the
// projection and emits exactly one RollbackExecuted, never the replayed
// events. C is deterministic: same commands ⇒ same log ⇒ same projection ⇒
// same trace (no stochastic clause is triggered).
//
// Provenance: C is implemented from scratch — no kernel/ (register file of
// A) or models_comp/ code is reused or adapted.
package e1_architectures_test

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"gaep-core-poc/harness"
)

// --- C internals: append-only log (source of truth) ---

type cFactKind int

const (
	cFactAdmission     cFactKind = iota // admission accepted
	cFactAuthorization                  // authorization granted
	cFactStart                          // capability started
	cFactSucceed                        // capability succeeded
	cFactFail                           // capability failed
	cFactEffect                         // external effect produced
	cFactTimeout                        // execution timed out
)

// cFact is one persisted fact. idx is the append position — the canonical
// sequence identity (never a timestamp or scheduler order).
type cFact struct {
	idx  int
	kind cFactKind
	cap  string
}

// checkCanonicalOrder asserts spec point 3: the sequence identity of a
// fact IS its log index (append position) — never a timestamp or scheduler
// order. Returns an error description, or "" when satisfied.
func (l *cLog) checkCanonicalOrder() string {
	for i, f := range l.facts {
		if f.idx != i {
			return fmt.Sprintf("fact %d has idx %d (want %d)", i, f.idx, i)
		}
	}
	return ""
}

// cLog is the append-only source of truth.
type cLog struct{ facts []cFact }

func (l *cLog) append(kind cFactKind, cap string) {
	l.facts = append(l.facts, cFact{idx: len(l.facts), kind: kind, cap: cap})
}

// --- projection: pure function of the log (never emits) ---

type cCapState int

const (
	cCapNone cCapState = iota
	cCapAdmitted
	cCapAuthorized
	cCapRunning
	cCapSucceeded
	cCapFailed
)

type cProjection struct {
	state    map[string]cCapState
	timedOut map[string]bool
}

// project derives the read model purely from the log. Same log ⇒ same
// projection. It is lazy and silent: computing it emits no observable
// event and appends no fact.
func project(l *cLog) cProjection {
	p := cProjection{state: map[string]cCapState{}, timedOut: map[string]bool{}}
	for _, f := range l.facts {
		switch f.kind {
		case cFactAdmission:
			p.state[f.cap] = cCapAdmitted
		case cFactAuthorization:
			p.state[f.cap] = cCapAuthorized
		case cFactStart:
			p.state[f.cap] = cCapRunning
		case cFactSucceed:
			p.state[f.cap] = cCapSucceeded
		case cFactFail:
			p.state[f.cap] = cCapFailed
		case cFactTimeout:
			p.timedOut[f.cap] = true
		}
	}
	return p
}

// effectAllowed governs whether an external effect is observable: only for
// an authorized (or running) capability that has not timed out.
func (p cProjection) effectAllowed(cap string) bool {
	if p.timedOut[cap] {
		return false
	}
	switch p.state[cap] {
	case cCapAuthorized, cCapRunning:
		return true
	}
	return false
}

// --- command processor (single writer) ---

type cCmdKind int

const (
	cCmdAdmit cCmdKind = iota
	cCmdAuthorize
	cCmdStart
	cCmdSucceed
	cCmdFail
	cCmdRollback
	cCmdEffect
	cCmdTimeout
)

type cCommand struct {
	kind cCmdKind
	cap  string
}

// eventSourcingC is architecture C. Single-writer: run() holds the mutex
// for the whole command batch, so every append is serialized at one point.
type eventSourcingC struct {
	mu     sync.Mutex
	log    cLog
	policy func(cap string) bool // admission policy (pure; nil ⇒ admit all)
	events []harness.Event
	seq    int
}

func newEventSourcingC(policy func(cap string) bool) *eventSourcingC {
	if policy == nil {
		policy = func(string) bool { return true }
	}
	return &eventSourcingC{policy: policy}
}

// run processes cmds in canonical order and returns the emitted observable
// events (a fresh copy). Deterministic: same cmds + same policy ⇒ same
// log ⇒ same projection ⇒ same trace.
func (c *eventSourcingC) run(cmds []cCommand) []harness.Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.log = cLog{}
	c.events = nil
	c.seq = 0
	for _, cmd := range cmds {
		c.handle(cmd)
	}
	out := make([]harness.Event, len(c.events))
	copy(out, c.events)
	return out
}

func (c *eventSourcingC) emit(kind harness.EventKind, cap string) {
	c.seq++
	c.events = append(c.events, harness.Event{Seq: c.seq, Kind: kind, Cap: cap})
}

func (c *eventSourcingC) handle(cmd cCommand) {
	switch cmd.kind {
	case cCmdAdmit:
		c.emit(harness.EvAdmissionRequested, "")
		if c.policy(cmd.cap) {
			c.log.append(cFactAdmission, cmd.cap) // admitted ⇒ append
			c.emit(harness.EvAdmissionAccepted, "")
		} else {
			c.emit(harness.EvAdmissionRejected, "") // rejected ⇒ no append
		}
		c.emit(harness.EvAuditWritten, "") // P1: state-affecting ⇒ audited
	case cCmdAuthorize:
		c.log.append(cFactAuthorization, cmd.cap)
		c.emit(harness.EvAuthorizationGranted, cmd.cap)
		c.emit(harness.EvAuditWritten, "")
	case cCmdStart:
		c.log.append(cFactStart, cmd.cap)
		c.emit(harness.EvCapabilityStarted, cmd.cap)
		c.emit(harness.EvAuditWritten, "")
	case cCmdSucceed:
		c.log.append(cFactSucceed, cmd.cap)
		c.emit(harness.EvCapabilitySucceeded, cmd.cap)
		c.emit(harness.EvAuditWritten, "")
	case cCmdFail:
		c.log.append(cFactFail, cmd.cap)
		c.emit(harness.EvCapabilityFailed, cmd.cap)
		c.emit(harness.EvAuditWritten, "")
	case cCmdRollback:
		// Rollback = replay of the log prefix: reconstruct the projection
		// from the prefix before the failed fact. No replayed event is
		// re-emitted; exactly one RollbackExecuted follows.
		c.rollback(cmd.cap)
		c.emit(harness.EvRollbackExecuted, cmd.cap)
		c.emit(harness.EvAuditWritten, "")
	case cCmdEffect:
		if project(&c.log).effectAllowed(cmd.cap) {
			c.log.append(cFactEffect, cmd.cap)
			c.emit(harness.EvExternalEffect, cmd.cap)
			c.emit(harness.EvAuditWritten, "")
		}
		// suppressed otherwise: the violation never becomes observable
	case cCmdTimeout:
		c.log.append(cFactTimeout, cmd.cap)
		c.emit(harness.EvTimedOut, cmd.cap)
		c.emit(harness.EvAuditWritten, "")
	}
}

// rollback truncates the log to the prefix before the failed fact of cap:
// the projection re-derives from that prefix (replay semantics). The
// failure and anything after it leave the source of truth.
func (c *eventSourcingC) rollback(cap string) {
	cut := -1
	for i, f := range c.log.facts {
		if f.kind == cFactFail && f.cap == cap {
			cut = i
		}
	}
	if cut >= 0 {
		c.log.facts = c.log.facts[:cut]
	}
}

// --- scenario command scripts (identical inputs for every run) ---

// cCommands drives C through each battery scenario. C's governance decides
// the observable emission; the same scripts are the protocol inputs.
var cCommands = map[string][]cCommand{
	// S1 — capability fails mid-execution (P1/P3/P4).
	"S1": {
		{cCmdAdmit, "c1"}, {cCmdAuthorize, "c1"}, {cCmdStart, "c1"},
		{cCmdFail, "c1"}, {cCmdRollback, "c1"},
	},
	// S2 — an external effect is attempted before authorization: C
	// suppresses it (projection: not authorized); after authorization the
	// same effect is allowed and becomes observable (P2).
	"S2": {
		{cCmdAdmit, "c1"},
		{cCmdEffect, "c1"}, // attempted pre-auth → suppressed
		{cCmdAuthorize, "c1"},
		{cCmdEffect, "c1"}, // now authorized → emitted
	},
	// S3 — late success after timeout: completion is recorded, but the
	// timed-out execution never produces an external effect (P5).
	"S3": {
		{cCmdAdmit, "c1"}, {cCmdAuthorize, "c1"}, {cCmdStart, "c1"},
		{cCmdTimeout, "c1"}, {cCmdSucceed, "c1"},
	},
	// S4 — two capabilities interleaved: C has no per-capability isolation
	// mechanism in its frozen spec, so the interleaving stays observable
	// (P8 probe: the oracle decides).
	"S4": {
		{cCmdAuthorize, "c1"}, {cCmdStart, "c1"},
		{cCmdStart, "c2"}, {cCmdSucceed, "c2"}, {cCmdSucceed, "c1"},
	},
	// S5 — identical admission request repeated (P7, two runs).
	"S5": {
		{cCmdAdmit, "c1"},
	},
}

// archC adapts C to the scaffold's archColumn convention (S1 wired; the
// full battery runs in TestE1_ArchitectureC_Battery).
func archC() archColumn {
	return archColumn{
		id:   "C",
		name: "Event Sourcing",
		emit: func(h ScenarioHandle) []harness.Event {
			if h == scnFailMid {
				return newEventSourcingC(nil).run(cCommands["S1"])
			}
			return nil
		},
	}
}

// TestE1_ArchitectureC_Battery runs C through S1–S5 via the Property
// Oracle, logs the P1–P8 × S1–S5 matrix and asserts the expected verdicts.
func TestE1_ArchitectureC_Battery(t *testing.T) {
	o := harness.NewOracle()
	props := []harness.PropertyID{
		harness.P1Auditability, harness.P2Authorization, harness.P3FailureBlocks,
		harness.P4Recoverable, harness.P5AsyncTolerant, harness.P6DomainFree,
		harness.P7Determinism, harness.P8Isolation,
	}

	t.Logf("E1 — architecture C (Event Sourcing) × battery S1–S5 (Property Oracle)")
	t.Logf("Deterministic by construction: same commands ⇒ same log ⇒ same projection ⇒ same trace.")

	traces := map[string][]harness.Event{}
	verdicts := map[string]map[harness.PropertyID]harness.Verdict{}
	for _, s := range harness.Battery {
		cmds, ok := cCommands[s.ID]
		if !ok {
			t.Fatalf("no command script for scenario %s", s.ID)
		}
		r1 := newEventSourcingC(nil).run(cmds)
		r2 := newEventSourcingC(nil).run(cmds)
		if !reflect.DeepEqual(r1, r2) {
			t.Errorf("%s: C is not deterministic (run 1 != run 2)", s.ID)
		}
		traces[s.ID] = r1
		if s.ID == "S5" {
			// P7 needs repeated runs of the same input.
			verdicts[s.ID] = o.EvaluateRuns([][]harness.Event{r1, r2})
		} else {
			verdicts[s.ID] = o.Evaluate(r1)
		}
	}

	// Integrity: C's S1 emission reproduces the frozen nominal trace.
	if !reflect.DeepEqual(traces["S1"], harness.ScenarioFailMid.Trace) {
		t.Errorf("S1: C emission diverges from the frozen nominal trace:\n got %v\nwant %v",
			traces["S1"], harness.ScenarioFailMid.Trace)
	}

	// Spec point 3 evidence: the canonical order of every persisted fact is
	// its log index (append position), never a timestamp or scheduler order.
	c := newEventSourcingC(nil)
	c.run(cCommands["S1"])
	if msg := c.log.checkCanonicalOrder(); msg != "" {
		t.Errorf("canonical order violated: %s", msg)
	}

	t.Logf("")
	t.Logf("  %-24s | %-4s | %-4s | %-4s | %-4s | %-4s", "property", "S1", "S2", "S3", "S4", "S5")
	t.Logf("  " + strings.Repeat("-", 52))
	for _, p := range props {
		row := fmt.Sprintf("  %-24s", p)
		for _, s := range harness.Battery {
			row += fmt.Sprintf(" | %-4s", verdicts[s.ID][p])
		}
		t.Logf(row)
	}
	t.Logf("")
	t.Logf("  UNKNOWN = the scenario did not exercise the property's trigger in C's trace.")
	t.Logf("  P6 is structural (dual-run + source inspection), not event-based.")

	// Assertions: properties that MUST hold (P1 everywhere), the probes C
	// must pass, and the FAIL C must admit — everything else UNKNOWN.
	expected := map[string]map[harness.PropertyID]harness.Verdict{
		"S1": {
			harness.P1Auditability:  harness.Pass,
			harness.P3FailureBlocks: harness.Pass,
			harness.P4Recoverable:   harness.Pass,
		},
		"S2": {
			harness.P1Auditability:  harness.Pass,
			harness.P2Authorization: harness.Pass,
		},
		"S3": {
			harness.P1Auditability:  harness.Pass,
			harness.P5AsyncTolerant: harness.Pass,
		},
		"S4": {
			harness.P1Auditability: harness.Pass,
			harness.P8Isolation:    harness.Fail,
		},
		"S5": {
			harness.P1Auditability: harness.Pass,
			harness.P7Determinism:  harness.Pass,
		},
	}
	for _, s := range harness.Battery {
		for _, p := range props {
			want, ok := expected[s.ID][p]
			if !ok {
				want = harness.Unknown
			}
			if verdicts[s.ID][p] != want {
				t.Errorf("%s: %s = %s, want %s", s.ID, p, verdicts[s.ID][p], want)
			}
		}
	}
}

// TestC_AdmissionRejectionMapping verifies the rejection path of the
// frozen spec: rejected ⇒ contract rejection event, NO append.
func TestC_AdmissionRejectionMapping(t *testing.T) {
	c := newEventSourcingC(func(cap string) bool { return cap != "deny" })
	tr := c.run([]cCommand{{kind: cCmdAdmit, cap: "deny"}})
	if len(tr) != 3 || tr[0].Kind != harness.EvAdmissionRequested ||
		tr[1].Kind != harness.EvAdmissionRejected || tr[2].Kind != harness.EvAuditWritten {
		t.Fatalf("rejection mapping wrong: %v", tr)
	}
	if len(c.log.facts) != 0 {
		t.Fatalf("rejected admission must not append, log=%v", c.log.facts)
	}
}

// TestC_RollbackEmitsOnce verifies rollback semantics: the projection is
// re-derived from the log prefix, exactly one RollbackExecuted is emitted,
// and no replayed event is re-emitted.
func TestC_RollbackEmitsOnce(t *testing.T) {
	c := newEventSourcingC(nil)
	tr := c.run(cCommands["S1"])
	n := 0
	for _, e := range tr {
		if e.Kind == harness.EvRollbackExecuted {
			n++
		}
	}
	if n != 1 {
		t.Errorf("rollback must emit exactly one RollbackExecuted, got %d", n)
	}
	// After rollback the projection no longer shows c1 as failed.
	if st := project(&c.log).state["c1"]; st == cCapFailed {
		t.Errorf("rollback left c1 in failed state: %v", st)
	}
}
