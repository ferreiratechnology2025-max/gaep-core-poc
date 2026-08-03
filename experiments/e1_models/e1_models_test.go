// Package e1_models_test is Experiment E1v2 of the GAEP Core POC: the
// three-model (plus challenger) comparison under EQUAL synchronization.
//
// The hypothesis is falsifiable: "an implementation without explicit
// registers cannot preserve the invariants I0–I4 under the same
// conditions." To test it fairly, the concurrent workload, the governance
// ride, the rollback check and the adversarial mutation battery are run
// against four candidates that all share the same mutex discipline:
//
//	Model A  — GenericContext (map[string]any), unguarded, raw surface.
//	Model B  — TypedState (plain struct, exported fields), unguarded.
//	Model B′ — GuardedState (closed struct, unexported fields) with the
//	           same write-time policy the register file enforces.
//	Model C  — kernel.RegisterSet, the $R0$–$R7$ register file.
//
// Metrics are architectural, not performance: invariant preservation under
// contention (M1), invalid states reachable through the raw surface (M2),
// rollback fidelity (M3), auditability and raw mutation surface (M4).
//
// Evidence output is captured to adr-evidence/E1v2_evidence.log.
package e1_models_test

import (
	"sync"
	"testing"
	"time"

	"gaep-core-poc/kernel"
	"gaep-core-poc/models_comp"
)

const (
	workers     = 8
	rounds      = 2000
	initialNext = 1000
)

// cellToReg/regToCell map the shared cell keys onto the register file.
var cellToReg = map[string]kernel.RegisterID{
	models_comp.CellSession:   kernel.R0Session,
	models_comp.CellWatchdog:  kernel.R1Watchdog,
	models_comp.CellNextAct:   kernel.R2NextAct,
	models_comp.CellIsolation: kernel.R3Isolation,
	models_comp.CellStderr:    kernel.R4Stderr,
	models_comp.CellBadge:     kernel.R5SKKBadge,
	models_comp.CellHealth:    kernel.R6Health,
	models_comp.CellTimestamp: kernel.R7Timestamp,
}

var regToCell = func() map[kernel.RegisterID]string {
	m := make(map[kernel.RegisterID]string, len(cellToReg))
	for cell, id := range cellToReg {
		m[id] = cell
	}
	return m
}()

// registerModel adapts kernel.RegisterSet to the shared StateModel surface,
// applying the same GovernancePolicy that GuardedState enforces.
type registerModel struct {
	rs *kernel.RegisterSet
}

func newRegisterModel() *registerModel {
	rs := kernel.NewRegisterSetWithPolicy(kernel.RegisterPolicy{
		AllowWrite: func(id kernel.RegisterID, oldV, newV uint64) bool {
			cell, ok := regToCell[id]
			if !ok || cell == models_comp.CellTimestamp {
				return true // system cell: always permitted
			}
			return models_comp.GovernancePolicy(cell, oldV, newV)
		},
	})
	return &registerModel{rs: rs}
}

func (m *registerModel) Model() string { return "C — RegisterSet ($R0$–$R7$)" }

func (m *registerModel) Get(key string) (uint64, bool) {
	id, ok := cellToReg[key]
	if !ok {
		return 0, false
	}
	return m.rs.Read(id), true
}

func (m *registerModel) Set(key string, v uint64) {
	if id, ok := cellToReg[key]; ok {
		m.rs.Write(id, v)
	}
}

func (m *registerModel) Apply(key string, fn func(uint64) uint64) uint64 {
	id, ok := cellToReg[key]
	if !ok {
		return 0 // unknown key: nothing to mutate
	}
	return m.rs.Mutate(id, fn)
}

func (m *registerModel) Snapshot() models_comp.Snapshot {
	a, writes := m.rs.ReadAllWithWrites()
	return models_comp.Snapshot{
		Session:   a[0],
		Watchdog:  a[1],
		NextAct:   a[2],
		Isolation: a[3],
		Stderr:    a[4],
		Badge:     a[5],
		Health:    a[6],
		Timestamp: a[7],
		Writes:    writes,
	}
}

func (m *registerModel) Restore(s models_comp.Snapshot) {
	m.rs.Restore([kernel.RegisterCount]uint64{
		s.Session, s.Watchdog, s.NextAct, s.Isolation, s.Stderr, s.Badge, s.Health, s.Timestamp,
	}, s.Writes)
}

func (m *registerModel) Rejected() uint64 { return m.rs.Rejected() }

func (m *registerModel) RawSurface() bool { return false }

// --- shared workload, ride, rollback and battery (identical for all models) ---

