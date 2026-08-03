package harness

// PropertyID identifies an observable architectural property (P1–P8).
// The operational criteria are frozen in
// adr-evidence/architectural-properties.md; this file implements them.
type PropertyID string

const (
	P1Auditability  PropertyID = "P1" // execution is auditable
	P2Authorization PropertyID = "P2" // authorization precedes any external effect
	P3FailureBlocks PropertyID = "P3" // a failure blocks subsequent effects
	P4Recoverable   PropertyID = "P4" // partial state is recoverable
	P5AsyncTolerant PropertyID = "P5" // asynchronous execution is tolerated
	P6DomainFree    PropertyID = "P6" // kernel is independent of domain semantics
	P7Determinism   PropertyID = "P7" // admission decision is deterministic
	P8Isolation     PropertyID = "P8" // executions are isolated
)

// Verdict is one of PASS, FAIL, UNKNOWN.
type Verdict int

const (
	Pass Verdict = iota
	Fail
	Unknown
)

// String returns PASS, FAIL or UNKNOWN.
func (v Verdict) String() string {
	switch v {
	case Pass:
		return "PASS"
	case Fail:
		return "FAIL"
	}
	return "UNKNOWN"
}

// Oracle decides property verdicts from observable event traces. It holds
// no state and knows nothing about architectures: it only reads events.
type Oracle struct{}

// NewOracle returns an empty oracle.
func NewOracle() *Oracle { return &Oracle{} }

// stateAffecting reports whether an event changes state and therefore must
// be covered by an audit record (P1).
func stateAffecting(k EventKind) bool {
	switch k {
	case EvAdmissionAccepted, EvAdmissionRejected, EvAuthorizationGranted,
		EvAuthorizationDenied, EvCapabilityStarted, EvCapabilitySucceeded,
		EvCapabilityFailed, EvRollbackExecuted, EvExternalEffect, EvTimedOut:
		return true
	}
	return false
}

// Evaluate grades every property over a single run. P7 (determinism)
// requires repeated runs of the same input; use EvaluateRuns.
func (o *Oracle) Evaluate(trace []Event) map[PropertyID]Verdict {
	return o.EvaluateRuns([][]Event{trace})
}

// EvaluateRuns grades every property over one or more runs of the same
// scenario. P7 compares admission verdicts across runs; the remaining
// properties use the first run. An empty runs slice yields all UNKNOWN.
func (o *Oracle) EvaluateRuns(runs [][]Event) map[PropertyID]Verdict {
	if len(runs) == 0 {
		return map[PropertyID]Verdict{
			P1Auditability: Unknown, P2Authorization: Unknown, P3FailureBlocks: Unknown,
			P4Recoverable: Unknown, P5AsyncTolerant: Unknown, P6DomainFree: Unknown,
			P7Determinism: Unknown, P8Isolation: Unknown,
		}
	}
	trace := runs[0]
	return map[PropertyID]Verdict{
		P1Auditability:  o.p1(trace),
		P2Authorization: o.p2(trace),
		P3FailureBlocks: o.p3(trace),
		P4Recoverable:   o.p4(trace),
		P5AsyncTolerant: o.p5(trace),
		P6DomainFree:    Unknown, // structural: dual-run with swapped policy + source inspection
		P7Determinism:   o.p7(runs),
		P8Isolation:     o.p8(trace),
	}
}

// P1 — every state-affecting event is followed by an AuditWritten that
// covers it (post-hoc audit). Audits are consumed in order; an audit that
// arrives before any uncovered event is ignored.
func (o *Oracle) p1(trace []Event) Verdict {
	uncovered := 0
	for _, e := range trace {
		if stateAffecting(e.Kind) {
			uncovered++
		} else if e.Kind == EvAuditWritten && uncovered > 0 {
			uncovered--
		}
	}
	switch {
	case uncovered == 0 && hasStateAffecting(trace):
		return Pass
	case !hasStateAffecting(trace):
		return Unknown
	}
	return Fail
}

func hasStateAffecting(trace []Event) bool {
	for _, e := range trace {
		if stateAffecting(e.Kind) {
			return true
		}
	}
	return false
}

