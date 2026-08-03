package models_comp

// GovernancePolicy is the single shared policy, expressed as write-time
// guards. It is applied identically by GuardedState (Model B′) and by the
// register-file adapter in experiment E1v2 (Model C), so the two models
// are compared under exactly the same rules. Denied writes are "repelled":
// rejected and counted, and the mutation does not take effect.
//
// The policy encodes the invariants of the core suite as write-time
// rejection rules: $R6$ Health is domain-checked (I3), $R2$ NextAct is
// monotonic (I1), $R1$ Watchdog may only be disarmed or extended (I4).
func GovernancePolicy(cell string, oldV, newV uint64) bool {
	switch cell {
	case CellHealth:
		return newV <= HealthDown
	case CellNextAct:
		return newV >= oldV
	case CellWatchdog:
		return newV == 0 || newV >= oldV
	}
	return true
}
