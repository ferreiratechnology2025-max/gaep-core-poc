package models_comp

// StateModel is the shared surface every candidate model implements, so
// that E1v2 can run the exact same workloads, batteries and checks against
// each of them. The three (plus a guarded challenger) are:
//
//	Model A  — GenericContext: an any-typed map, unguarded.
//	Model B  — TypedState: a plain struct with exported fields, unguarded.
//	Model B′ — GuardedState: a closed struct with unexported fields and the
//	           same write-time policy the register file enforces.
//	Model C  — kernel.RegisterSet, the $R0$–$R7$ register file (adapted).
type StateModel interface {
	// Model returns the human-readable model name.
	Model() string
	// Get returns the value of the named cell and whether it exists.
	Get(key string) (uint64, bool)
	// Set stores v in the named cell. May be rejected by policy (Rejected).
	Set(key string, v uint64)
	// Apply performs an atomic read-modify-write on the named cell and
	// returns the value stored (the previous value if policy denied it).
	Apply(key string, fn func(uint64) uint64) uint64
	// Snapshot projects the model onto the neutral Snapshot type.
	Snapshot() Snapshot
	// Restore overwrites model state from a snapshot (trusted rollback).
	Restore(s Snapshot)
	// Rejected returns how many mutations the model's policy has repelled.
	Rejected() uint64
	// RawSurface reports whether callers can mutate state directly,
	// bypassing Set/Apply (e.g. exported fields or the live map).
	RawSurface() bool
}

// Cell keys, mirroring $R0$–$R7$.
const (
	CellSession   = "session"
	CellWatchdog  = "watchdog"
	CellNextAct   = "next_act"
	CellIsolation = "isolation"
	CellStderr    = "stderr"
	CellBadge     = "badge"
	CellHealth    = "health"
	CellTimestamp = "timestamp"
)
