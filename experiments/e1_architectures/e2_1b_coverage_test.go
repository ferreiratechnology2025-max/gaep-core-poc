// E2.1b — coverage expansion battery. Adds scenarios S6, S7 and S8 (per
// the frozen design in adr-evidence/E2_1a_scenario_design.md) and fixes
// the P7 evaluation protocol (2 fresh runs for every admission-bearing
// scenario, matching what S5 already did) for B, C and E. This phase is
// COVERAGE, not discrimination: the acceptance condition is that every
// S1–S5 cell that had a verdict in E1 reproduces the identical verdict
// here, with the single documented exception of P7×{S1,S2,S3} for B/C/E
// (UNKNOWN → PASS, purely a protocol-invocation change). D already
// evaluates P7 "in conjunto" over N runs for every scenario under its
// pre-registered stochastic clause (see d_test.go), so D's S1–S5 P7
// verdicts do not move.
//
// harness/oracle.go, harness/events.go, P1–P8 and the S1–S5 definitions
// in harness/scenarios.go are untouched. S6/S7/S8 are additive: they live
// only in this file's local scenario list and in the per-architecture
// command maps (bCommands/cCommands/dCommands/eCommands), extended in
// their own files. kernel/ and models_comp/ are untouched (irrelevant to
// B/C/D/E). D's actor implementation is untouched; S6/S7/S8 are new
// entries in its command map only.
package e1_architectures_test

import (
	"fmt"
	"strings"
	"testing"

	"gaep-core-poc/harness"
)

// fullIDs is the expanded battery: E1's frozen S1–S5 plus the E2.1b
// additive coverage scenarios S6–S8.
var fullIDs = []string{"S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8"}

// admissionBearing marks scenarios whose command script contains at least
// one Admit command — per the E2.1a design, P7 is evaluated over 2 fresh
// runs for these (protocol fix); S4 and S7 have no Admit command at all,
// so P7 correctly stays UNKNOWN there (CENÁRIO — no trigger — not a
// protocol gap).
var admissionBearing = map[string]bool{
	"S1": true, "S2": true, "S3": true, "S5": true, "S6": true, "S8": true,
}

// e1RefS1toS5 is the frozen E1 reference matrix for B/C/E over S1–S5
// (adr-evidence/E1_B_findings.md, E1_C_findings.md, E1_E_findings.md — the
// three converge byte-for-byte). Used only to assert non-regression; it is
// not derived from a live E1 run (E1 is closed, commit c665b40).
func e1RefS1toS5() map[string]map[harness.PropertyID]harness.Verdict {
	u, p, f := harness.Unknown, harness.Pass, harness.Fail
	return map[string]map[harness.PropertyID]harness.Verdict{
		"S1": {harness.P1Auditability: p, harness.P2Authorization: u, harness.P3FailureBlocks: p, harness.P4Recoverable: p, harness.P5AsyncTolerant: u, harness.P6DomainFree: u, harness.P7Determinism: u, harness.P8Isolation: u},
		"S2": {harness.P1Auditability: p, harness.P2Authorization: p, harness.P3FailureBlocks: u, harness.P4Recoverable: u, harness.P5AsyncTolerant: u, harness.P6DomainFree: u, harness.P7Determinism: u, harness.P8Isolation: u},
		"S3": {harness.P1Auditability: p, harness.P2Authorization: u, harness.P3FailureBlocks: u, harness.P4Recoverable: u, harness.P5AsyncTolerant: p, harness.P6DomainFree: u, harness.P7Determinism: u, harness.P8Isolation: u},
		"S4": {harness.P1Auditability: p, harness.P2Authorization: u, harness.P3FailureBlocks: u, harness.P4Recoverable: u, harness.P5AsyncTolerant: u, harness.P6DomainFree: u, harness.P7Determinism: u, harness.P8Isolation: f},
		"S5": {harness.P1Auditability: p, harness.P2Authorization: u, harness.P3FailureBlocks: u, harness.P4Recoverable: u, harness.P5AsyncTolerant: u, harness.P6DomainFree: u, harness.P7Determinism: p, harness.P8Isolation: u},
	}
}

// expectedProtocolDelta is the ONLY admissible change vs the E1 reference:
// P7 at S1/S2/S3 moves UNKNOWN → PASS for B/C/E (protocol fix). Any other
// moved cell is a regression, per the E2.1b acceptance condition.
var expectedProtocolDelta = map[string]harness.PropertyID{
	"S1": harness.P7Determinism,
	"S2": harness.P7Determinism,
	"S3": harness.P7Determinism,
}

// detArch is a deterministic architecture column (B, C, E) for the
// coverage battery.
type detArch struct {
	id   string
	name string
	run  func(id string) []harness.Event
}