// runConcurrentWorkload runs workers×rounds of locked read-modify-write on
// next_act and stderr, sampling snapshots along the way (M1).
func runConcurrentWorkload(m models_comp.StateModel) []models_comp.Snapshot {
	var wg sync.WaitGroup
	var trace []models_comp.Snapshot
	var traceMu sync.Mutex
	const sampleEvery = 128
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				m.Apply(models_comp.CellNextAct, func(v uint64) uint64 { return v + 1 })
				m.Apply(models_comp.CellStderr, func(v uint64) uint64 { return v + 1 })
				if r%sampleEvery == 0 {
					// Take the snapshot while holding the trace mutex so that
					// append order is chronological (the previous version
					// sampled before locking, letting an older snapshot be
					// appended after a newer one and faking regressions).
					traceMu.Lock()
					trace = append(trace, m.Snapshot())
					traceMu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	// Final sample after all writes: guarantees at least one adjacent pair
	// with an intervening write, so I2 (timestamp advances per write) is
	// deterministically exercised even for models that never stamp.
	traceMu.Lock()
	trace = append(trace, m.Snapshot())
	traceMu.Unlock()
	return trace
}

// runGovernanceRide drives the full governance trajectory (health ride,
// watchdog arm/extend) and samples snapshots (M1b).
func runGovernanceRide(m models_comp.StateModel) []models_comp.Snapshot {
	var trace []models_comp.Snapshot
	now := uint64(time.Now().UnixNano())
	m.Set(models_comp.CellHealth, 1)
	trace = append(trace, m.Snapshot())
	m.Set(models_comp.CellWatchdog, now+10e9)
	trace = append(trace, m.Snapshot())
	m.Set(models_comp.CellHealth, 2)
	trace = append(trace, m.Snapshot())
	m.Set(models_comp.CellHealth, 3)
	trace = append(trace, m.Snapshot())
	m.Apply(models_comp.CellNextAct, func(v uint64) uint64 { return v + 1 })
	trace = append(trace, m.Snapshot())
	return trace
}

// runRollbackCheck saves a snapshot, mutates, restores and verifies exact
// equality (M3).
func runRollbackCheck(m models_comp.StateModel) bool {
	saved := m.Snapshot()
	for r := 0; r < 50; r++ {
		m.Apply(models_comp.CellStderr, func(v uint64) uint64 { return v + 1 })
	}
	m.Restore(saved)
	return m.Snapshot() == saved
}

// runAdversarialBattery attacks each model where it is weakest: through the
// raw surface for A and B (live map, exported fields), through the
// policy-guarded API for B′ and C (M2). Returns attempted, stuck (invalid
// mutations that actually corrupted state) and rejected (repelled).
func runAdversarialBattery(m models_comp.StateModel) (attempted, stuck, rejected int) {
	const nAttempts = 3
	rejectedBefore := int(m.Rejected())
	pre := m.Snapshot()
	switch model := m.(type) {
	case *models_comp.GenericContext:
		// raw surface: the live map — wrong types and regressions stick.
		model.Data()[models_comp.CellHealth] = "DOWN"
		model.Data()[models_comp.CellNextAct] = uint64(0)
		model.Data()[models_comp.CellWatchdog] = uint64(1)
	case *models_comp.TypedState:
		// raw surface: exported fields — invalid values and regressions stick.
		model.SessionState.Health = 99
		model.SessionState.NextActSeq = 0
		model.SessionState.WatchdogN = 1
	default:
		// closed surface: the same mutations through the guarded API.
		m.Set(models_comp.CellHealth, 99)
		m.Set(models_comp.CellNextAct, 0)
		m.Set(models_comp.CellWatchdog, 1)
		rejected = int(m.Rejected())
	}
	post := m.Snapshot()
	return nAttempts, countViolations(pre, post), int(m.Rejected()) - rejectedBefore
}

// batteryInvariants are the invariants an invalid mutation can trip. I2
// ($R7$ Timestamp advances) is deliberately excluded: a rejected (or absent)
// mutation legitimately leaves the timestamp untouched, so comparing two
// identical snapshots must not count as corruption.
var batteryInvariants = []models_comp.Invariant{
	models_comp.InvTypeIntegrity,
	models_comp.InvNextActMonotonic,
	models_comp.InvHealthDomain,
	models_comp.InvWatchdogFuture,
}

// countViolations returns how many battery invariants are violated by cur
// given the previous sample pre.
func countViolations(pre, cur models_comp.Snapshot) int {
	n := 0
	for _, inv := range batteryInvariants {
		if ok, _ := inv.Check(pre, cur); !ok {
			n++
		}
	}
	return n
}

func checkInvariants(t *testing.T, label string, trace []models_comp.Snapshot) []models_comp.InvariantResult {
	var results []models_comp.InvariantResult
	if len(trace) == 0 {
		t.Logf("  %s: (no samples)", label)
		return results
	}
	allOK := true
	for _, inv := range models_comp.CoreInvariants {
		res := models_comp.CheckTrace(inv, trace)
		status := "OK"
		if !res.OK {
			status = "VIOLATED: " + res.Detail
			allOK = false
		}
		t.Logf("  %-4s %-46s %s", res.ID, res.Name, status)
		results = append(results, res)
	}
	t.Logf("  → %s: %v", label, allOK)
	return results
}

func allOK(results []models_comp.InvariantResult) bool {
	for _, r := range results {
		if !r.OK {
			return false
		}
	}
	return true
}

func supportsAutoTimestamp(m models_comp.StateModel) bool {
	switch m.(type) {
	case *models_comp.GuardedState, *registerModel:
		return true
	}
	return false
}

// TestE1v2_ThreeModelComparison runs every model through the same battery.
func TestE1v2_ThreeModelComparison(t *testing.T) {
	models := []models_comp.StateModel{
		models_comp.NewGenericContext(),
		models_comp.NewTypedState(),
		models_comp.NewGuardedState(),
		newRegisterModel(),
	}
	for _, m := range models {
		m.Set(models_comp.CellSession, 42)
		m.Set(models_comp.CellBadge, 0x5eed)
		m.Set(models_comp.CellIsolation, 4)
		m.Set(models_comp.CellNextAct, initialNext)
	}

	t.Logf("E1v2 — Model A (map) / Model B (plain struct) / Model B′ (guarded struct) / Model C (registers)")
	t.Logf("All four run under EQUAL synchronization (mutex). Differences: raw mutation surface")
	t.Logf("and write-time policy. Metrics are architectural, not performance.")
	t.Logf("")

	for _, m := range models {
		t.Run(m.Model(), func(t *testing.T) {
			t.Logf("model: %s", m.Model())
			t.Logf("  raw surface: %v | automatic timestamp: %v", m.RawSurface(), supportsAutoTimestamp(m))

			// M1 + M1b: invariant preservation under contention and ride.
			trace := runConcurrentWorkload(m)
			res1 := checkInvariants(t, "M1 concurrent 8×2000 RMW (next_act/stderr)", trace)
			ride := runGovernanceRide(m)
			res2 := checkInvariants(t, "M1b governance ride (health/watchdog)", ride)

			// M3: rollback fidelity.
			ok := runRollbackCheck(m)
			t.Logf("  M3 rollback: save → mutate → restore → exact: %v", ok)
			if !ok {
				t.Errorf("M3 rollback fidelity failed for %s", m.Model())
			}

			// M2: adversarial invalid-mutation battery.
			attempted, stuck, rejected := runAdversarialBattery(m)
			t.Logf("  M2 invalid mutations: attempted=%d stuck=%d rejected=%d", attempted, stuck, rejected)
			if stuck < attempted && m.RawSurface() {
				t.Logf("    note: %d corruption(s) invisible to the invariant suite — models without an", attempted-stuck)
				t.Logf("    automatic timestamp let I4 under-detect invalid watchdog values.")
			}

			switch m.(type) {
			case *models_comp.GenericContext, *models_comp.TypedState:
				// Unguarded models: I2 (automatic timestamp) must fail, and
				// invalid mutations must stick.
				if allOK(res1) {
					t.Errorf("%s: expected at least one invariant violation in M1 (no automatic timestamp)", m.Model())
				}
				if stuck == 0 {
					t.Errorf("%s: expected invalid mutations to stick through the raw surface", m.Model())
				}
			default:
				// Guarded models: everything preserved, everything repelled.
				if !allOK(res1) || !allOK(res2) {
					t.Errorf("%s: expected all invariants to hold in M1 and M1b", m.Model())
				}
				if stuck != 0 || rejected != 3 {
					t.Errorf("%s: expected full repulsion (stuck=%d rejected=%d, want 0/3)", m.Model(), stuck, rejected)
				}
			}
		})
	}

	t.Logf("")
	t.Logf("VERDICT (refutation-first):")
	t.Logf(" 1. Concurrency safety (M1) is a property of LOCKING: with equal mutexes, every")
	t.Logf("    model preserves monotonic counters. The register file adds nothing here.")
	t.Logf(" 2. Invariant enforcement (M2) is a property of the CLOSED, POLICY-GUARDED SURFACE:")
	t.Logf("    A (map) and B (plain struct) admit invalid states (wrong types, out-of-domain")
	t.Logf("    values, regressions); B′ (guarded struct) and C (registers) repel them all.")
	t.Logf(" 3. B′ matches C on every measured property (M1, M1b, M2, M3, auditability). The")
	t.Logf("    register abstraction is NOT proven necessary by E1v2 — the load-bearing")
	t.Logf("    property is the guarded surface. Necessity is tested component-by-component in")
	t.Logf("    E0; the remaining case for $R0$–$R7$ rests on notation, audit-by-construction")
	t.Logf("    and the hardware metaphor, not on these metrics.")
}
