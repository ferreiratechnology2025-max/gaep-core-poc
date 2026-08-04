// Battery tests for architecture D (Actor Model) under the pre-registered
// stochastic clause: N=50 fresh runs per scenario; the representative
// verdict per cell is the MODAL verdict over the N runs; cells whose modal
// frequency is below 50% are marked "high variance". Single-capability
// scenarios (S1/S2/S3/S5) must be point-mass (one actor ⇒ internal order
// preserved) — a multi-verdict cell there is an implementation bug, not a
// paradigm property. P7 is evaluated over the N runs TOGETHER
// (EvaluateRuns), which is the Oracle's protocol for determinism. D runs
// the same S1–S5 command scripts as B/C/E (identical inputs).
//
// The convergence table compares D's representative verdict against the
// B/C/E matrix, which is computed LIVE from the existing deterministic
// implementations (all three converge to the same matrix, verified in
// phases 3–5). D's implementation itself does not import or call B/C/E.
package e1_architectures_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gaep-core-poc/harness"
)

var allProps = []harness.PropertyID{
	harness.P1Auditability, harness.P2Authorization, harness.P3FailureBlocks,
	harness.P4Recoverable, harness.P5AsyncTolerant, harness.P6DomainFree,
	harness.P7Determinism, harness.P8Isolation,
}

// modalOf returns the verdict with the highest count and its count.
func modalOf(counts map[harness.Verdict]int) (harness.Verdict, int) {
	modal := harness.Unknown
	maxC := 0
	for _, v := range []harness.Verdict{harness.Pass, harness.Fail, harness.Unknown} {
		if counts[v] > maxC {
			maxC = counts[v]
			modal = v
		}
	}
	return modal, maxC
}

// refBCEMatrix computes the B/C/E matrix live (all three are deterministic
// and converge to the same matrix; any divergence is a harness failure).
// Evaluation protocol mirrors the B/C/E battery tests: one run per
// scenario via Evaluate; S5 via EvaluateRuns over two runs (P7).
func refBCEMatrix(t *testing.T, o *harness.Oracle) map[string]map[harness.PropertyID]harness.Verdict {
	scripts := []struct {
		name string
		run  func(id string) []harness.Event
	}{
		{"B", func(id string) []harness.Event { return newContextPipelineB(nil).run(bCommands[id]) }},
		{"C", func(id string) []harness.Event { return newEventSourcingC(nil).run(cCommands[id]) }},
		{"E", func(id string) []harness.Event { return newEStateMachine(nil).run(eCommands[id]) }},
	}
	m := map[string]map[harness.PropertyID]harness.Verdict{}
	for _, s := range harness.Battery {
		var ref map[harness.PropertyID]harness.Verdict
		for i, sc := range scripts {
			var v map[harness.PropertyID]harness.Verdict
			if s.ID == "S5" {
				r1, r2 := sc.run(s.ID), sc.run(s.ID)
				v = o.EvaluateRuns([][]harness.Event{r1, r2})
			} else {
				v = o.Evaluate(sc.run(s.ID))
			}
			if i == 0 {
				ref = v
				continue
			}
			for _, p := range allProps {
				if v[p] != ref[p] {
					t.Fatalf("reference matrices diverge: %s vs %s at %s×%s (%s vs %s)",
						scripts[0].name, sc.name, s.ID, p, ref[p], v[p])
				}
			}
		}
		m[s.ID] = ref
	}
	return m
}

