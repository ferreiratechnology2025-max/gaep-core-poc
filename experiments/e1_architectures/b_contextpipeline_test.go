// Architecture B — Context Object + Policy Pipeline — rival of the E1 battery.
//
// Born from the context-object/pipeline paradigm (frozen spec,
// experiment-plan.md, block B): state is a set of typed immutable context
// objects (session, transaction, capability, audit); zero indexed cells or
// registers. Mutations produce NEW context objects (functional update) —
// the previous object is never modified. A single dispatcher serializes
// every pipeline traversal (single-writer); the canonical order IS the
// traversal position (never a timestamp, goroutine arrival or scheduler
// order). Policy is applied as fixed, ordered pipeline stages; each stage
// is a pure function (context) → (context', events, decision). A stage may
// reject — halting the traversal and emitting the contract rejection event
// (EvAdmissionRejected for a rejected admission, EvAuthorizationDenied for
// a denied effect) — or transform the context. Stages never emit outside
// decision boundaries. Rollback = re-traversal of the pipeline from a
// previous context snapshot (replay semantics): the current context is
// re-derived from the last pre-failure snapshot, exactly one
// RollbackExecuted is emitted and no prior stage decision is re-emitted.
// B is deterministic: pure stages + fixed order + single-writer ⇒ same
// commands ⇒ same context trajectory ⇒ same events (no stochastic clause).
//
// Provenance: B is implemented from scratch — no kernel/ (register file of
// A), models_comp/, C (c_eventsourcing_test.go) or E (e_statemachine_test.go)
// code is reused or adapted.
package e1_architectures_test

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"gaep-core-poc/harness"
)

// --- immutable context objects (snapshots; functional updates only) ---

// bCapCtx is the capability context snapshot. All fields are immutable:
// every change produces a NEW value; a snapshot already handed to history
// is never modified in place.
type bCapCtx struct {
	cap        string
	authorized bool
	running    bool
	succeeded  bool
	failed     bool
	timedOut   bool
}

// bSessCtx is the session context snapshot (admission decision).
type bSessCtx struct {
	cap      string
	admitted bool
}

// bTxCtx is the transaction context snapshot (execution lifecycle).
type bTxCtx struct {
	cap    string
	rolled bool
}

// bAuditCtx is the audit context snapshot (audit record count).
type bAuditCtx struct {
	written int
}

// --- fixed pipeline stages ---

type bStageID int

const (
	bStageAdmission bStageID = iota
	bStageAuthorization
	bStageExecution
	bStageEffect
	bStageTimeout
	bStageRollback
)

// bPipeline is the FIXED stage order (frozen spec point 4): a traversal
// walks this list; only the stage that decides for the command transforms
// the context, and a rejection halts the traversal (no further stage runs).
var bPipeline = []bStageID{
	bStageAdmission, bStageAuthorization, bStageExecution,
	bStageEffect, bStageTimeout, bStageRollback,
}

// bDecision is the pure output of a decision stage: the observable events
// emitted at the stage's decision boundary. made=true means the stage
// decided (emission expected); a decision with no events would be a
// passthrough.
type bDecision struct {
	made   bool
	events []harness.EventKind
}

// --- pure stage functions: (context) → (context', events, decision) ---

// bAdmissionDecide is pure. policyOK=false rejects the admission: the
// traversal halts and the contract rejection event is emitted; no
// admission is recorded in the session.
func bAdmissionDecide(s bSessCtx, policyOK bool) (bSessCtx, bDecision) {
	s.admitted = policyOK
	if !policyOK {
		return s, bDecision{made: true, events: []harness.EventKind{harness.EvAdmissionRequested, harness.EvAdmissionRejected}}
	}
	return s, bDecision{made: true, events: []harness.EventKind{harness.EvAdmissionRequested, harness.EvAdmissionAccepted}}
}

