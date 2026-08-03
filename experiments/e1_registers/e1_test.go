// Package e1_registers_test is Experiment E1 of the GAEP Core POC.
//
// It runs the two candidate representations of governance state under
// identical concurrent workloads:
//
//   - Scenario 1: direct pointer mutation on the monolithic SessionState
//     control struct (models_comp) — no synchronization at all.
//   - Scenario 2: mutation through the kernel.RegisterSet, whose locked
//     read-modify-write primitive "repels" concurrent mutation of the
//     $R0$–$R7$ register file.
//
// Both scenarios run the same workload: `workers` goroutines, each
// repeating read-modify-write of the next-action cell. Scenario 1 uses a
// two-phase barrier so every worker reads the same value before any worker
// writes, which makes the lost-update corruption deterministic and
// repeatable: exactly one of `workers` writes survives per round. Under
// `go test -race`, the same scenario is additionally reported as a data
// race on SessionState.NextActSeq (intentional — that is the point).
//
// Evidence output is captured to adr-evidence/E1_evidence.log.
package e1_registers_test

import (
	"sync"
	"testing"
	"time"

	"gaep-core-poc/kernel"
	"gaep-core-poc/models_comp"
)

const (
	workers    = 8
	rounds     = 2000
	initialSeq = 1000
)

// racyRound runs one unsynchronized read-modify-write round over the
// NextActSeq field: all `n` workers read the current value first, then all
// write value+1. Because every read precedes every write (two-phase
// barrier), exactly n-1 updates are lost per round.
func racyRound(s *models_comp.SessionState, n int) {
	readDone := make(chan struct{}, n)
	release := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := s.NextActSeq // read
			readDone <- struct{}{}
			<-release            // wait until every worker has read
			s.NextActSeq = v + 1 // write
		}()
	}
	for i := 0; i < n; i++ {
		<-readDone
	}
	close(release)
	wg.Wait()
}

// TestScenario1_DirectPointerMutationOnControlStruct demonstrates that
// concurrent direct pointer mutation of the monolithic SessionState loses
// updates: only one of `workers` writes survives each round, so the final
// value falls far short of the correct total. The outcome is asserted
// deterministically so the experiment is repeatable.
func TestScenario1_DirectPointerMutationOnControlStruct(t *testing.T) {
	s := &models_comp.SessionState{NextActSeq: initialSeq}

	for r := 0; r < rounds; r++ {
		racyRound(s, workers)
	}

	expected := uint64(initialSeq + workers*rounds) // correct outcome
	predicted := uint64(initialSeq + rounds)        // exactly 1 of `workers` writes survives per round
	got := s.NextActSeq
	lost := expected - got

	t.Logf("Scenario 1 — direct pointer mutation on monolithic SessionState")
	t.Logf("  workload      : %d workers × %d rounds of unsynchronized read-modify-write", workers, rounds)
	t.Logf("  expected final: %d (correct, no lost updates)", expected)
	t.Logf("  observed final: %d", got)
	t.Logf("  lost updates  : %d of %d", lost, uint64(workers*rounds))
	t.Logf("  verdict       : control struct silently lost updates — governance state corrupted")

	if got != predicted {
		t.Fatalf("unexpected outcome: got %d, want deterministic corruption at %d", got, predicted)
	}
}

// TestScenario2_RegisterMutationRepulsion runs the same workload through
// the thread-safe RegisterSet. The Mutate primitive serializes the
// read-modify-write under one lock, so every update is applied and the
// final values are exact — no lost updates, no torn values.
func TestScenario2_RegisterMutationRepulsion(t *testing.T) {
	rs := kernel.NewRegisterSet()
	rs.Write(kernel.R0Session, 42)
	rs.Write(kernel.R1Watchdog, uint64(time.Now().UnixNano()+10e9))
	rs.Write(kernel.R2NextAct, initialSeq)
	rs.Write(kernel.R3Isolation, 4) // strict isolation
	rs.Write(kernel.R5SKKBadge, 0x5eed)
	rs.Write(kernel.R6Health, 1) // healthy

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				rs.Next(kernel.R2NextAct) // dedicated sequence primitive
				rs.Mutate(kernel.R4Stderr, func(v uint64) uint64 { return v + 1 })
			}
		}()
	}
	wg.Wait()

	expectedNext := uint64(initialSeq + workers*rounds)
	expectedStderr := uint64(workers * rounds)
	gotNext := rs.Read(kernel.R2NextAct)
	gotStderr := rs.Read(kernel.R4Stderr)

	t.Logf("Scenario 2 — $R0$–$R7$ mutation repulsion via kernel.RegisterSet")
	t.Logf("  workload      : %d workers × %d rounds of locked read-modify-write on R2/R4", workers, rounds)
	t.Logf("  $R2$ NextAct  : got %d, expected %d (lost: %d)", gotNext, expectedNext, expectedNext-gotNext)
	t.Logf("  $R4$ Stderr   : got %d, expected %d (lost: %d)", gotStderr, expectedStderr, expectedStderr-gotStderr)
	t.Logf("  snapshot      : %v", rs.ReadAll())
	t.Logf("  verdict       : all concurrent mutation repelled — final state exact")

	if gotNext != expectedNext {
		t.Fatalf("$R2$ NextAct diverged: got %d, want %d", gotNext, expectedNext)
	}
	if gotStderr != expectedStderr {
		t.Fatalf("$R4$ Stderr diverged: got %d, want %d", gotStderr, expectedStderr)
	}
}
