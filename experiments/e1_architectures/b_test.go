// Battery tests for architecture B (Context Object + Policy Pipeline): B
// runs the same S1–S5 scenario protocol as C and E (identical command
// content) and is graded only by the Property Oracle. Deterministic by
// construction.
package e1_architectures_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gaep-core-poc/harness"
)

// TestE1_ArchitectureB_Battery runs B through S1–S5 via the Property
// Oracle, logs the P1–P8 × S1–S5 matrix and asserts the expected verdicts.
func TestE1_ArchitectureB_Battery(t *testing.T) {
	o := harness.NewOracle()
	props := []harness.PropertyID{
		harness.P1Auditability, harness.P2Authorization, harness.P3FailureBlocks,
		harness.P4Recoverable, harness.P5AsyncTolerant, harness.P6DomainFree,
		harness.P7Determinism, harness.P8Isolation,
	}

	t.Logf("E1 — architecture B (Context Object + Policy Pipeline) × battery S1–S5 (Property Oracle)")
	t.Logf("Deterministic by construction: same commands ⇒ same context trajectory ⇒ same events.")

	traces := map[string][]harness.Event{}
	verdicts := map[string]map[harness.PropertyID]harness.Verdict{}
	for _, s := range harness.Battery {
		cmds, ok := bCommands[s.ID]
		if !ok {
			t.Fatalf("no command script for scenario %s", s.ID)
		}
		r1 := newContextPipelineB(nil).run(cmds)
		r2 := newContextPipelineB(nil).run(cmds)
		if !reflect.DeepEqual(r1, r2) {
			t.Errorf("%s: B is not deterministic (run 1 != run 2)", s.ID)
		}
		traces[s.ID] = r1
		if s.ID == "S5" {
			// P7 needs repeated runs of the same input.
			verdicts[s.ID] = o.EvaluateRuns([][]harness.Event{r1, r2})
		} else {
			verdicts[s.ID] = o.Evaluate(r1)
		}
	}

	// Integrity: B's S1 emission reproduces the frozen nominal trace.
	if !reflect.DeepEqual(traces["S1"], harness.ScenarioFailMid.Trace) {
		t.Errorf("S1: B emission diverges from the frozen nominal trace:\n got %v\nwant %v",
			traces["S1"], harness.ScenarioFailMid.Trace)
	}

	// Spec point 3 evidence: the canonical order of every emitted event is
	// its traversal position, never a timestamp or scheduler order.
	b := newContextPipelineB(nil)
	b.run(bCommands["S1"])
	if msg := b.checkCanonicalOrder(); msg != "" {
		t.Errorf("canonical order violated: %s", msg)
	}

	// Audit context evidence: one audit record per state-affecting
	// decision (P1), mirrored by the audit context snapshot.
	audits := 0
	for _, e := range traces["S1"] {
		if e.Kind == harness.EvAuditWritten {
			audits++
		}
	}
	if b.audit.written != audits {
		t.Errorf("audit context out of sync: written=%d, emitted=%d", b.audit.written, audits)
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
	t.Logf("  UNKNOWN = the scenario did not exercise the property's trigger in B's trace.")
	t.Logf("  P6 is structural (dual-run + source inspection), not event-based.")

	// Assertions: properties that MUST hold (P1 everywhere), the probes B
	// must pass, and the FAIL B must admit — everything else UNKNOWN.
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