func detArchs() []detArch {
	return []detArch{
		{"B", "Context Object + Policy Pipeline", func(id string) []harness.Event {
			return newContextPipelineB(nil).run(bCommands[id])
		}},
		{"C", "Event Sourcing", func(id string) []harness.Event {
			return newEventSourcingC(nil).run(cCommands[id])
		}},
		{"E", "Hierarchical State Machine", func(id string) []harness.Event {
			return newEStateMachine(nil).run(eCommands[id])
		}},
	}
}

// evalCell evaluates one scenario for a deterministic architecture: 2
// fresh runs always (needed both for the P7 protocol and the existing
// determinism check), routed through EvaluateRuns when the scenario is
// admission-bearing, else through a single-run Evaluate (P7 stays
// correctly UNKNOWN — no trigger, not a protocol gap).
func evalCell(o *harness.Oracle, run func(id string) []harness.Event, id string) (map[harness.PropertyID]harness.Verdict, []harness.Event, []harness.Event) {
	r1 := run(id)
	r2 := run(id)
	if admissionBearing[id] {
		return o.EvaluateRuns([][]harness.Event{r1, r2}), r1, r2
	}
	return o.Evaluate(r1), r1, r2
}

// TestE2_1b_CoverageBattery runs B, C and E through the expanded S1–S8
// battery, asserts non-regression against the frozen E1 reference matrix,
// and logs the P1–P8 × S1–S8 matrix plus the delta vs E1 for each
// architecture. D is covered separately (TestE2_1b_CoverageBattery_D)
// because it runs under the pre-registered stochastic clause.
func TestE2_1b_CoverageBattery(t *testing.T) {
	o := harness.NewOracle()
	ref := e1RefS1toS5()

	for _, a := range detArchs() {
		t.Run(a.id, func(t *testing.T) {
			verdicts := map[string]map[harness.PropertyID]harness.Verdict{}
			for _, id := range fullIDs {
				v, r1, r2 := evalCell(o, a.run, id)
				if fmt.Sprint(r1) != fmt.Sprint(r2) {
					t.Errorf("%s: %s is not deterministic (run 1 != run 2)", id, a.id)
				}
				verdicts[id] = v
			}

			// --- Non-regression (acceptance condition) ---
			for _, id := range []string{"S1", "S2", "S3", "S4", "S5"} {
				for _, p := range allProps {
					want := ref[id][p]
					if exp, ok := expectedProtocolDelta[id]; ok && exp == p {
						want = harness.Pass // the one admissible delta
					}
					if verdicts[id][p] != want {
						t.Errorf("REGRESSION %s: %s×%s = %s, E1 reference (adjusted for the P7 protocol fix) wants %s",
							a.id, id, p, verdicts[id][p], want)
					}
				}
			}

			// --- Log the full P1–P8 × S1–S8 matrix ---
			t.Logf("E2.1b — architecture %s (%s) × battery S1–S8 (Property Oracle)", a.id, a.name)
			t.Logf("  %-24s | %-4s | %-4s | %-4s | %-4s | %-4s | %-4s | %-4s | %-4s", "property", "S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8")
			t.Logf("  " + strings.Repeat("-", 84))
			for _, p := range allProps {
				row := fmt.Sprintf("  %-24s", p)
				for _, id := range fullIDs {
					row += fmt.Sprintf(" | %-4s", verdicts[id][p])
				}
				t.Logf(row)
			}

			// --- Delta table vs E1 (S1–S5 only; S6–S8 are new cells) ---
			t.Logf("")
			t.Logf("  Delta vs E1 (S1-S5):")
			for _, id := range []string{"S1", "S2", "S3", "S4", "S5"} {
				for _, p := range allProps {
					before := ref[id][p]
					after := verdicts[id][p]
					mark := "unchanged"
					if before != after {
						mark = "CHANGED"
					}
					if before != after || (expectedProtocolDelta[id] == p) {
						t.Logf("    %s×%s: E1=%s now=%s (%s)", id, p, before, after, mark)
					}
				}
			}
		})
	}
}

