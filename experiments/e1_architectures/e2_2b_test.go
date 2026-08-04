// E2.2b — S9 (cross-capability effect scoping) and S10 (effect after
// Fail+Rollback without a fresh Authorize), per the design frozen in the
// E2.2a conversation and DECISÃO A (rollback does not restore
// authorization — the semantics already fixed by the frozen P3 criterion,
// oracle.go:p3, where blocked[cap] is cleared only by
// AuthorizationGranted, never by RollbackExecuted).
//
// S10 is a CONFORMANCE test, not a divergence probe: it measures whether
// each architecture's already-frozen rollback/effect-gate logic (unmodified
// here) satisfies the contract semantics fixed by DECISÃO A. The verdict is
// reported, not asserted to match a prediction — if execution disagrees
// with the E2.2a code-reading prediction, that is the finding.
//
// S11 was REJECTED by the Arquiteto (low value; requires a parameterized
// admission policy in the shared battery, a protocol change not justified
// here) and is not implemented.
//
// harness/oracle.go, harness/events.go, P1–P8, kernel/, models_comp/ and
// the core logic of B/C/D/E (log, projection, pipeline, state machine,
// guards, rollback) are untouched — only new command-map entries were
// added per architecture file.
package e1_architectures_test

import (
	"fmt"
	"testing"

	"gaep-core-poc/harness"
)

func init() {
	// S9 and S10 both contain Admit: fold them into the 2-run P7 protocol
	// established in E2.1b (adr-evidence/E2_1a_scenario_design.md).
	admissionBearing["S9"] = true
	admissionBearing["S10"] = true
}

var newIDs = []string{"S9", "S10"}

// TestE2_2b_CoverageBattery runs B, C and E through S9 and S10, logs the
// P1–P8 matrix, and reports (does not assert) S10's P3 conformance verdict
// per architecture.
func TestE2_2b_CoverageBattery(t *testing.T) {
	o := harness.NewOracle()

	for _, a := range detArchs() {
		t.Run(a.id, func(t *testing.T) {
			verdicts := map[string]map[harness.PropertyID]harness.Verdict{}
			for _, id := range newIDs {
				v, r1, r2 := evalCell(o, a.run, id)
				if fmt.Sprint(r1) != fmt.Sprint(r2) {
					t.Errorf("%s: %s is not deterministic (run 1 != run 2)", id, a.id)
				}
				verdicts[id] = v
			}

			t.Logf("E2.2b — architecture %s (%s) x {S9, S10} (Property Oracle)", a.id, a.name)
			t.Logf("  %-24s | %-4s | %-4s", "property", "S9", "S10")
			for _, p := range allProps {
				row := fmt.Sprintf("  %-24s", p)
				for _, id := range newIDs {
					row += fmt.Sprintf(" | %-4s", verdicts[id][p])
				}
				t.Logf(row)
			}

			// S10 conformance report (E2.2a prediction, DECISÃO A): the
			// contract-conformant behavior is P3 = PASS (effect blocked,
			// re-authorization required after rollback). Reported, not
			// asserted — a FAIL here is a real conformance finding, not a
			// test failure.
			p3 := verdicts["S10"][harness.P3FailureBlocks]
			conformant := p3 == harness.Pass
			t.Logf("  S10 P3 = %s -> %s with DECISÃO A (rollback does not restore authorization)",
				p3, map[bool]string{true: "CONFORMANT", false: "NON-CONFORMANT"}[conformant])

			// S9 sanity: P2 must be PASS in the nominal run (both effects
			// are legitimate; the illegal c2 attempt must be blocked by
			// every conformant architecture, per the E2.1b guard-code
			// inspection). If this ever fails, it means an architecture
			// leaked authorization across capabilities — report, don't
			// silently pass.
			if verdicts["S9"][harness.P2Authorization] != harness.Pass {
				t.Errorf("S9: P2 = %s, want PASS (cross-capability authorization leak?) — real finding, not expected", verdicts["S9"][harness.P2Authorization])
			}
		})
	}
}