// bAuthorizeDecide is pure: grants authorization on the capability context.
func bAuthorizeDecide(c bCapCtx) (bCapCtx, bDecision) {
	c.authorized = true
	return c, bDecision{made: true, events: []harness.EventKind{harness.EvAuthorizationGranted}}
}

// bExecutionDecide is pure: start/succeed/fail on the capability context.
func bExecutionDecide(c bCapCtx, k harness.EventKind) (bCapCtx, bDecision) {
	switch k {
	case harness.EvCapabilityStarted:
		c.running = true
	case harness.EvCapabilitySucceeded:
		c.running = false
		c.succeeded = true
	case harness.EvCapabilityFailed:
		c.running = false
		c.failed = true
	}
	return c, bDecision{made: true, events: []harness.EventKind{k}}
}

// bEffectDecide is pure. allowed=false means the effect stage REJECTS the
// traversal: the contract denial event is emitted and the capability
// context is left unchanged (the violation never becomes observable).
func bEffectDecide(c bCapCtx, allowed bool) (bCapCtx, bDecision) {
	if !allowed {
		return c, bDecision{made: true, events: []harness.EventKind{harness.EvAuthorizationDenied}}
	}
	return c, bDecision{made: true, events: []harness.EventKind{harness.EvExternalEffect}}
}

// bTimeoutDecide is pure: marks the capability as timed out.
func bTimeoutDecide(c bCapCtx) (bCapCtx, bDecision) {
	c.timedOut = true
	return c, bDecision{made: true, events: []harness.EventKind{harness.EvTimedOut}}
}

// --- architecture B: single-writer dispatcher + traversal ---

type bCmdKind int

const (
	bCmdAdmit bCmdKind = iota
	bCmdAuthorize
	bCmdStart
	bCmdSucceed
	bCmdFail
	bCmdRollback
	bCmdEffect
	bCmdTimeout
)

type bCommand struct {
	kind bCmdKind
	cap  string
}

// contextPipelineB is architecture B. Single-writer: run() holds the mutex
// for the whole command batch, so every pipeline traversal is serialized
// at one point (frozen spec point 2).
type contextPipelineB struct {
	mu     sync.Mutex
	policy func(cap string) bool // admission policy (pure; nil ⇒ admit all)
	sess   map[string]bSessCtx   // current session snapshot per capability
	caps   map[string]bCapCtx    // current capability snapshot per capability
	tx     map[string]bTxCtx     // current transaction snapshot per capability
	audit  bAuditCtx             // current audit snapshot
	hist   map[string][]bCapCtx  // immutable snapshot history per capability
	events []harness.Event
	seq    int
}

func newContextPipelineB(policy func(cap string) bool) *contextPipelineB {
	if policy == nil {
		policy = func(string) bool { return true }
	}
	return &contextPipelineB{
		policy: policy,
		sess:   map[string]bSessCtx{},
		caps:   map[string]bCapCtx{},
		tx:     map[string]bTxCtx{},
		hist:   map[string][]bCapCtx{},
	}
}

// run processes cmds in canonical (traversal) order and returns the emitted
// observable events (a fresh copy). Deterministic: same cmds + same policy
// ⇒ same context trajectory ⇒ same events.
func (b *contextPipelineB) run(cmds []bCommand) []harness.Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sess = map[string]bSessCtx{}
	b.caps = map[string]bCapCtx{}
	b.tx = map[string]bTxCtx{}
	b.hist = map[string][]bCapCtx{}
	b.audit = bAuditCtx{}
	b.events = nil
	b.seq = 0
	for _, cmd := range cmds {
		b.traverse(cmd)
	}
	out := make([]harness.Event, len(b.events))
	copy(out, b.events)
	return out
}

func (b *contextPipelineB) emit(kind harness.EventKind, cap string) {
	b.seq++
	b.events = append(b.events, harness.Event{Seq: b.seq, Kind: kind, Cap: cap})
}