// TestE2_1b_CoverageBattery_D runs D through the expanded S1–S8 battery
// under the same pre-registered stochastic clause as the E1 battery
// (N=50, modal verdict; P7 evaluated "in conjunto" over the N runs for
// every scenario — D already did this for S1–S5, so its P7 verdicts do
// not move). It asserts non-regression against D's own frozen E1 matrix
// (adr-evidence/E1_D_findings.md) and reports the distribution for S6–S8.
func TestE2_1b_CoverageBattery_D(t *testing.T) {
	const N = 50
	o := harness.NewOracle()

	// D's own frozen E1 reference (adr-evidence/E1_D_findings.md — modal
	// verdict over N=50, all cells point-mass except S4×P8).
	u, p := harness.Unknown, harness.Pass
	dRef := map[string]map[harness.PropertyID]harness.Verdict{
		"S1": {harness.P1Auditability: p, harness.P2Authorization: u, harness.P3FailureBlocks: p, harness.P4Recoverable: p, harness.P5AsyncTolerant: u, harness.P6DomainFree: u, harness.P7Determinism: p, harness.P8Isolation: u},
		"S2": {harness.P1Auditability: p, harness.P2Authorization: p, harness.P3FailureBlocks: u, harness.P4Recoverable: u, harness.P5AsyncTolerant: u, harness.P6DomainFree: u, harness.P7Determinism: p, harness.P8Isolation: u},
		"S3": {harness.P1Auditability: p, harness.P2Authorization: u, harness.P3FailureBlocks: u, harness.P4Recoverable: u, harness.P5AsyncTolerant: p, harness.P6DomainFree: u, harness.P7Determinism: p, harness.P8Isolation: u},
		"S4": {harness.P1Auditability: p, harness.P2Authorization: u, harness.P3FailureBlocks: u, harness.P4Recoverable: u, harness.P5AsyncTolerant: u, harness.P6DomainFree: u, harness.P7Determinism: u, harness.P8Isolation: p}, // D's S4×P8 modal is PASS (E1 finding; environment/GOMAXPROCS-sensitive per Fase 5B Parte 0)
		"S5": {harness.P1Auditability: p, harness.P2Authorization: u, harness.P3FailureBlocks: u, harness.P4Recoverable: u, harness.P5AsyncTolerant: u, harness.P6DomainFree: u, harness.P7Determinism: p, harness.P8Isolation: u},
	}

	dist := map[string]map[harness.PropertyID]map[harness.Verdict]int{}
	traces := map[string][][]harness.Event{}
	for _, id := range fullIDs {
		cmds, ok := dCommands[id]
		if !ok {
			t.Fatalf("no command script for scenario %s", id)
		}
		dist[id] = map[harness.PropertyID]map[harness.Verdict]int{}
		for i := 0; i < N; i++ {
			tr, ok := newActorModelD(nil).run(cmds)
			if !ok {
				t.Fatalf("%s run %d: D.run() timed out — implementation failure (deadlock), NOT a result", id, i)
			}
			traces[id] = append(traces[id], tr)
			v := o.Evaluate(tr)
			for _, pr := range allProps {
				if dist[id][pr] == nil {
					dist[id][pr] = map[harness.Verdict]int{}
				}
				dist[id][pr][v[pr]]++
			}
		}
	}

	rep := map[string]map[harness.PropertyID]harness.Verdict{}
	var highVar []string
	for _, id := range fullIDs {
		rep[id] = map[harness.PropertyID]harness.Verdict{}
		for _, pr := range allProps {
			if pr == harness.P7Determinism {
				rep[id][pr] = o.EvaluateRuns(traces[id])[pr]
				continue
			}
			modal, maxC := modalOf(dist[id][pr])
			rep[id][pr] = modal
			if float64(maxC)/float64(N) < 0.5 {
				highVar = append(highVar, fmt.Sprintf("%s×%s", id, pr))
			}
		}
	}

	// --- Non-regression against D's own frozen E1 matrix ---
	for _, id := range []string{"S1", "S2", "S3", "S4", "S5"} {
		for _, pr := range allProps {
			if rep[id][pr] != dRef[id][pr] {
				t.Errorf("REGRESSION D: %s×%s = %s, E1 reference wants %s", id, pr, rep[id][pr], dRef[id][pr])
			}
		}
	}

	t.Logf("E2.1b — architecture D (Actor Model) × battery S1–S8, N=%d (stochastic clause)", N)
	t.Logf("  %-24s | %-4s | %-4s | %-4s | %-4s | %-4s | %-4s | %-4s | %-4s", "property", "S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8")
	t.Logf("  " + strings.Repeat("-", 84))
	for _, pr := range allProps {
		row := fmt.Sprintf("  %-24s", pr)
		for _, id := range fullIDs {
			row += fmt.Sprintf(" | %-4s", rep[id][pr])
		}
		t.Logf(row)
	}
	t.Logf("")
	t.Logf("  High-variance cells (modal freq < 50%%): %v", highVar)
	t.Logf("")
	t.Logf("  Distribution detail, S6-S8:")
	for _, id := range []string{"S6", "S7", "S8"} {
		for _, pr := range allProps {
			c := dist[id][pr]
			if c[harness.Pass] > 0 && c[harness.Fail] > 0 {
				t.Logf("    %s×%s: PASS=%d FAIL=%d UNKNOWN=%d (VARIANCE)", id, pr, c[harness.Pass], c[harness.Fail], c[harness.Unknown])
			}
		}
	}
}
