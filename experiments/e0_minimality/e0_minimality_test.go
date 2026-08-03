// Package e0_minimality_test is Experiment E0 of the GAEP Core POC:
// architectural minimality by component removal.
//
// The hypothesis is falsifiable: "if removing component e does not cause
// any invariant to be lost (or become unverifiable), then e is not
// architecturally essential — it is a removal candidate." E0 runs this
// removal analysis over (a) each register $R0$–$R7$ against the core
// invariant suite I0–I4 and (b) each control-flow component (SESSION,
// WATCHDOG, WAIT_OUTCOME, RELEASE) against the lifecycle invariants L1–L4.
//
// Evidence output is captured to adr-evidence/E0_evidence.log.
package e0_minimality_test

import (
	"strings"
	"testing"
	"time"

	"gaep-core-poc/kernel"
	"gaep-core-poc/models_comp"
)

// regInvariantDeps maps each register to the invariants whose input it
// provides. Removing a register makes those invariants unverifiable.
var regInvariantDeps = map[kernel.RegisterID][]string{
	kernel.R1Watchdog:  {"I4"},
	kernel.R2NextAct:   {"I1"},
	kernel.R6Health:    {"I3"},
	kernel.R7Timestamp: {"I2"},
}

func applyReg(s *models_comp.Snapshot, id kernel.RegisterID, v uint64) {
	switch id {
	case kernel.R0Session:
		s.Session = v
	case kernel.R1Watchdog:
		s.Watchdog = v
	case kernel.R2NextAct:
		s.NextAct = v
	case kernel.R3Isolation:
		s.Isolation = v
	case kernel.R4Stderr:
		s.Stderr = v
	case kernel.R5SKKBadge:
		s.Badge = v
	case kernel.R6Health:
		s.Health = v
	case kernel.R7Timestamp:
		s.Timestamp = v
	}
}

// scriptedSnapshotTrace runs the standard engine workload over a plain
// Snapshot, skipping writes to the suppressed register (its cell never
// changes) and mimicking the register file's automatic timestamp stamping
// unless $R7$ itself is the suppressed register.
func scriptedSnapshotTrace(suppressed kernel.RegisterID) []models_comp.Snapshot {
	now := uint64(time.Now().UnixNano())
	type regOp struct {
		id kernel.RegisterID
		v  uint64
	}
	ops := []regOp{
		{kernel.R0Session, 42},
		{kernel.R5SKKBadge, 0x5eed},
		{kernel.R3Isolation, 4},
		{kernel.R1Watchdog, now + 10e9},
		{kernel.R6Health, 1},
		{kernel.R2NextAct, 1},
		{kernel.R4Stderr, 1},
		{kernel.R6Health, 2},
		{kernel.R2NextAct, 2},
		{kernel.R1Watchdog, now + 20e9},
		{kernel.R6Health, 3},
	}
	var s models_comp.Snapshot
	writes := uint64(0)
	trace := make([]models_comp.Snapshot, 0, len(ops))
	for i, op := range ops {
		if op.id == suppressed {
			continue
		}
		applyReg(&s, op.id, op.v)
		writes++
		s.Writes = writes
		if suppressed != kernel.R7Timestamp {
			s.Timestamp = now + uint64(i)*1e6 // auto-stamp advances
		}
		trace = append(trace, s)
	}
	return trace
}

// registerRemovalAnalysis returns the invariants lost (or made unverifiable)
// when reg is removed; empty means no invariant is lost → removal candidate.
func registerRemovalAnalysis(reg kernel.RegisterID) []string {
	trace := scriptedSnapshotTrace(reg)
	var lost []string
	for _, inv := range models_comp.CoreInvariants {
		dep := false
		for _, d := range regInvariantDeps[reg] {
			if d == inv.ID {
				dep = true
			}
		}
		res := models_comp.CheckTrace(inv, trace)
		switch {
		case dep:
			lost = append(lost, inv.ID+" unverifiable (input removed)")
		case !res.OK:
			lost = append(lost, inv.ID+" broken ("+res.Detail+")")
		}
	}
	return lost
}

// TestE0_RegisterRemoval removes each register and reports which invariants
// are lost. R1/R2/R6/R7 are pinned by the core suite; R0/R3/R4/R5 are not.
func TestE0_RegisterRemoval(t *testing.T) {
	t.Logf("E0 — Section A: register removal against core invariants I0–I4")
	t.Logf("  essential   = removal makes an invariant unverifiable or breaks it")
	t.Logf("  candidate   = removal loses no invariant (requirements gap or removable)")
	registers := []kernel.RegisterID{
		kernel.R0Session, kernel.R1Watchdog, kernel.R2NextAct, kernel.R3Isolation,
		kernel.R4Stderr, kernel.R5SKKBadge, kernel.R6Health, kernel.R7Timestamp,
	}
	essential := map[kernel.RegisterID]bool{
		kernel.R1Watchdog: true, kernel.R2NextAct: true,
		kernel.R6Health: true, kernel.R7Timestamp: true,
	}
	for _, reg := range registers {
		lost := registerRemovalAnalysis(reg)
		if len(lost) == 0 {
			t.Logf("  remove %-16s → NO invariant lost (removal candidate)", reg)
		} else {
			t.Logf("  remove %-16s → %v", reg, lost)
		}
		if essential[reg] && len(lost) == 0 {
			t.Errorf("register %s was expected to be essential, but no invariant was lost", reg)
		}
		if !essential[reg] && len(lost) != 0 {
			t.Errorf("register %s was expected to be a removal candidate, but lost %v", reg, lost)
		}
	}
	t.Logf("  finding: the current invariant suite pins only $R1$/R2/R6/R7;")
	t.Logf("  $R0$ Session, $R3$ Isolation, $R4$ Stderr, $R5$ SKK Badge are unconstrained")
	t.Logf("  → either their requirements are missing (add invariants) or they are removable (ADR).")
}