// TestE1_ArchitectureD_Battery runs D through S1–S5 under the stochastic
// clause (N=50), logs the distributional matrix, the convergence table vs
// B/C/E, high-variance cells and observability gaps, and asserts the
// deterministic invariants: point-mass single-capability scenarios, the
// expected modals (except S4×P8, which is deliberately NOT asserted — it
// is stochastic), and the frozen S1 trace.
func TestE1_ArchitectureD_Battery(t *testing.T) {
	const N = 50
	o := harness.NewOracle()
	ref := refBCEMatrix(t, o)

	t.Logf("E1 — architecture D (Actor Model) × battery S1–S5, N=%d (stochastic clause, pre-registered)", N)
	t.Logf("Pre-registration: representative verdict = MODAL over N fresh runs; modal frequency < 50%% ⇒ high variance.")
	t.Logf("Single-capability scenarios (S1/S2/S3/S5) must be point-mass (one actor ⇒ internal order preserved).")
	t.Logf("P7 is evaluated over the N runs TOGETHER (EvaluateRuns) — the Oracle's protocol for determinism.")
	t.Logf("")

	dist := map[string]map[harness.PropertyID]map[harness.Verdict]int{}
	traces := map[string][][]harness.Event{}
	for _, s := range harness.Battery {
		dist[s.ID] = map[harness.PropertyID]map[harness.Verdict]int{}
		cmds, ok := dCommands[s.ID]
		if !ok {
			t.Fatalf("no command script for scenario %s", s.ID)
		}
		for i := 0; i < N; i++ {
			tr, ok := newActorModelD(nil).run(cmds)
			if !ok {
				t.Fatalf("%s run %d: D.run() timed out — implementation failure (deadlock), NOT a result", s.ID, i)
			}
			traces[s.ID] = append(traces[s.ID], tr)
			v := o.Evaluate(tr)
			for _, p := range allProps {
				if dist[s.ID][p] == nil {
					dist[s.ID][p] = map[harness.Verdict]int{}
				}
				dist[s.ID][p][v[p]]++
			}
		}
	}

	// Integrity: D's S1 emission reproduces the frozen nominal trace
	// (single actor ⇒ deterministic).
	if !reflect.DeepEqual(traces["S1"][0], harness.ScenarioFailMid.Trace) {
		t.Errorf("S1: D emission diverges from the frozen nominal trace:\n got %v\nwant %v",
			traces["S1"][0], harness.ScenarioFailMid.Trace)
	}

	// Representative verdict per cell: modal over N (P7: set evaluation
	// over the N runs together).
	rep := map[string]harness.Verdict{}
	var highVar []string
	for _, s := range harness.Battery {
		t.Logf("  %s — %s", s.ID, s.Name)
		for _, p := range allProps {
			counts := dist[s.ID][p]
			key := fmt.Sprintf("%s×%s", s.ID, p)
			if p == harness.P7Determinism {
				setV := o.EvaluateRuns(traces[s.ID])[p]
				rep[key] = setV
				t.Logf("    %-3s | individual PASS=%d FAIL=%d UNKNOWN=%d | conjunto(%d runs) = %s",
					p, counts[harness.Pass], counts[harness.Fail], counts[harness.Unknown], N, setV)
				continue
			}
			modal, maxC := modalOf(counts)
			freq := float64(maxC) / float64(N)
			rep[key] = modal
			tag := ""
			if freq < 0.5 {
				tag = "  ← ALTA VARIÂNCIA (freq modal < 50%)"
				highVar = append(highVar, key)
			}
			t.Logf("    %-3s | PASS=%d FAIL=%d UNKNOWN=%d | modal %-7s | freq %.0f%%%s",
				p, counts[harness.Pass], counts[harness.Fail], counts[harness.Unknown], modal, freq*100, tag)
		}
	}

	// Convergence table: informative cells only (B/C/E verdict ≠ UNKNOWN).
	t.Logf("")
	t.Logf("  " + strings.Repeat("─", 62))
	t.Logf("Convergence D × (B/C/E) — informative cells only (B/C/E verdict ≠ UNKNOWN):")
	var diverged []string
	convCells, convYes := 0, 0
	for _, s := range harness.Battery {
		for _, p := range allProps {
			refV := ref[s.ID][p]
			if refV == harness.Unknown {
				continue
			}
			key := fmt.Sprintf("%s×%s", s.ID, p)
			dV := rep[key]
			convCells++
			mark := "SIM"
			if dV != refV {
				mark = "NÃO"
				diverged = append(diverged, fmt.Sprintf("%s: D=%s vs B/C/E=%s", key, dV, refV))
			} else {
				convYes++
			}
			t.Logf("    %-8s | D %-7s | B/C/E %-7s | converge %s", key, dV, refV, mark)
		}
	}

	// Observability gaps: B/C/E = UNKNOWN while D has a verdict. These are
	// protocol asymmetries (e.g. P7: B/C/E evaluated with a single run in
	// S1–S3, D with N=50), not mechanism divergences.
	t.Logf("")
	t.Logf("Observability gaps (B/C/E = UNKNOWN, D has a verdict — protocol asymmetry, NOT mechanism divergence):")
	gapCount := 0
	for _, s := range harness.Battery {
		for _, p := range allProps {
			refV := ref[s.ID][p]
			if refV != harness.Unknown {
				continue
			}
			key := fmt.Sprintf("%s×%s", s.ID, p)
			dV := rep[key]
			if dV != harness.Unknown {
				gapCount++
				t.Logf("    %-8s | D %-7s | B/C/E UNKNOWN", key, dV)
			}
		}
	}
	if gapCount == 0 {
		t.Logf("    (none)")
	}

	t.Logf("")
	t.Logf("Summary: convergence %d/%d informative cells; high-variance cells: %d",
		convYes, convCells, len(highVar))

	// --- Assertions (deterministic invariants only; S4×P8 is stochastic
	// and deliberately NOT asserted) ---

	// Point-mass: every property in a single-capability scenario must have
	// exactly one verdict over the N runs.
	for _, id := range []string{"S1", "S2", "S3", "S5"} {
		for _, p := range allProps {
			distinct := 0
			for _, v := range []harness.Verdict{harness.Pass, harness.Fail, harness.Unknown} {
				if dist[id][p][v] > 0 {
					distinct++
				}
			}
			if distinct != 1 {
				t.Errorf("%s × %s: NOT point-mass (%d distinct verdicts over N=%d) — implementation bug",
					id, p, distinct, N)
			}
		}
	}

	// Expected representative verdicts (single-capability modals are the
	// every-run verdict; S4 asserts only P1 — structural; P8 is free).
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
		},
		"S5": {
			harness.P1Auditability: harness.Pass,
			harness.P7Determinism:  harness.Pass, // set evaluation over N runs
		},
	}
	for _, s := range harness.Battery {
		for _, p := range allProps {
			want, ok := expected[s.ID][p]
			if !ok {
				continue
			}
			key := fmt.Sprintf("%s×%s", s.ID, p)
			if rep[key] != want {
				t.Errorf("%s: representative = %s, want %s", key, rep[key], want)
			}
		}
	}
}
