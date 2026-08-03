package harness_test

import (
	"testing"

	"gaep-core-poc/harness"
)

func TestScenarioVerdicts(t *testing.T) {
	o := harness.NewOracle()
	cases := []struct {
		scenario harness.Scenario
		want     map[harness.PropertyID]harness.Verdict
	}{
		{harness.ScenarioFailMid, map[harness.PropertyID]harness.Verdict{
			harness.P1Auditability:  harness.Pass,
			harness.P2Authorization: harness.Unknown, // no external effect in S1
			harness.P3FailureBlocks: harness.Pass,
			harness.P4Recoverable:   harness.Pass,
			harness.P5AsyncTolerant: harness.Unknown, // no timeout in S1
			harness.P8Isolation:     harness.Unknown, // single capability in S1
		}},
		{harness.ScenarioEffectBeforeAuth, map[harness.PropertyID]harness.Verdict{
			harness.P2Authorization: harness.Fail, // effect before authorization
		}},
		{harness.ScenarioLateSuccessIgnored, map[harness.PropertyID]harness.Verdict{
			harness.P5AsyncTolerant: harness.Pass, // late success ignored
		}},
		{harness.ScenarioConcurrent, map[harness.PropertyID]harness.Verdict{
			harness.P8Isolation: harness.Fail, // interleaving observed
		}},
	}
	for _, c := range cases {
		got := o.Evaluate(c.scenario.Trace)
		for p, want := range c.want {
			if got[p] != want {
				t.Errorf("%s: %s = %s, want %s", c.scenario.ID, p, got[p], want)
			}
		}
	}
}

func TestP7RequiresRepeatedRuns(t *testing.T) {
	o := harness.NewOracle()
	run := harness.ScenarioDeterministicAdmission.Trace

	// Single run: cannot establish determinism.
	if v := o.Evaluate(run)[harness.P7Determinism]; v != harness.Unknown {
		t.Errorf("P7 with a single run = %s, want UNKNOWN", v)
	}

	// Two identical runs: PASS.
	got := o.EvaluateRuns([][]harness.Event{run, run})
	if v := got[harness.P7Determinism]; v != harness.Pass {
		t.Errorf("P7 with two identical runs = %s, want PASS", v)
	}

	// Diverging runs: FAIL.
	divergent := make([]harness.Event, len(run))
	copy(divergent, run)
	divergent[1] = harness.Event{Seq: 2, Kind: harness.EvAdmissionRejected, Cap: "c1"}
	got2 := o.EvaluateRuns([][]harness.Event{run, divergent})
	if v := got2[harness.P7Determinism]; v != harness.Fail {
		t.Errorf("P7 with diverging runs = %s, want FAIL", v)
	}
}

func TestAdversarialVariantFails(t *testing.T) {
	o := harness.NewOracle()
	// Late success APPLIED (adversarial S3): P5 must flip to FAIL.
	v := o.Evaluate(harness.ScenarioLateSuccessIgnored.Adversarial)[harness.P5AsyncTolerant]
	if v != harness.Fail {
		t.Errorf("adversarial S3: P5 = %s, want FAIL", v)
	}
}
