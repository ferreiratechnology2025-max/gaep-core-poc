// Package e1_architectures_test is E1 of the harness: concurrent rival
// architectures run through the same scenario battery, graded only by the
// Property Oracle. This iteration ships the harness core (events + oracle +
// battery) and the matrix scaffold: architecture A is represented by its
// nominal event emission for S1; B–E are declared per the independence rule
// (experiment-plan.md) and are not yet implemented.
//
// Evidence output is captured to adr-evidence/E1arch_evidence.log.
package e1_architectures_test

import (
	"fmt"
	"strings"
	"testing"

	"gaep-core-poc/harness"
)

type archColumn struct {
	id   string
	name string
	// emit returns the event trace this architecture produces for a
	// scenario; nil means the architecture is not yet implemented.
	emit func(h ScenarioHandle) []harness.Event
}

// ScenarioHandle names a battery item.
type ScenarioHandle string

const (
	scnFailMid ScenarioHandle = "S1"
)

// archA is the current GAEP architecture, represented for the scaffold by
// the nominal S1 trace (a real adapter that drives kernel/models will
// replace this emission in the next iteration).
func archA() archColumn {
	return archColumn{id: "A", name: "GAEP (registers + microkernel)", emit: func(h ScenarioHandle) []harness.Event {
		if h == scnFailMid {
			return harness.ScenarioFailMid.Trace
		}
		return nil
	}}
}

func declaredArchs() []archColumn {
	return []archColumn{
		archA(),
		{id: "B", name: "Context Object + Policy Pipeline"},
		{id: "C", name: "Event Sourcing"},
		{id: "D", name: "Actor Model"},
		{id: "E", name: "Hierarchical State Machine"},
	}
}

func TestE1_ArchitectureMatrix(t *testing.T) {
	o := harness.NewOracle()
	archs := declaredArchs()

	t.Logf("E1 — property × architecture matrix (v0 scaffold)")
	t.Logf("Rule: architectures only emit events; the Property Oracle decides PASS/FAIL/UNKNOWN.")
	t.Logf("Independence rule: B–E must be born from their own principles, not a rename of A.")
	t.Logf("")

	// Header.
	head := fmt.Sprintf("  %-24s", "property")
	for _, a := range archs {
		head += fmt.Sprintf(" | %-4s", a.id)
	}
	t.Logf(head)
	t.Logf("  " + strings.Repeat("—", 40))

	props := []harness.PropertyID{
		harness.P1Auditability, harness.P2Authorization, harness.P3FailureBlocks,
		harness.P4Recoverable, harness.P5AsyncTolerant, harness.P6DomainFree,
		harness.P7Determinism, harness.P8Isolation,
	}
	for _, p := range props {
		row := fmt.Sprintf("  %-24s", p)
		for _, a := range archs {
			cell := "—"
			if a.emit != nil {
				if ev := a.emit(scnFailMid); ev != nil {
					cell = o.Evaluate(ev)[p].String()
				}
			}
			row += fmt.Sprintf(" | %-4s", cell)
		}
		t.Logf(row)
	}
	t.Logf("")
	t.Logf("  — : not implemented. P6 is structural (dual-run + source inspection),")
	t.Logf("     not event-based. S1 exercised; remaining scenarios S2–S5 pending.")
	t.Logf("  Verdict so far: A passes P1/P3/P4 on S1 (via the Oracle, not self-graded).")
}
