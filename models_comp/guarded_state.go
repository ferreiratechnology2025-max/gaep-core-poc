package models_comp

import (
	"sync"
	"time"
)

// GuardedState is Model B′: the strongest challenger to the register file.
// It is the same typed object as TypedState, but with unexported fields and
// a closed, policy-guarded API — the same GovernancePolicy the register
// model enforces, plus automatic timestamps. If B′ preserves every property
// the register file preserves, then the *register abstraction itself* is
// not what carries the weight; the closed, guarded surface is.
type GuardedState struct {
	mu        sync.RWMutex
	session   uint64
	watchdog  uint64
	nextAct   uint64
	isolation uint64
	stderr    uint64
	badge     uint64
	health    uint64
	timestamp uint64
	rejected  uint64
	writes    uint64
}

// NewGuardedState returns an empty guarded state.
func NewGuardedState() *GuardedState { return &GuardedState{} }

func (g *GuardedState) Model() string {
	return "B′ — GuardedState (closed struct, unexported fields)"
}

func (g *GuardedState) Get(key string) (uint64, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.getLocked(key)
}

func (g *GuardedState) getLocked(key string) (uint64, bool) {
	switch key {
	case CellSession:
		return g.session, true
	case CellWatchdog:
		return g.watchdog, true
	case CellNextAct:
		return g.nextAct, true
	case CellIsolation:
		return g.isolation, true
	case CellStderr:
		return g.stderr, true
	case CellBadge:
		return g.badge, true
	case CellHealth:
		return g.health, true
	case CellTimestamp:
		return g.timestamp, true
	}
	return 0, false
}

func (g *GuardedState) setLocked(key string, v uint64) {
	switch key {
	case CellSession:
		g.session = v
	case CellWatchdog:
		g.watchdog = v
	case CellNextAct:
		g.nextAct = v
	case CellIsolation:
		g.isolation = v
	case CellStderr:
		g.stderr = v
	case CellBadge:
		g.badge = v
	case CellHealth:
		g.health = v
	case CellTimestamp:
		g.timestamp = v
	}
}

// stampLocked records a strictly-increasing time; caller must hold mu.
func (g *GuardedState) stampLocked() {
	now := uint64(time.Now().UnixNano())
	if now <= g.timestamp {
		now = g.timestamp + 1
	}
	g.timestamp = now
}

func (g *GuardedState) Set(key string, v uint64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	old, _ := g.getLocked(key)
	if !GovernancePolicy(key, old, v) {
		g.rejected++
		return
	}
	g.setLocked(key, v)
	g.writes++
	if key != CellTimestamp {
		g.stampLocked()
	}
}

func (g *GuardedState) Apply(key string, fn func(uint64) uint64) uint64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	old, _ := g.getLocked(key)
	nv := fn(old)
	if !GovernancePolicy(key, old, nv) {
		g.rejected++
		return old
	}
	g.setLocked(key, nv)
	g.writes++
	if key != CellTimestamp {
		g.stampLocked()
	}
	return nv
}

func (g *GuardedState) Snapshot() Snapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return Snapshot{
		Session:   g.session,
		Watchdog:  g.watchdog,
		NextAct:   g.nextAct,
		Isolation: g.isolation,
		Stderr:    g.stderr,
		Badge:     g.badge,
		Health:    g.health,
		Timestamp: g.timestamp,
		Writes:    g.writes,
	}
}

func (g *GuardedState) Restore(s Snapshot) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.session, g.watchdog, g.nextAct, g.isolation = s.Session, s.Watchdog, s.NextAct, s.Isolation
	g.stderr, g.badge, g.health, g.timestamp = s.Stderr, s.Badge, s.Health, s.Timestamp
	g.writes = s.Writes
}

func (g *GuardedState) Rejected() uint64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.rejected
}

func (g *GuardedState) RawSurface() bool { return false }
