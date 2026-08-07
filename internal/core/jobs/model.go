// Package jobs provides private, bounded execution of classified core work.
// It deliberately knows neither parser nor Store types.
package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidJob            = errors.New("jobs: invalid job")
	ErrInvalidClassification = errors.New("jobs: invalid classification")
	ErrInvalidExecutor       = errors.New("jobs: invalid executor")
)

type Kind string

const (
	KindFullRebuild      Kind = "full_rebuild"
	KindIndexBatch       Kind = "index_batch"
	KindReconcileScan    Kind = "reconcile_scan"
	KindWALCheckpoint    Kind = "wal_checkpoint"
	KindApplyTransaction Kind = "apply_transaction"
)

func (k Kind) valid() bool {
	switch k {
	case KindFullRebuild, KindIndexBatch, KindReconcileScan, KindWALCheckpoint, KindApplyTransaction:
		return true
	default:
		return false
	}
}

type Priority string

const (
	PriorityInteractive Priority = "interactive"
	PriorityIncremental Priority = "incremental"
	PriorityMaintenance Priority = "maintenance"
)

func (p Priority) valid() bool {
	switch p {
	case PriorityInteractive, PriorityIncremental, PriorityMaintenance:
		return true
	default:
		return false
	}
}

type IdentityKey struct{ raw string }

func (k IdentityKey) String() string { return k.raw }

type JobSpec struct {
	Kind        Kind
	IdentityKey string
	Checkpoint  string
	Priority    Priority
}

type Job struct {
	kind        Kind
	identityKey IdentityKey
	checkpoint  string
	priority    Priority
}

func NewJob(spec JobSpec) (Job, error) {
	if !spec.Kind.valid() {
		return Job{}, fmt.Errorf("jobs: kind %q: %w", spec.Kind, ErrInvalidJob)
	}
	if strings.TrimSpace(spec.IdentityKey) == "" {
		return Job{}, fmt.Errorf("jobs: empty identity key: %w", ErrInvalidJob)
	}
	if !spec.Priority.valid() {
		return Job{}, fmt.Errorf("jobs: priority %q: %w", spec.Priority, ErrInvalidJob)
	}
	return Job{kind: spec.Kind, identityKey: IdentityKey{raw: spec.IdentityKey}, checkpoint: spec.Checkpoint, priority: spec.Priority}, nil
}

func (j Job) Kind() Kind               { return j.kind }
func (j Job) IdentityKey() IdentityKey { return j.identityKey }
func (j Job) Checkpoint() string       { return j.checkpoint }
func (j Job) Priority() Priority       { return j.priority }

type Classification string

const (
	ClassificationParse Classification = "parse"
	ClassificationStore Classification = "store"
)

func (c Classification) valid() bool {
	switch c {
	case ClassificationParse, ClassificationStore:
		return true
	default:
		return false
	}
}

type Classifier interface {
	Classify(context.Context, Job) (Classification, error)
}

type Commit struct {
	job            Job
	classification Classification
}

func (c Commit) Job() Job                       { return c.job }
func (c Commit) Classification() Classification { return c.classification }

type Writer interface {
	Commit(context.Context, Commit) error
}

type State string

const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateSucceeded State = "succeeded"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
	StateCoalesced State = "coalesced"
)

type Severity string

const (
	SeverityInfo  Severity = "info"
	SeverityError Severity = "error"
)

type Diagnostic struct {
	Code     string
	Severity Severity
}

const (
	DiagnosticSucceeded = "LITH-J-0001"
	DiagnosticCancelled = "LITH-J-0002"
	DiagnosticFailed    = "LITH-J-0003"
	DiagnosticCoalesced = "LITH-J-0004"
)

type Result struct {
	Job        Job
	State      State
	Attempt    int
	Diagnostic Diagnostic
	Err        error
}

type FailureClass string

const (
	FailureDeterministic    FailureClass = "deterministic"
	FailureNonDeterministic FailureClass = "non_deterministic"
)

type Failure struct {
	class FailureClass
	err   error
}

func (f *Failure) Error() string              { return f.err.Error() }
func (f *Failure) Unwrap() error              { return f.err }
func NewDeterministicFailure(err error) error { return newFailure(FailureDeterministic, err) }
func NewTransientFailure(err error) error     { return newFailure(FailureNonDeterministic, err) }

func newFailure(class FailureClass, err error) error {
	if err == nil {
		return nil
	}
	return &Failure{class: class, err: err}
}
