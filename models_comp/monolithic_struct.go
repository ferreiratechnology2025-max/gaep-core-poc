// Package models_comp holds the flat, unsynchronized control structures
// used as the baseline in GAEP Core experiments. They mirror the kernel
// register file ($R0$–$R7$) one-to-one but carry no locking, which makes
// them cheap to touch directly and exactly as vulnerable to concurrent
// mutation as the Scenario 1 evidence in experiments/e1_registers shows.
package models_comp

// SessionState is the monolithic control struct: eight fields mirroring
// the governance registers of kernel.RegisterSet, with no synchronization.
type SessionState struct {
	SessionID   uint64 // $R0$ Session: active session identifier
	WatchdogN   int64  // $R1$ Watchdog: expiry as unix nanoseconds (0 = disarmed)
	NextActSeq  uint64 // $R2$ NextAct: sequence number of the next governance action
	Isolation   uint32 // $R3$ Isolation: isolation-level bitmask
	StderrCount uint64 // $R4$ Stderr: accumulated stderr event counter
	SKKBadge    uint64 // $R5$ SKK Badge: signed badge token bound to the session
	Health      uint32 // $R6$ Health: engine health status code
	UpdatedAt   int64  // $R7$ Timestamp: unix nanoseconds of last mutation
}
