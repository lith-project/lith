package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

const defaultMaxAttempts = 3

type RetryWaiter func(context.Context, int) error

type Config struct {
	Workers     int
	MaxAttempts int
	RetryWaiter RetryWaiter
}

type Executor struct {
	classifier Classifier
	writer     Writer
	workers    int
	maxAttempt int
	wait       RetryWaiter
	commitGate chan struct{}
}

func NewExecutor(classifier Classifier, writer Writer, config Config) (*Executor, error) {
	if classifier == nil || writer == nil {
		return nil, fmt.Errorf("jobs: classifier and writer are required: %w", ErrInvalidExecutor)
	}
	if config.Workers < 0 || config.MaxAttempts < 0 {
		return nil, fmt.Errorf("jobs: worker and attempt counts must not be negative: %w", ErrInvalidExecutor)
	}
	if config.Workers == 0 {
		config.Workers = 1
	}
	if config.MaxAttempts == 0 {
		config.MaxAttempts = defaultMaxAttempts
	}
	gate := make(chan struct{}, 1)
	gate <- struct{}{}
	return &Executor{classifier: classifier, writer: writer, workers: config.Workers, maxAttempt: config.MaxAttempts, wait: config.RetryWaiter, commitGate: gate}, nil
}

func (e *Executor) Execute(ctx context.Context, jobs []Job) []Result {
	results := make([]Result, len(jobs))
	if len(jobs) == 0 {
		return results
	}
	leaders := make([]int, 0, len(jobs))
	seen := make(map[IdentityKey]struct{}, len(jobs))
	for index, job := range jobs {
		if _, ok := seen[job.identityKey]; ok {
			results[index] = Result{Job: job, State: StateCoalesced, Diagnostic: Diagnostic{Code: DiagnosticCoalesced, Severity: SeverityInfo}}
			continue
		}
		seen[job.identityKey] = struct{}{}
		leaders = append(leaders, index)
	}
	tasks := make(chan int)
	workerCount := e.workers
	if workerCount > len(leaders) {
		workerCount = len(leaders)
	}
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range tasks {
				results[index] = e.executeOne(ctx, jobs[index])
			}
		}()
	}
	for position, index := range leaders {
		select {
		case tasks <- index:
		case <-ctx.Done():
			for _, remaining := range leaders[position:] {
				results[remaining] = cancelledResult(jobs[remaining], ctx.Err())
			}
			position = len(leaders)
		}
		if position == len(leaders) {
			break
		}
	}
	close(tasks)
	workers.Wait()
	return results
}

func (e *Executor) executeOne(ctx context.Context, job Job) Result {
	for attempt := 1; attempt <= e.maxAttempt; attempt++ {
		if err := ctx.Err(); err != nil {
			return cancelledResult(job, err)
		}
		classification, err := e.classifier.Classify(ctx, job)
		if err == nil && !classification.valid() {
			err = fmt.Errorf("jobs: %q: %w", classification, ErrInvalidClassification)
		}
		if err == nil && classification == ClassificationParse {
			return successResult(job, attempt)
		}
		if err == nil {
			err = e.commit(ctx, Commit{job: job, classification: classification})
		}
		if err == nil {
			return successResult(job, attempt)
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return cancelledResult(job, ctxErr)
		}
		if classifyFailure(err) != FailureNonDeterministic || attempt == e.maxAttempt {
			return failedResult(job, attempt, err)
		}
		if e.wait != nil {
			if waitErr := e.wait(ctx, attempt); waitErr != nil {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return cancelledResult(job, ctxErr)
				}
				return failedResult(job, attempt, waitErr)
			}
		}
	}
	return failedResult(job, e.maxAttempt, ErrInvalidExecutor)
}

func (e *Executor) commit(ctx context.Context, commit Commit) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.commitGate:
	}
	defer func() { e.commitGate <- struct{}{} }()
	return e.writer.Commit(ctx, commit)
}

func classifyFailure(err error) FailureClass {
	var failure *Failure
	if errors.As(err, &failure) {
		return failure.class
	}
	return FailureDeterministic
}

func successResult(job Job, attempt int) Result {
	return Result{Job: job, State: StateSucceeded, Attempt: attempt, Diagnostic: Diagnostic{Code: DiagnosticSucceeded, Severity: SeverityInfo}}
}

func cancelledResult(job Job, err error) Result {
	return Result{Job: job, State: StateCancelled, Diagnostic: Diagnostic{Code: DiagnosticCancelled, Severity: SeverityInfo}, Err: err}
}

func failedResult(job Job, attempt int, err error) Result {
	return Result{Job: job, State: StateFailed, Attempt: attempt, Diagnostic: Diagnostic{Code: DiagnosticFailed, Severity: SeverityError}, Err: err}
}
