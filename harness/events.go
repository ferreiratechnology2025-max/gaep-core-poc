// Package harness is the core of the GAEP Core Architecture Validation
// Harness: a solution-free observable event vocabulary, a Property Oracle
// (P1–P8) with pre-defined operational criteria, and a scenario battery.
//
// Architectures never grade themselves: they only emit observable events;
// the Oracle decides PASS / FAIL / UNKNOWN. The vocabulary is deliberately
// free of any solution language (no registers, kernel, plan, outcome) and is
// an **observable contract**: frozen in
// adr-evidence/observable-contract-specification.md, it is the coupling
// point between rival architectures and survives mechanism replacement.
// The harness is a contract verifier, not an architecture comparer: rivals
// A–E may differ internally as long as they honor this vocabulary.
package harness

// EventKind is one observable, solution-free fact an architecture may emit.
type EventKind int

const (
	EvAdmissionRequested EventKind = iota
	EvAdmissionAccepted
	EvAdmissionRejected
	EvAuthorizationGranted
	EvAuthorizationDenied
	EvCapabilityStarted
	EvCapabilitySucceeded
	EvCapabilityFailed
	EvRollbackExecuted
	EvExternalEffect
	EvAuditWritten
	EvTimedOut
)

var eventNames = [...]string{
	"AdmissionRequested",
	"AdmissionAccepted",
	"AdmissionRejected",
	"AuthorizationGranted",
	"AuthorizationDenied",
	"CapabilityStarted",
	"CapabilitySucceeded",
	"CapabilityFailed",
	"RollbackExecuted",
	"ExternalEffect",
	"AuditWritten",
	"TimedOut",
}

// String returns the symbolic name of the event kind.
func (k EventKind) String() string {
	if k >= 0 && int(k) < len(eventNames) {
		return eventNames[k]
	}
	return "UnknownEvent"
}

// Event is one observable fact. Cap is the capability the event refers to
// ("" for global events such as admission or audit).
type Event struct {
	Seq  int
	Kind EventKind
	Cap  string
	Note string
}
