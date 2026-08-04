// Battery tests for architecture E (Hierarchical State Machine): E runs
// the same S1–S5 scenario protocol as C (identical command content) and is
// graded only by the Property Oracle. Deterministic by construction.
package e1_architectures_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gaep-core-poc/harness"
)

// TestE1_ArchitectureE_Battery runs E through S1–S5 via the Property
// Oracle, logs the P1–P8 × S1–S5 matrix and asserts the expected verdicts.
func TestE1_ArchitectureE_Battery(t *testing.T) {
	o := harness.NewOracle()
	props := []harness.PropertyID{
		harness.P1Auditability, harness.P2Authorization, harness.P3FailureBlocks,
		harness.P4Recoverable, harness.P5AsyncTolerant, harness.P6DomainFree,
		harness.P7Determinism, harness.P8Isolation,
	}

	t.Logf("E1 — architecture E (Hierarchical State Machine) × battery S1–S5 (Property Oracle)")
	t.Logf("Deterministic by construction: same inputs ⇒ same trajectory ⇒ same events.")

	traces := map[string][]harness.Event{}
	verdicts := map[string]map[harness.PropertyID]harness.Verdict{}
	for _, s := range harness.Battery {
		cmds, ok := eCommands[s.ID]
		if !ok {
			t.Fatalf("no command script for scenario %s", s.ID)
		}
		r1 := newEStateMachine(nil).run(cmds)
		r2 := newEStateMachine(nil).run(cmds)
		if !reflect.DeepEqual(r1, r2) {
			t.Errorf("%s: E is not deterministic (run 1 != run 2)", s.ID)
		}
		traces[s.ID] = r1
		if s.ID == "S5" {
			// P7 needs repeated runs of the same input.
			verdicts[s.ID] = o.EvaluateRuns([][]harness.Event{r1, r2})
		} else {
			verdicts[s.ID] = o.Evaluate(r1)
		}
	}

	// Integrity: E's S1 emission reproduces the frozen nominal trace.
	if !reflect.DeepEqual(traces["S1"], harness.ScenarioFailMid.Trace) {
		t.Errorf("S1: E emission diverges from the frozen nominal trace:\n got %v\nwant %v",
			traces["S1"], harness.ScenarioFailMid.Trace)
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
	t.Logf("  UNKNOWN = the scenario did not exercise the property's trigger in E's trace.")
	t.Logf("  P6 is structural (dual-run + source inspection), not event-based.")

	// Assertions: properties that MUST hold (P1 everywhere), the probes E
	// must pass, and the FAIL E must admit — everything else UNKNOWN.
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

// TestE_AdmissionBlockSilence verifies E's frozen guard semantics: a
// blocked transition fires nothing — no event, no state change (spec
// point 4). Unlike C, E does NOT emit AdmissionRejected on a blocked
// admission: the vocabulary event exists, but E's contract is silence.
func TestE_AdmissionBlockSilence(t *testing.T) {
	e := newEStateMachine(func(cap string) bool { return cap != "deny" })
	tr := e.run([]eCommand{{in: eInAdmit, cap: "deny"}})
	if len(tr) != 0 {
		t.Fatalf("blocked admission must emit no events, got %v", tr)
	}
	if e.mach["deny"] != eNone {
		t.Fatalf("blocked transition must not change state, got %v", e.mach["deny"])
	}
}

// TestE_RollbackEmitsOnce verifies rollback semantics: exactly one
// RollbackExecuted, and the machine reaches the recovery position without
// re-emitting prior transition outputs.
func TestE_RollbackEmitsOnce(t *testing.T) {
	e := newEStateMachine(nil)
	tr := e.run(eCommands["S1"])
	n := 0
	for _, ev := range tr {
		if ev.Kind == harness.EvRollbackExecuted {
			n++
		}
	}
	if n != 1 {
		t.Errorf("rollback must emit exactly one RollbackExecuted, got %d", n)
	}
	if e.mach["c1"] != eRecovered {
		t.Errorf("after rollback c1 should be in recovery position, got %v", e.mach["c1"])
	}
}
