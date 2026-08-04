// Fase 5B, Parte 0 — concurrency verification of D's S4×P8 result.
//
// Phases 3-5 observed D at 100% PASS on S4×P8 over 500 runs (10 executions
// of the N=50 battery). That result alone does not distinguish "genuine
// mailbox isolation between actors" from "this environment's scheduler
// happens to drain c1's mailbox before scheduling c2, every time". This
// file adds a GOMAXPROCS variation probe to tell those apart: it does NOT
// modify D's implementation, the frozen S4 command script or the
// pre-registered stochastic clause (N=50, modal verdict) — it only reruns
// the same S4 scenario under two parallelism regimes and reports the
// resulting P8 distribution. No causal conclusion is drawn here: the
// Observed/Hypothesized Cause/Evidence Required/Status block below states
// the interpretation as hypothesis, per the pre-registration discipline
// already used for S4×P8 in adr-evidence/E1_D_findings.md.
package e1_architectures_test

import (
	"runtime"
	"testing"

	"gaep-core-poc/harness"
)

// TestE1_ArchitectureD_ConcurrencyCheck runs D's S4 scenario N=50 times
// under GOMAXPROCS=1 and GOMAXPROCS=4 and compares the P8 distributions.
func TestE1_ArchitectureD_ConcurrencyCheck(t *testing.T) {
	const N = 50
	o := harness.NewOracle()
	cmds := dCommands["S4"]

	measure := func(procs int) map[harness.Verdict]int {
		prev := runtime.GOMAXPROCS(procs)
		defer runtime.GOMAXPROCS(prev)
		counts := map[harness.Verdict]int{}
		for i := 0; i < N; i++ {
			tr, ok := newActorModelD(nil).run(cmds)
			if !ok {
				t.Fatalf("GOMAXPROCS=%d run %d: D.run() timed out — implementation failure (deadlock), NOT a result", procs, i)
			}
			v := o.Evaluate(tr)[harness.P8Isolation]
			counts[v]++
		}
		return counts
	}

	d1 := measure(1)
	d4 := measure(4)

	t.Logf("E1 — D (Actor Model) × S4×P8 concurrency probe (Fase 5B, Parte 0)")
	t.Logf("N=%d per regime. Does not alter D's implementation, the S4 command script, or the pre-registered stochastic clause.", N)
	t.Logf("")
	t.Logf("  GOMAXPROCS=1: PASS=%d FAIL=%d UNKNOWN=%d (freq PASS %.0f%%)",
		d1[harness.Pass], d1[harness.Fail], d1[harness.Unknown], float64(d1[harness.Pass])/float64(N)*100)
	t.Logf("  GOMAXPROCS=4: PASS=%d FAIL=%d UNKNOWN=%d (freq PASS %.0f%%)",
		d4[harness.Pass], d4[harness.Fail], d4[harness.Unknown], float64(d4[harness.Pass])/float64(N)*100)

	changed := d1[harness.Pass] != d4[harness.Pass] || d1[harness.Fail] != d4[harness.Fail]
	t.Logf("  Distribution changed between regimes (GOMAXPROCS=1 vs 4): %v", changed)
	t.Logf("")
	t.Logf("Observed Results")
	t.Logf("  GOMAXPROCS=1 → S4×P8 PASS=%d/%d FAIL=%d/%d", d1[harness.Pass], N, d1[harness.Fail], N)
	t.Logf("  GOMAXPROCS=4 → S4×P8 PASS=%d/%d FAIL=%d/%d", d4[harness.Pass], N, d4[harness.Fail], N)
	t.Logf("Hypothesized Cause")
	if changed {
		t.Logf("  Variance appeared at higher parallelism. Consistent with: the original 100%% PASS (phases 3-5,")
		t.Logf("  single-procs-unspecified runs) depended on scheduler behavior rather than solely on per-actor")
		t.Logf("  mailbox isolation. Also consistent with GOMAXPROCS=1 itself constraining interleaving opportunity")
		t.Logf("  (fewer OS-thread-level preemption points), independent of what >1 procs would show on another host.")
	} else {
		t.Logf("  No variance appeared even at GOMAXPROCS=4. Consistent with (but not proof of) genuine per-actor")
		t.Logf("  mailbox isolation holding regardless of scheduler parallelism. Equally consistent with this")
		t.Logf("  specific command script/workload not giving the scheduler an opportunity to interleave c1/c2")
		t.Logf("  delivery even at 4 procs (e.g. too few actor goroutines, too-fast handlers, buffered mailboxes")
		t.Logf("  masking any interleaving that occurs at the channel level).")
	}
	t.Logf("Evidence Required")
	t.Logf("  (a) Repetition on other environments/schedulers (e.g. Linux, higher GOMAXPROCS, GOMAXPROCS>NumCPU);")
	t.Logf("  (b) an independent probe that forces/detects actor-level interleaving directly (e.g. instrumented")
	t.Logf("  handler with an injected delay) to confirm the scheduler CAN interleave c1/c2 mailboxes here at all;")
	t.Logf("  (c) Fase 5B Parte 1 (bateria de A) to add a fifth data point to the discriminative-power table.")
	t.Logf("Status")
	t.Logf("  HIPÓTESE. Not established as fact by this test; no causal conclusion is drawn.")
}
