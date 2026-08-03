package harness

// Scenario is one battery item: a property probe. Trace is the event
// sequence the architecture produced for the scenario — for well-behaved
// architectures it is the nominal sequence; for property probes such as S2
// it is the violating sequence by construction (the oracle must flag it).
// Adversarial, when present, is the same scenario with one or more
// violations injected; the oracle must return FAIL on it.
type Scenario struct {
	ID   string
	Name string
	// Need is the architectural need this scenario probes (see
	// adr-evidence/architectural-necessity-matrix.md). Solution-free.
	Need string
	// Trace is the event sequence produced for this scenario.
	Trace []Event
	// Adversarial is the same scenario with violations injected; the
	// oracle must return FAIL on it.
	Adversarial []Event
}

// ScenarioFailMid: a capability fails mid-execution and is rolled back and
// audited. Exercises P1 (audit), P3 (failure blocks effects), P4 (rollback).
var ScenarioFailMid = Scenario{
	ID:   "S1",
	Name: "capability fails mid-execution",
	Need: "N4/N8 — partial state recoverable; a failure blocks subsequent effects",
	Trace: []Event{
		{Seq: 1, Kind: EvAdmissionRequested},
		{Seq: 2, Kind: EvAdmissionAccepted},
		{Seq: 3, Kind: EvAuditWritten},
		{Seq: 4, Kind: EvAuthorizationGranted, Cap: "c1"},
		{Seq: 5, Kind: EvAuditWritten},
		{Seq: 6, Kind: EvCapabilityStarted, Cap: "c1"},
		{Seq: 7, Kind: EvAuditWritten},
		{Seq: 8, Kind: EvCapabilityFailed, Cap: "c1"},
		{Seq: 9, Kind: EvAuditWritten},
		{Seq: 10, Kind: EvRollbackExecuted, Cap: "c1"},
		{Seq: 11, Kind: EvAuditWritten},
	},
}

// ScenarioEffectBeforeAuth: adversarial by construction — an external
// effect occurs before authorization. Exercises P2.
var ScenarioEffectBeforeAuth = Scenario{
	ID:   "S2",
	Name: "external effect before authorization",
	Need: "N1 — never execute an effect after loss of authorization",
	Trace: []Event{
		{Seq: 1, Kind: EvAdmissionAccepted},
		{Seq: 2, Kind: EvExternalEffect, Cap: "c1"}, // violation of P2
		{Seq: 3, Kind: EvAuthorizationGranted, Cap: "c1"},
	},
}

// ScenarioLateSuccessIgnored: a late success after timeout is ignored
// (nominal) or applied (adversarial). Exercises P5.
var ScenarioLateSuccessIgnored = Scenario{
	ID:   "S3",
	Name: "late success after timeout",
	Need: "N2 — be able to interrupt a pending execution",
	Trace: []Event{
		{Seq: 1, Kind: EvAdmissionAccepted},
		{Seq: 2, Kind: EvAuthorizationGranted, Cap: "c1"},
		{Seq: 3, Kind: EvCapabilityStarted, Cap: "c1"},
		{Seq: 4, Kind: EvTimedOut, Cap: "c1"},
		{Seq: 5, Kind: EvCapabilitySucceeded, Cap: "c1"}, // late, ignored: no effect
	},
	Adversarial: []Event{
		{Seq: 1, Kind: EvAdmissionAccepted},
		{Seq: 2, Kind: EvAuthorizationGranted, Cap: "c1"},
		{Seq: 3, Kind: EvCapabilityStarted, Cap: "c1"},
		{Seq: 4, Kind: EvTimedOut, Cap: "c1"},
		{Seq: 5, Kind: EvCapabilitySucceeded, Cap: "c1"}, // late…
		{Seq: 6, Kind: EvExternalEffect, Cap: "c1"},      // …but applied: violation of P5
	},
}

// ScenarioConcurrent: two capabilities interleave — one execution is
// observed inside the other's window. Exercises P8.
var ScenarioConcurrent = Scenario{
	ID:   "S4",
	Name: "two capabilities, interleaved",
	Need: "N6 — executions must not interfere with each other",
	Trace: []Event{
		{Seq: 1, Kind: EvAuthorizationGranted, Cap: "c1"},
		{Seq: 2, Kind: EvCapabilityStarted, Cap: "c1"},
		{Seq: 3, Kind: EvCapabilityStarted, Cap: "c2"}, // interleaving inside c1's window
		{Seq: 4, Kind: EvCapabilitySucceeded, Cap: "c2"},
		{Seq: 5, Kind: EvCapabilitySucceeded, Cap: "c1"},
	},
}

// ScenarioDeterministicAdmission: the same admission request repeated;
// the Oracle compares runs for P7.
var ScenarioDeterministicAdmission = Scenario{
	ID:   "S5",
	Name: "identical admission request, repeated",
	Need: "N5 — deterministic admission decision for the same input",
	Trace: []Event{
		{Seq: 1, Kind: EvAdmissionRequested},
		{Seq: 2, Kind: EvAdmissionAccepted, Cap: "c1"},
	},
}

// Battery is the ordered v1 scenario set.
var Battery = []Scenario{
	ScenarioFailMid,
	ScenarioEffectBeforeAuth,
	ScenarioLateSuccessIgnored,
	ScenarioConcurrent,
	ScenarioDeterministicAdmission,
}
