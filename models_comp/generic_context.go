package models_comp

import "sync"

// GenericContext is Model A: the "generic context" baseline — an any-typed
// map, protected by a mutex only for the concurrency workload. The live
// map is exposed through Data: callers can write any key with any value,
// so typos, wrong types and invalid values all silently "stick". There is
// no policy and no automatic timestamp.
type GenericContext struct {
	mu     sync.RWMutex
	data   map[string]any
	writes uint64
}

// NewGenericContext returns an empty generic context.
func NewGenericContext() *GenericContext {
	return &GenericContext{data: map[string]any{}}
}

func (g *GenericContext) Model() string { return "A — GenericContext (map[string]any)" }

// Data exposes the live map: the raw mutation surface.
func (g *GenericContext) Data() map[string]any { return g.data }

func (g *GenericContext) Get(key string) (uint64, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	v, ok := g.data[key].(uint64)
	return v, ok
}

func (g *GenericContext) Set(key string, v uint64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.data[key] = v
	g.writes++
}

func (g *GenericContext) Apply(key string, fn func(uint64) uint64) uint64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	v, _ := g.data[key].(uint64)
	nv := fn(v)
	g.data[key] = nv
	g.writes++
	return nv
}

func (g *GenericContext) Snapshot() Snapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()
	cell := func(key string) uint64 {
		v, ok := g.data[key]
		if !ok {
			return 0 // absent cell is neutral, not corrupt
		}
		if u, isUint := v.(uint64); isUint {
			return u
		}
		return CorruptSentinel // present but wrong type: corruption
	}
	return Snapshot{
		Session:   cell(CellSession),
		Watchdog:  cell(CellWatchdog),
		NextAct:   cell(CellNextAct),
		Isolation: cell(CellIsolation),
		Stderr:    cell(CellStderr),
		Badge:     cell(CellBadge),
		Health:    cell(CellHealth),
		Timestamp: cell(CellTimestamp),
		Writes:    g.writes,
	}
}

func (g *GenericContext) Restore(s Snapshot) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.data = map[string]any{
		CellSession:   s.Session,
		CellWatchdog:  s.Watchdog,
		CellNextAct:   s.NextAct,
		CellIsolation: s.Isolation,
		CellStderr:    s.Stderr,
		CellBadge:     s.Badge,
		CellHealth:    s.Health,
		CellTimestamp: s.Timestamp,
	}
	g.writes = s.Writes
}

func (g *GenericContext) Rejected() uint64 { return 0 }

func (g *GenericContext) RawSurface() bool { return true }