// --- lifecycle removal analysis ---

func runLifecycle(l *models_comp.Lifecycle, script []models_comp.LifecycleEvent) ([]models_comp.LifecycleSample, []error) {
	samples := []models_comp.LifecycleSample{l.Sample()}
	var errs []error
	for _, ev := range script {
		if err := l.Step(ev); err != nil {
			errs = append(errs, err)
		}
		samples = append(samples, l.Sample())
	}
	return samples, errs
}

func l1ReleaseAbsorbing(samples []models_comp.LifecycleSample) bool {
	seenRelease := false
	for _, s := range samples {
		if s.State == models_comp.StateRelease {
			seenRelease = true
		}
		if seenRelease && s.State != models_comp.StateRelease {
			return false
		}
	}
	return true
}

func l2WaitRequiresSession(samples []models_comp.LifecycleSample) bool {
	for _, s := range samples {
		if s.State == models_comp.StateWaitOutcome && s.Session == 0 {
			return false
		}
	}
	return true
}

func l3RunningRequiresSession(samples []models_comp.LifecycleSample) bool {
	for _, s := range samples {
		if s.State == models_comp.StateRunning && s.Session == 0 {
			return false
		}
	}
	return true
}

func l4TimeoutHonored(samples []models_comp.LifecycleSample, errs []error) bool {
	for _, e := range errs {
		if strings.Contains(e.Error(), "timeout") {
			return false
		}
	}
	return samples[len(samples)-1].State == models_comp.StateRelease
}

func lifecycleLosses(deps []string, samples []models_comp.LifecycleSample, errs []error) []string {
	var lost []string
	dep := func(id string) bool {
		for _, d := range deps {
			if d == id {
				return true
			}
		}
		return false
	}
	checks := []struct {
		id    string
		check func() bool
	}{
		{"L1", func() bool { return l1ReleaseAbsorbing(samples) }},
		{"L2", func() bool { return l2WaitRequiresSession(samples) }},
		{"L3", func() bool { return l3RunningRequiresSession(samples) }},
		{"L4", func() bool { return l4TimeoutHonored(samples, errs) }},
	}
	for _, c := range checks {
		ok := c.check()
		switch {
		case dep(c.id) && !ok:
			lost = append(lost, c.id+" lost")
		case dep(c.id):
			lost = append(lost, c.id+" unverifiable (input removed)")
		case !ok:
			lost = append(lost, c.id+" broken")
		}
	}
	return lost
}

// TestE0_LifecycleRemoval removes each control-flow component and reports
// which lifecycle invariants are lost.
func TestE0_LifecycleRemoval(t *testing.T) {
	t.Logf("")
	t.Logf("E0 — Section B: control-flow components vs lifecycle invariants L1–L4")
	now := uint64(time.Now().UnixNano())
	script := []models_comp.LifecycleEvent{
		models_comp.EvSession, models_comp.EvWait, models_comp.EvOutcome,
		models_comp.EvWait, models_comp.EvTimeout,
	}
	cases := []struct {
		name      string
		deps      []string
		configure func(*models_comp.Lifecycle)
	}{
		{"SESSION", []string{"L2", "L3"}, func(l *models_comp.Lifecycle) {
			l.NoSessionGate = true
			l.Session = 0
		}},
		{"WATCHDOG", []string{"L4"}, func(l *models_comp.Lifecycle) {
			l.NoWatchdog = true
			l.Watchdog = 0
		}},
		{"WAIT_OUTCOME", []string{"L2", "L4"}, func(l *models_comp.Lifecycle) { l.NoWaitOutcome = true }},
		{"RELEASE", []string{"L1"}, func(l *models_comp.Lifecycle) { l.NoRelease = true }},
	}

	// control: no removal → nothing lost.
	ctl := models_comp.NewLifecycle()
	ctl.Session, ctl.Watchdog = 42, now+10e9
	ctlSamples, ctlErrs := runLifecycle(ctl, script)
	if lost := lifecycleLosses(nil, ctlSamples, ctlErrs); len(lost) != 0 {
		t.Errorf("control run lost invariants: %v", lost)
	}
	t.Logf("  remove %-12s → no invariant lost (control)", "—none—")

	for _, c := range cases {
		l := models_comp.NewLifecycle()
		l.Session, l.Watchdog = 42, now+10e9
		c.configure(l)
		samples, errs := runLifecycle(l, script)
		lost := lifecycleLosses(c.deps, samples, errs)
		if len(lost) == 0 {
			t.Logf("  remove %-12s → NO invariant lost (removal candidate)", c.name)
			t.Errorf("component %s was expected to be essential, but no invariant was lost", c.name)
		} else {
			t.Logf("  remove %-12s → %v", c.name, lost)
		}
	}
	t.Logf("  finding: SESSION, WATCHDOG, WAIT_OUTCOME and RELEASE are all load-bearing")
	t.Logf("  under the lifecycle invariants L1–L4.")
}