// checkCanonicalOrder asserts spec point 3: the sequence identity of an
// event IS its pipeline traversal position — never a timestamp or
// scheduler order. Returns an error description, or "" when satisfied.
func (b *contextPipelineB) checkCanonicalOrder() string {
	for i, e := range b.events {
		if e.Seq != i+1 {
			return fmt.Sprintf("event %d has seq %d (want %d)", i, e.Seq, i+1)
		}
	}
	return ""
}

// capCtx reads the current capability snapshot, seeding the zero value with
// the capability name (functional update discipline: never mutate in place).
func (b *contextPipelineB) capCtx(cap string) bCapCtx {
	c := b.caps[cap]
	if c.cap == "" {
		c.cap = cap
	}
	return c
}

// traverse walks the FIXED pipeline (frozen spec points 4 and 7): each
// command starts a traversal for the target context object; the stage that
// decides for the command transforms the immutable context and emits at the
// decision boundary; a rejected traversal halts (no further stage runs).
func (b *contextPipelineB) traverse(cmd bCommand) {
	for _, stage := range bPipeline {
		if b.stageFor(cmd.kind) != stage {
			continue // passthrough: this stage does not decide for cmd
		}
		b.runStage(stage, cmd)
		return
	}
}

// stageFor maps a command to the pipeline stage that decides it.
func (b *contextPipelineB) stageFor(k bCmdKind) bStageID {
	switch k {
	case bCmdAdmit:
		return bStageAdmission
	case bCmdAuthorize:
		return bStageAuthorization
	case bCmdStart, bCmdSucceed, bCmdFail:
		return bStageExecution
	case bCmdEffect:
		return bStageEffect
	case bCmdTimeout:
		return bStageTimeout
	}
	return bStageRollback // bCmdRollback
}

func (b *contextPipelineB) runStage(stage bStageID, cmd bCommand) {
	switch stage {
	case bStageAdmission:
		s := b.sess[cmd.cap]
		s2, d := bAdmissionDecide(s, b.policy(cmd.cap))
		b.sess[cmd.cap] = s2 // new snapshot; the previous s is untouched
		b.applyDecision(d, "")
	case bStageAuthorization:
		c := b.capCtx(cmd.cap)
		c2, d := bAuthorizeDecide(c)
		b.caps[cmd.cap] = c2
		b.hist[cmd.cap] = append(b.hist[cmd.cap], c2)
		b.applyDecision(d, cmd.cap)
	case bStageExecution:
		var kind harness.EventKind
		switch cmd.kind {
		case bCmdStart:
			kind = harness.EvCapabilityStarted
		case bCmdSucceed:
			kind = harness.EvCapabilitySucceeded
		case bCmdFail:
			kind = harness.EvCapabilityFailed
		}
		c := b.capCtx(cmd.cap)
		c2, d := bExecutionDecide(c, kind)
		b.caps[cmd.cap] = c2
		b.hist[cmd.cap] = append(b.hist[cmd.cap], c2)
		b.applyDecision(d, cmd.cap)
	case bStageEffect:
		c := b.capCtx(cmd.cap)
		allowed := c.authorized && !c.failed && !c.timedOut
		_, d := bEffectDecide(c, allowed)
		// Effect does not change the capability state: no new snapshot.
		b.applyDecision(d, cmd.cap)
	case bStageTimeout:
		c := b.capCtx(cmd.cap)
		c2, d := bTimeoutDecide(c)
		b.caps[cmd.cap] = c2
		b.hist[cmd.cap] = append(b.hist[cmd.cap], c2)
		b.applyDecision(d, cmd.cap)
	case bStageRollback:
		// Rollback = re-traversal from a previous snapshot (frozen spec
		// point 6): the current capability context is re-derived from the
		// last pre-failure snapshot, silently — no prior stage decision is
		// re-emitted; the transaction context closes.
		b.caps[cmd.cap] = b.restoreFromSnapshot(cmd.cap)
		tx := b.tx[cmd.cap]
		tx.cap = cmd.cap
		tx.rolled = true
		b.tx[cmd.cap] = tx
		b.applyDecision(bDecision{made: true, events: []harness.EventKind{harness.EvRollbackExecuted}}, cmd.cap)
	}
}