// TestE2_2b_CoverageBattery_D runs D through S9 and S10 under the
// pre-registered stochastic clause (N=50, modal verdict), same protocol as
// E2.1b's D coverage test.
func TestE2_2b_CoverageBattery_D(t *testing.T) {
	const N = 50
	o := harness.NewOracle()

	dist := map[string]map[harness.PropertyID]map[harness.Verdict]int{}
	traces := map[string][][]harness.Event{}
	for _, id := range newIDs {
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
	for _, id := range newIDs {
		rep[id] = map[harness.PropertyID]harness.Verdict{}
		for _, pr := range allProps {
			if pr == harness.P7Determinism {
				rep[id][pr] = o.EvaluateRuns(traces[id])[pr]
				continue
			}
			modal, maxC := modalOf(dist[id][pr])
			rep[id][pr] = modal
			if float64(maxC)/float64(N) < 0.5 {
				highVar = append(highVar, fmt.Sprintf("%s x %s", id, pr))
			}
		}
	}

	t.Logf("E2.2b — architecture D (Actor Model) x {S9, S10}, N=%d (stochastic clause)", N)
	t.Logf("  %-24s | %-4s | %-4s", "property", "S9", "S10")
	for _, pr := range allProps {
		row := fmt.Sprintf("  %-24s", pr)
		for _, id := range newIDs {
			row += fmt.Sprintf(" | %-4s", rep[id][pr])
		}
		t.Logf(row)
	}
	t.Logf("  High-variance cells (modal freq < 50%%): %v", highVar)

	// Point-mass check: S9/S10 are single/sequential-actor scripts
	// (S9 dispatches c1 fully before c2's Admit is even sent; S10 is a
	// single actor) — same expectation as S1-S3/S5-S7 in prior phases.
	for _, id := range newIDs {
		for _, pr := range allProps {
			c := dist[id][pr]
			distinct := 0
			for _, v := range []harness.Verdict{harness.Pass, harness.Fail, harness.Unknown} {
				if c[v] > 0 {
					distinct++
				}
			}
			if distinct != 1 {
				t.Logf("  NOTE (not asserted as bug): %s x %s is NOT point-mass (%d distinct verdicts over N=%d): PASS=%d FAIL=%d UNKNOWN=%d",
					id, pr, distinct, N, c[harness.Pass], c[harness.Fail], c[harness.Unknown])
			}
		}
	}

	p3 := rep["S10"][harness.P3FailureBlocks]
	conformant := p3 == harness.Pass
	t.Logf("  S10 P3 (modal) = %s -> %s with DECISÃO A", p3, map[bool]string{true: "CONFORMANT", false: "NON-CONFORMANT"}[conformant])
}

// TestE2_2b_S9Adversarial evaluates a synthetic trace (NOT run against any
// architecture) where c2's illegal effect attempt is forced to actually
// surface as ExternalEffect before c2's AuthorizationGranted, while c1 is
// legitimately authorized and effecting. This is the first verification,
// in the whole project, that oracle.go:p2's FAIL branch actually fires.
func TestE2_2b_S9Adversarial(t *testing.T) {
	o := harness.NewOracle()
	adv := []harness.Event{
		{Seq: 1, Kind: harness.EvAdmissionRequested},
		{Seq: 2, Kind: harness.EvAdmissionAccepted},
		{Seq: 3, Kind: harness.EvAuditWritten},
		{Seq: 4, Kind: harness.EvAuthorizationGranted, Cap: "c1"},
		{Seq: 5, Kind: harness.EvAuditWritten},
		{Seq: 6, Kind: harness.EvExternalEffect, Cap: "c1"}, // legitimate
		{Seq: 7, Kind: harness.EvAuditWritten},
		{Seq: 8, Kind: harness.EvAdmissionRequested},
		{Seq: 9, Kind: harness.EvAdmissionAccepted},
		{Seq: 10, Kind: harness.EvAuditWritten},
		{Seq: 11, Kind: harness.EvExternalEffect, Cap: "c2"}, // VIOLATION: precedes c2's AuthorizationGranted
		{Seq: 12, Kind: harness.EvAuditWritten},
		{Seq: 13, Kind: harness.EvAuthorizationGranted, Cap: "c2"},
		{Seq: 14, Kind: harness.EvAuditWritten},
		{Seq: 15, Kind: harness.EvExternalEffect, Cap: "c2"},
		{Seq: 16, Kind: harness.EvAuditWritten},
	}
	v := o.Evaluate(adv)
	if v[harness.P2Authorization] != harness.Fail {
		t.Errorf("S9-adv: expected P2 = FAIL (first synthetic verification of p2's FAIL branch), got %s", v[harness.P2Authorization])
	}
	t.Logf("S9-adv (synthetic, not run against B/C/D/E): P2 = %s", v[harness.P2Authorization])
}
