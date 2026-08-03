package models_comp

import "sync"

// TypedState is Model B: a plain typed object (SessionState) with exported
// fields, protected by a mutex only for the concurrency workload. There is
// no register concept and no policy: the exported fields are the raw
// mutation surface, so invalid values and regressions silently "stick",
// and nothing maintains the timestamp automatically.
type TypedState struct {
	mu sync.RWMutex
	SessionState
	writes uint64
}

// NewTypedState returns an empty typed state.
func NewTypedState() *TypedState { return &TypedState{} }

func (t *TypedState) Model() string { return "B — TypedState (plain struct, exported fields)" }

func (t *TypedState) Get(key string) (uint64, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.getLocked(key)
}

func (t *TypedState) getLocked(key string) (uint64, bool) {
	switch key {
	case CellSession:
		return t.SessionID, true
	case CellWatchdog:
		return uint64(t.WatchdogN), true
	case CellNextAct:
		return t.NextActSeq, true
	case CellIsolation:
		return uint64(t.Isolation), true
	case CellStderr:
		return t.StderrCount, true
	case CellBadge:
		return t.SKKBadge, true
	case CellHealth:
		return uint64(t.Health), true
	case CellTimestamp:
		return uint64(t.UpdatedAt), true
	}
	return 0, false
}

func (t *TypedState) setLocked(key string, v uint64) {
	switch key {
	case CellSession:
		t.SessionID = v
	case CellWatchdog:
		t.WatchdogN = int64(v)
	case CellNextAct:
		t.NextActSeq = v
	case CellIsolation:
		t.Isolation = uint32(v)
	case CellStderr:
		t.StderrCount = v
	case CellBadge:
		t.SKKBadge = v
	case CellHealth:
		t.Health = uint32(v)
	case CellTimestamp:
		t.UpdatedAt = int64(v)
	}
}

func (t *TypedState) Set(key string, v uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.setLocked(key, v)
	t.writes++
}

func (t *TypedState) Apply(key string, fn func(uint64) uint64) uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	old, _ := t.getLocked(key)
	nv := fn(old)
	t.setLocked(key, nv)
	t.writes++
	return nv
}

func (t *TypedState) Snapshot() Snapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return Snapshot{
		Session:   t.SessionID,
		Watchdog:  uint64(t.WatchdogN),
		NextAct:   t.NextActSeq,
		Isolation: uint64(t.Isolation),
		Stderr:    t.StderrCount,
		Badge:     t.SKKBadge,
		Health:    uint64(t.Health),
		Timestamp: uint64(t.UpdatedAt),
		Writes:    t.writes,
	}
}

func (t *TypedState) Restore(s Snapshot) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.SessionID, t.WatchdogN, t.NextActSeq = s.Session, int64(s.Watchdog), s.NextAct
	t.Isolation, t.StderrCount, t.SKKBadge = uint32(s.Isolation), s.Stderr, s.Badge
	t.Health, t.UpdatedAt = uint32(s.Health), int64(s.Timestamp)
	t.writes = s.Writes
}

func (t *TypedState) Rejected() uint64 { return 0 }

func (t *TypedState) RawSurface() bool { return true }