// restoreFromSnapshot re-derives the current capability context from its
// last pre-failure snapshot (frozen spec point 6, replay semantics). The
// restored snapshot IS the result of re-running the pipeline prefix from
// that point silently: every stage transform is pure and idempotent, so
// re-derivation yields the snapshot itself. Sharing the historical
// snapshot is safe: contexts are never modified in place, so the failed
// snapshot stays intact in history while current re-derives.
func (b *contextPipelineB) restoreFromSnapshot(cap string) bCapCtx {
	hist := b.hist[cap]
	restore := -1
	for i := range hist {
		if !hist[i].failed {
			restore = i
		}
	}
	if restore < 0 {
		return bCapCtx{cap: cap}
	}
	return hist[restore]
}

// applyDecision emits the stage's decision events at the decision boundary
// and closes the audit record (P1: every state-affecting decision is
// audited). The audit context snapshot advances.
func (b *contextPipelineB) applyDecision(d bDecision, cap string) {
	if !d.made {
		return
	}
	for _, k := range d.events {
		b.emit(k, bCapFor(k, cap))
	}
	b.audit.written++
	b.emit(harness.EvAuditWritten, "")
}

// bCapFor maps a decision event to the capability it refers to. Admission
// and audit are GLOBAL events (Cap "") per the observable contract
// (events.go: "global events such as admission or audit").
func bCapFor(kind harness.EventKind, cap string) string {
	switch kind {
	case harness.EvAdmissionRequested, harness.EvAdmissionAccepted,
		harness.EvAdmissionRejected, harness.EvAuditWritten:
		return ""
	}
	return cap
}

// --- scenario command scripts (identical inputs for every run) ---

// bCommands drives B through each battery scenario. Content is IDENTICAL
// to C's and E's command scripts (same capabilities, same order — identical
// protocol inputs); only the implementation is B's own.
var bCommands = map[string][]bCommand{
	// S1 — capability fails mid-execution (P1/P3/P4).
	"S1": {
		{bCmdAdmit, "c1"}, {bCmdAuthorize, "c1"}, {bCmdStart, "c1"},
		{bCmdFail, "c1"}, {bCmdRollback, "c1"},
	},
	// S2 — an external effect is attempted before authorization: the
	// effect stage REJECTS the traversal (EvAuthorizationDenied, frozen
	// spec point 4) and the violation never becomes observable; after
	// authorization the same effect is allowed (P2).
	"S2": {
		{bCmdAdmit, "c1"},
		{bCmdEffect, "c1"}, // denied: capability not authorized yet
		{bCmdAuthorize, "c1"},
		{bCmdEffect, "c1"}, // allowed: capability authorized
	},
	// S3 — late success after timeout: timeout is recorded, the late
	// completion records a success, and no effect is emitted (P5).
	"S3": {
		{bCmdAdmit, "c1"}, {bCmdAuthorize, "c1"}, {bCmdStart, "c1"},
		{bCmdTimeout, "c1"}, {bCmdSucceed, "c1"},
	},
	// S4 — two capabilities interleaved: B's frozen spec has no
	// per-capability isolation mechanism (contexts are per-capability but
	// the traversal order is the fixed dispatch order), so the
	// interleaving stays observable (P8 probe: the oracle decides).
	"S4": {
		{bCmdAuthorize, "c1"}, {bCmdStart, "c1"},
		{bCmdStart, "c2"}, {bCmdSucceed, "c2"}, {bCmdSucceed, "c1"},
	},
	// S5 — identical admission request repeated (P7, two runs).
	"S5": {
		{bCmdAdmit, "c1"},
	},
}

