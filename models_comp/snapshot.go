package models_comp

import (
	"fmt"
	"math"
)

// Snapshot is a neutral, copyable projection of governance state that all
// candidate state models (generic map, typed struct, register file) can be
// reduced to. Invariants are expressed over this projection only, so the
// same checker runs against every model. The fields mirror $R0$–$R7$ in
// order.
type Snapshot struct {
	Session   uint64
	Watchdog  uint64
	NextAct   uint64
	Isolation uint64
	Stderr    uint64
	Badge     uint64
	Health    uint64
	Timestamp uint64
	// Writes is a write-generation counter: it advances on every successful
	// mutation. It lets I2 check that the timestamp advances *per write*
	// rather than per sample (adjacent samples may have no write between
	// them, in which case an unchanged timestamp is legitimate).
	Writes uint64
}

// CorruptSentinel marks a cell whose value could not be projected as a
// uint64 (e.g. a wrong-typed value in the generic map model).
const CorruptSentinel = math.MaxUint64

// Health status codes (domain of $R6$ Health).
const (
	HealthUnknown  = 0
	HealthOK       = 1
	HealthDegraded = 2
	HealthDown     = 3
)

// Invariant is a single check over a trace of snapshots.
type Invariant struct {
	ID   string
	Name string
	// Check reports whether cur satisfies the invariant given the previous
	// sample. prev is the zero Snapshot for the first sample of a trace.
	Check func(prev, cur Snapshot) (bool, string)
}

// InvariantResult is the outcome of checking one invariant over a trace.
type InvariantResult struct {
	ID     string
	Name   string
	OK     bool
	Detail string
}

// Core invariants. I0–I4 are deliberately few: E0 measures which registers
// they actually pin (R1/R2/R6/R7) and which they leave unconstrained
// (R0/R3/R4/R5), surfacing requirement gaps instead of assuming coverage.
var (
	InvTypeIntegrity = Invariant{
		ID:   "I0",
		Name: "cell types are uint64",
		Check: func(_, cur Snapshot) (bool, string) {
			if cur.Session == CorruptSentinel || cur.Watchdog == CorruptSentinel ||
				cur.NextAct == CorruptSentinel || cur.Isolation == CorruptSentinel ||
				cur.Stderr == CorruptSentinel || cur.Badge == CorruptSentinel ||
				cur.Health == CorruptSentinel || cur.Timestamp == CorruptSentinel {
				return false, "a cell could not be projected as uint64 (wrong type)"
			}
			return true, ""
		},
	}

	InvNextActMonotonic = Invariant{
		ID:   "I1",
		Name: "$R2$ NextAct is monotonic",
		Check: func(prev, cur Snapshot) (bool, string) {
			if prev == (Snapshot{}) {
				return true, ""
			}
			if cur.NextAct < prev.NextAct {
				return false, fmt.Sprintf("NextAct regressed %d → %d", prev.NextAct, cur.NextAct)
			}
			return true, ""
		},
	}

	InvTimestampAdvances = Invariant{
		ID:   "I2",
		Name: "$R7$ Timestamp advances on every write",
		Check: func(prev, cur Snapshot) (bool, string) {
			if prev == (Snapshot{}) {
				return true, ""
			}
			if cur.Writes == prev.Writes {
				return true, "" // no write between samples: unchanged is fine
			}
			if cur.Writes < prev.Writes {
				return true, "" // rollback: generation and timestamp legitimately regress
			}
			if cur.Timestamp <= prev.Timestamp {
				return false, fmt.Sprintf("Timestamp did not advance after a write (%d → %d)", prev.Timestamp, cur.Timestamp)
			}
			return true, ""
		},
	}

	InvHealthDomain = Invariant{
		ID:   "I3",
		Name: "$R6$ Health is a valid status",
		Check: func(_, cur Snapshot) (bool, string) {
			if cur.Health == CorruptSentinel || cur.Health > HealthDown {
				return false, fmt.Sprintf("invalid health code %d", cur.Health)
			}
			return true, ""
		},
	}

	InvWatchdogFuture = Invariant{
		ID:   "I4",
		Name: "$R1$ Watchdog is disarmed or in the future",
		Check: func(_, cur Snapshot) (bool, string) {
			if cur.Watchdog == 0 {
				return true, ""
			}
			if cur.Watchdog <= cur.Timestamp {
				return false, fmt.Sprintf("watchdog %d expired at %d", cur.Watchdog, cur.Timestamp)
			}
			return true, ""
		},
	}

	// CoreInvariants is the suite shared by E0 and E1v2.
	CoreInvariants = []Invariant{
		InvTypeIntegrity,
		InvNextActMonotonic,
		InvTimestampAdvances,
		InvHealthDomain,
		InvWatchdogFuture,
	}
)

// CheckTrace evaluates inv over a trace of snapshots, returning the first
// violation (or OK if the invariant holds for every adjacent pair).
func CheckTrace(inv Invariant, samples []Snapshot) InvariantResult {
	var prev Snapshot
	for i, cur := range samples {
		ok, detail := inv.Check(prev, cur)
		if !ok {
			return InvariantResult{
				ID:     inv.ID,
				Name:   inv.Name,
				OK:     false,
				Detail: fmt.Sprintf("sample %d: %s", i, detail),
			}
		}
		prev = cur
	}
	return InvariantResult{ID: inv.ID, Name: inv.Name, OK: true}
}
