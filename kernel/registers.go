// Package kernel defines the GAEP Core governance register file: eight
// hardware-style registers ($R0$–$R7$) that carry the session, watchdog,
// next-action, isolation, stderr, badge, health and timestamp state of the
// validation engine.
//
// RegisterSet is the concurrency-safe alternative to the flat control
// structs in package models_comp. Every read and write is serialized, and
// the Mutate primitive performs an atomic read-modify-write under a single
// lock — concurrent mutations are "repelled" (never lost, never observed
// torn). An optional RegisterPolicy can additionally reject invalid
// mutations (wrong value domains, regressions), counting each rejection.
package kernel

import (
	"strconv"
	"sync"
	"time"
)

// RegisterID enumerates the eight governance registers.
type RegisterID int

const (
	// R0Session is the active session identifier ($R0$).
	R0Session RegisterID = iota
	// R1Watchdog is the watchdog expiry as unix nanoseconds; 0 = disarmed.
	R1Watchdog
	// R2NextAct is the sequence number of the next governance action to dispatch.
	R2NextAct
	// R3Isolation is the isolation-level bitmask enforced by the engine.
	R3Isolation
	// R4Stderr is the accumulated stderr event counter.
	R4Stderr
	// R5SKKBadge is the signed SKK badge token bound to the session.
	R5SKKBadge
	// R6Health is the current engine health status code.
	R6Health
	// R7Timestamp is the unix-nanosecond timestamp of the last mutation.
	R7Timestamp

	registerCount
)

// RegisterCount is the fixed number of registers in the file ($R0$–$R7$).
const RegisterCount = int(registerCount)

// registerNames maps each register id to its symbolic name.
var registerNames = [...]string{
	R0Session:   "$R0$ Session",
	R1Watchdog:  "$R1$ Watchdog",
	R2NextAct:   "$R2$ NextAct",
	R3Isolation: "$R3$ Isolation",
	R4Stderr:    "$R4$ Stderr",
	R5SKKBadge:  "$R5$ SKK Badge",
	R6Health:    "$R6$ Health",
	R7Timestamp: "$R7$ Timestamp",
}

// String returns the symbolic name of the register (e.g. "$R2$ NextAct").
func (id RegisterID) String() string {
	if id >= 0 && int(id) < len(registerNames) {
		return registerNames[id]
	}
	return "invalid register"
}

// RegisterPolicy is an optional write-time guard consulted by Write and
// Mutate. Denied mutations are repelled and counted (see Rejected).
type RegisterPolicy struct {
	// AllowWrite reports whether newV may replace oldV in register id.
	// A nil function (or nil policy) permits every write.
	AllowWrite func(id RegisterID, oldV, newV uint64) bool
}

// RegisterSet is a fixed-width, thread-safe register file holding
// $R0$–$R7$. All eight cells are plain 64-bit words guarded by a single
// RWMutex. Writers use Write (or the Mutate/Next primitives); readers use
// Read or take a consistent snapshot with ReadAll.
type RegisterSet struct {
	mu       sync.RWMutex
	reg      [RegisterCount]uint64
	policy   *RegisterPolicy
	rejected uint64
	writes   uint64 // write-generation counter (every successful mutation)
}

// NewRegisterSet returns an unguarded register file with $R7$ Timestamp
// seeded to the current time (unix nanoseconds).
func NewRegisterSet() *RegisterSet {
	return NewRegisterSetWithPolicy(RegisterPolicy{})
}

// NewRegisterSetWithPolicy returns a register file enforcing p on every
// Write/Mutate/Next; denied mutations are repelled and counted.
func NewRegisterSetWithPolicy(p RegisterPolicy) *RegisterSet {
	rs := &RegisterSet{}
	if p.AllowWrite != nil {
		rs.policy = &p
	}
	rs.Write(R7Timestamp, uint64(time.Now().UnixNano()))
	return rs
}

// valid reports whether id names one of the eight registers.
func (id RegisterID) valid() bool { return id >= 0 && int(id) < int(registerCount) }

// allowedLocked reports whether oldV→newV is permitted for id and counts
// rejections; caller must hold mu.
func (rs *RegisterSet) allowedLocked(id RegisterID, oldV, newV uint64) bool {
	if rs.policy == nil || rs.policy.AllowWrite == nil {
		return true
	}
	if !rs.policy.AllowWrite(id, oldV, newV) {
		rs.rejected++
		return false
	}
	return true
}

// stampLocked records a strictly-increasing time in $R7$ Timestamp;
// caller must hold mu.
func (rs *RegisterSet) stampLocked() {
	now := uint64(time.Now().UnixNano())
	if now <= rs.reg[R7Timestamp] {
		now = rs.reg[R7Timestamp] + 1
	}
	rs.reg[R7Timestamp] = now
}

// Read returns the current value of register id.
func (rs *RegisterSet) Read(id RegisterID) uint64 {
	if !id.valid() {
		panic("kernel: Read: invalid register id " + strconv.Itoa(int(id)))
	}
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.reg[id]
}

// Write stores v into register id, stamping $R7$ Timestamp with a
// strictly-increasing time unless id is $R7$ itself. Writes denied by the
// policy are repelled and counted, and do not advance the timestamp.
func (rs *RegisterSet) Write(id RegisterID, v uint64) {
	if !id.valid() {
		panic("kernel: Write: invalid register id " + strconv.Itoa(int(id)))
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if !rs.allowedLocked(id, rs.reg[id], v) {
		return
	}
	rs.reg[id] = v
	rs.writes++
	if id != R7Timestamp {
		rs.stampLocked()
	}
}

// ReadAll returns a consistent snapshot of all eight registers under a
// single read lock.
func (rs *RegisterSet) ReadAll() [RegisterCount]uint64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.reg
}

// ReadAllWithWrites returns a consistent snapshot of all eight registers
// plus the write-generation counter under a single read lock.
func (rs *RegisterSet) ReadAllWithWrites() ([RegisterCount]uint64, uint64) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.reg, rs.writes
}

// Mutate applies fn to the current value of register id and stores the
// result, all under a single write lock. This is the mutation-repulsion
// primitive: concurrent read-modify-write sequences are serialized, so no
// update is lost and no torn value is ever observed. Mutations denied by
// the policy are repelled and counted, and the previous value is returned.
func (rs *RegisterSet) Mutate(id RegisterID, fn func(uint64) uint64) uint64 {
	if !id.valid() {
		panic("kernel: Mutate: invalid register id " + strconv.Itoa(int(id)))
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	newV := fn(rs.reg[id])
	if !rs.allowedLocked(id, rs.reg[id], newV) {
		return rs.reg[id]
	}
	rs.reg[id] = newV
	rs.writes++
	if id != R7Timestamp {
		rs.stampLocked()
	}
	return newV
}

// Next atomically increments register id and returns the new value. It is
// the dedicated primitive for advancing sequence registers such as
// $R2$ NextAct.
func (rs *RegisterSet) Next(id RegisterID) uint64 {
	return rs.Mutate(id, func(v uint64) uint64 { return v + 1 })
}

// Restore overwrites the whole register file with vals and resets the
// write-generation counter (a trusted system operation that bypasses the
// policy, e.g. exact rollback). Unlike a mutation, a restore does not
// stamp $R7$: the file returns to the exact saved state, so a post-restore
// snapshot equals the saved snapshot.
func (rs *RegisterSet) Restore(vals [RegisterCount]uint64, writes uint64) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.reg = vals
	rs.writes = writes
}

// Rejected returns how many mutations the policy has repelled so far.
func (rs *RegisterSet) Rejected() uint64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.rejected
}