// archB adapts B to the scaffold's archColumn convention (S1 wired; the
// full battery runs in TestE1_ArchitectureB_Battery).
func archB() archColumn {
	return archColumn{
		id:   "B",
		name: "Context Object + Policy Pipeline",
		emit: func(h ScenarioHandle) []harness.Event {
			if h == scnFailMid {
				return newContextPipelineB(nil).run(bCommands["S1"])
			}
			return nil
		},
	}
}

// TestB_PipelineFixedOrder sanity-checks the frozen pipeline structure: the
// stage list is fixed, non-empty and ordered as specified.
func TestB_PipelineFixedOrder(t *testing.T) {
	want := []bStageID{bStageAdmission, bStageAuthorization, bStageExecution,
		bStageEffect, bStageTimeout, bStageRollback}
	if !reflect.DeepEqual(bPipeline, want) {
		t.Errorf("pipeline order changed: %v", bPipeline)
	}
}

// TestB_AdmissionRejectionMapping verifies the frozen rejection path: the
// admission stage rejects ⇒ contract rejection event, and no admission is
// recorded in the session snapshot.
func TestB_AdmissionRejectionMapping(t *testing.T) {
	b := newContextPipelineB(func(cap string) bool { return cap != "deny" })
	tr := b.run([]bCommand{{kind: bCmdAdmit, cap: "deny"}})
	if len(tr) != 3 || tr[0].Kind != harness.EvAdmissionRequested ||
		tr[1].Kind != harness.EvAdmissionRejected || tr[2].Kind != harness.EvAuditWritten {
		t.Fatalf("rejection mapping wrong: %v", tr)
	}
	if b.sess["deny"].admitted {
		t.Fatalf("rejected admission must not be recorded, sess=%v", b.sess["deny"])
	}
}

// TestB_EffectDenial verifies the effect stage's rejection boundary: an
// effect before authorization emits exactly one EvAuthorizationDenied (the
// violation never becomes observable); after authorization the same effect
// is emitted.
func TestB_EffectDenial(t *testing.T) {
	b := newContextPipelineB(nil)
	tr := b.run(bCommands["S2"])
	denials, effects := 0, 0
	deniedBeforeGrant := false
	granted := false
	for _, e := range tr {
		switch e.Kind {
		case harness.EvAuthorizationDenied:
			denials++
			if !granted {
				deniedBeforeGrant = true
			}
		case harness.EvAuthorizationGranted:
			granted = true
		case harness.EvExternalEffect:
			effects++
		}
	}
	if denials != 1 || !deniedBeforeGrant {
		t.Errorf("expected exactly one AuthorizationDenied before the grant, got denials=%d", denials)
	}
	if effects != 1 || !granted {
		t.Errorf("expected one ExternalEffect after authorization, got effects=%d", effects)
	}
}

// TestB_RollbackEmitsOnce verifies rollback semantics: exactly one
// RollbackExecuted, the current context is re-derived from the last
// pre-failure snapshot, and the failed snapshot stays intact in history
// (immutability: the previous object is never modified).
func TestB_RollbackEmitsOnce(t *testing.T) {
	b := newContextPipelineB(nil)
	tr := b.run(bCommands["S1"])
	n := 0
	for _, e := range tr {
		if e.Kind == harness.EvRollbackExecuted {
			n++
		}
	}
	if n != 1 {
		t.Errorf("rollback must emit exactly one RollbackExecuted, got %d", n)
	}
	cur := b.caps["c1"]
	if cur.failed {
		t.Errorf("rollback left c1 in failed state: %v", cur)
	}
	// The failed snapshot must still exist in history, untouched.
	found := false
	for _, s := range b.hist["c1"] {
		if s.failed {
			found = true
		}
	}
	if !found {
		t.Errorf("failed snapshot missing from history (immutability violated?): %v", b.hist["c1"])
	}
	// Canonical order: every emitted event's seq equals its traversal
	// position (frozen spec point 3).
	if msg := b.checkCanonicalOrder(); msg != "" {
		t.Errorf("canonical order violated: %s", msg)
	}
}