// P2 — no ExternalEffect precedes the first AuthorizationGranted of the
// same capability.
func (o *Oracle) p2(trace []Event) Verdict {
	auth := map[string]bool{}
	hasExt := false
	for _, e := range trace {
		switch e.Kind {
		case EvAuthorizationGranted:
			auth[e.Cap] = true
		case EvExternalEffect:
			hasExt = true
			if !auth[e.Cap] {
				return Fail
			}
		}
	}
	if !hasExt {
		return Unknown
	}
	return Pass
}

// P3 — after CapabilityFailed, no ExternalEffect until a fresh
// AuthorizationGranted for that capability.
func (o *Oracle) p3(trace []Event) Verdict {
	blocked := map[string]bool{}
	hasFail := false
	for _, e := range trace {
		switch e.Kind {
		case EvCapabilityFailed:
			blocked[e.Cap] = true
			hasFail = true
		case EvAuthorizationGranted:
			blocked[e.Cap] = false
		case EvExternalEffect:
			if blocked[e.Cap] {
				return Fail
			}
		}
	}
	if !hasFail {
		return Unknown
	}
	return Pass
}

// P4 — every failed capability is rolled back before the next start of
// that capability, and no failure is left pending at the end of the trace.
func (o *Oracle) p4(trace []Event) Verdict {
	pending := map[string]bool{}
	hasFail := false
	for _, e := range trace {
		switch e.Kind {
		case EvCapabilityFailed:
			pending[e.Cap] = true
			hasFail = true
		case EvRollbackExecuted:
			pending[e.Cap] = false
		case EvCapabilityStarted:
			if pending[e.Cap] {
				return Fail
			}
		}
	}
	for _, p := range pending {
		if p {
			return Fail
		}
	}
	if !hasFail {
		return Unknown
	}
	return Pass
}

// P5 — no ExternalEffect occurs after a TimedOut for the same capability:
// a late CapabilitySucceeded after timeout is ignored, and any effect after
// the timeout is a violation (stricter than the late-success case alone).
func (o *Oracle) p5(trace []Event) Verdict {
	timedOut := map[string]bool{}
	hasTO := false
	for _, e := range trace {
		switch e.Kind {
		case EvTimedOut:
			timedOut[e.Cap] = true
			hasTO = true
		case EvExternalEffect:
			if timedOut[e.Cap] {
				return Fail
			}
		}
	}
	if !hasTO {
		return Unknown
	}
	return Pass
}

// P7 — identical admission verdicts across repeated runs of the same input.
func (o *Oracle) p7(runs [][]Event) Verdict {
	var pattern []bool
	runsWithAdmission := 0
	for _, run := range runs {
		var acc []bool
		for _, e := range run {
			switch e.Kind {
			case EvAdmissionAccepted:
				acc = append(acc, true)
			case EvAdmissionRejected:
				acc = append(acc, false)
			}
		}
		if len(acc) == 0 {
			continue
		}
		runsWithAdmission++
		if pattern == nil {
			pattern = acc
			continue
		}
		if len(acc) != len(pattern) {
			return Fail
		}
		for i := range acc {
			if acc[i] != pattern[i] {
				return Fail
			}
		}
	}
	if runsWithAdmission < 2 {
		return Unknown
	}
	return Pass
}

// P8 — events of each capability are contiguous: no execution is observed
// inside another execution's window. v1 operational criterion (documented
// simplification in architectural-properties.md).
func (o *Oracle) p8(trace []Event) Verdict {
	last := map[string]int{}
	caps := map[string]bool{}
	for i, e := range trace {
		if e.Cap == "" {
			continue
		}
		caps[e.Cap] = true
		if li, ok := last[e.Cap]; ok {
			for j := li + 1; j < i; j++ {
				if trace[j].Cap != "" && trace[j].Cap != e.Cap {
					return Fail
				}
			}
		}
		last[e.Cap] = i
	}
	if len(caps) <= 1 {
		return Unknown
	}
	return Pass
}
