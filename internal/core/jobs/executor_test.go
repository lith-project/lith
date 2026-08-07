package jobs

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestExecutorRejectsInvalidClassBeforeWriter(t *testing.T) {
	var commits atomic.Int32
	executor := testExecutor(t, classifierFunc(func(context.Context, Job) (Classification, error) {
		return "invalid", nil
	}), writerFunc(func(context.Context, Commit) error { commits.Add(1); return nil }))
	result := executor.Execute(context.Background(), []Job{testJob(t, "one")})[0]
	if result.State != StateFailed || !errors.Is(result.Err, ErrInvalidClassification) || commits.Load() != 0 {
		t.Fatalf("result = %#v, commits = %d, want invalid-classification failure without commit", result, commits.Load())
	}
}

func TestExecutorParseClassDoesNotCommit(t *testing.T) {
	var commits atomic.Int32
	executor := testExecutor(t, classifierFunc(func(context.Context, Job) (Classification, error) {
		return ClassificationParse, nil
	}), writerFunc(func(context.Context, Commit) error { commits.Add(1); return nil }))
	result := executor.Execute(context.Background(), []Job{testJob(t, "one")})[0]
	if result.State != StateSucceeded || commits.Load() != 0 {
		t.Fatalf("result = %#v, commits = %d, want parse success without commit", result, commits.Load())
	}
}

func TestExecutorRetryClassification(t *testing.T) {
	var attempts atomic.Int32
	executor := testExecutorWithConfig(t, classifierFunc(func(context.Context, Job) (Classification, error) {
		if attempts.Add(1) < 3 {
			return "", NewTransientFailure(errors.New("temporary"))
		}
		return ClassificationParse, nil
	}), writerFunc(func(context.Context, Commit) error { return nil }), Config{Workers: 1, MaxAttempts: 3})
	result := executor.Execute(context.Background(), []Job{testJob(t, "one")})[0]
	if result.State != StateSucceeded || result.Attempt != 3 || attempts.Load() != 3 {
		t.Fatalf("result = %#v, classifier attempts = %d, want capped transient retry", result, attempts.Load())
	}
}

func TestExecutorDoesNotRetryDeterministicWriterFailure(t *testing.T) {
	var calls atomic.Int32
	executor := testExecutorWithConfig(t, validClassifier, writerFunc(func(context.Context, Commit) error {
		calls.Add(1)
		return NewDeterministicFailure(errors.New("invalid"))
	}), Config{Workers: 1, MaxAttempts: 3})
	result := executor.Execute(context.Background(), []Job{testJob(t, "one")})[0]
	if result.State != StateFailed || result.Attempt != 1 || calls.Load() != 1 {
		t.Fatalf("result = %#v, writer calls = %d, want one deterministic attempt", result, calls.Load())
	}
}

func TestExecutorRetriesIdempotentPartialWrite(t *testing.T) {
	var calls atomic.Int32
	var applied atomic.Int32
	var mu sync.Mutex
	seen := make(map[IdentityKey]struct{})
	writer := writerFunc(func(_ context.Context, commit Commit) error {
		calls.Add(1)
		mu.Lock()
		defer mu.Unlock()
		if _, ok := seen[commit.Job().IdentityKey()]; ok {
			return nil
		}
		seen[commit.Job().IdentityKey()] = struct{}{}
		applied.Add(1)
		return NewTransientFailure(errors.New("ack lost after write"))
	})
	executor := testExecutorWithConfig(t, validClassifier, writer, Config{Workers: 1, MaxAttempts: 3})
	result := executor.Execute(context.Background(), []Job{testJob(t, "one")})[0]
	if result.State != StateSucceeded || result.Attempt != 2 || calls.Load() != 2 || applied.Load() != 1 {
		t.Fatalf("result = %#v, calls = %d, applied = %d, want idempotent retry", result, calls.Load(), applied.Load())
	}
}

func TestExecutorCoalescesDuplicateIdentity(t *testing.T) {
	var calls atomic.Int32
	executor := testExecutor(t, validClassifier, writerFunc(func(context.Context, Commit) error { calls.Add(1); return nil }))
	job := testJob(t, "one")
	results := executor.Execute(context.Background(), []Job{job, job})
	if calls.Load() != 1 || results[0].State != StateSucceeded || results[1].State != StateCoalesced {
		t.Fatalf("calls = %d, results = %#v, want one commit and one coalesced result", calls.Load(), results)
	}
}

func TestExecutorCancellationCoversRunningAndQueued(t *testing.T) {
	started := make(chan struct{})
	classifier := classifierFunc(func(ctx context.Context, job Job) (Classification, error) {
		if job.IdentityKey().String() == "one" {
			close(started)
			<-ctx.Done()
			return "", ctx.Err()
		}
		return ClassificationStore, nil
	})
	var calls atomic.Int32
	executor := testExecutorWithConfig(t, classifier, writerFunc(func(context.Context, Commit) error { calls.Add(1); return nil }), Config{Workers: 1, MaxAttempts: 1})
	ctx, cancel := context.WithCancel(context.Background())
	resultsCh := make(chan []Result, 1)
	go func() { resultsCh <- executor.Execute(ctx, []Job{testJob(t, "one"), testJob(t, "two")}) }()
	<-started
	cancel()
	results := <-resultsCh
	if calls.Load() != 0 {
		t.Fatal("cancelled job reached writer")
	}
	for _, result := range results {
		if result.State != StateCancelled || result.Diagnostic.Code != DiagnosticCancelled {
			t.Fatalf("result = %#v, want cancellation", result)
		}
	}
}

func TestExecutorSerializesWriterCommits(t *testing.T) {
	ready := make(chan struct{}, 2)
	releaseClassify := make(chan struct{})
	firstEntered := make(chan struct{})
	secondEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	var calls atomic.Int32
	var releaseAllowed atomic.Bool
	var startedEarly atomic.Bool
	classifier := classifierFunc(func(context.Context, Job) (Classification, error) {
		ready <- struct{}{}
		<-releaseClassify
		return ClassificationStore, nil
	})
	writer := writerFunc(func(context.Context, Commit) error {
		if calls.Add(1) == 1 {
			close(firstEntered)
			<-releaseFirst
			return nil
		}
		if !releaseAllowed.Load() {
			startedEarly.Store(true)
		}
		close(secondEntered)
		return nil
	})
	executor := testExecutorWithConfig(t, classifier, writer, Config{Workers: 2, MaxAttempts: 1})
	resultsCh := make(chan []Result, 1)
	go func() {
		resultsCh <- executor.Execute(context.Background(), []Job{testJob(t, "one"), testJob(t, "two")})
	}()
	<-ready
	<-ready
	close(releaseClassify)
	<-firstEntered
	releaseAllowed.Store(true)
	close(releaseFirst)
	<-secondEntered
	results := <-resultsCh
	if startedEarly.Load() {
		t.Fatal("second writer entered before first commit was released")
	}
	for _, result := range results {
		if result.State != StateSucceeded {
			t.Fatalf("result = %#v, want success", result)
		}
	}
}

type classifierFunc func(context.Context, Job) (Classification, error)

func (f classifierFunc) Classify(ctx context.Context, job Job) (Classification, error) {
	return f(ctx, job)
}

var validClassifier = classifierFunc(func(context.Context, Job) (Classification, error) { return ClassificationStore, nil })

type writerFunc func(context.Context, Commit) error

func (f writerFunc) Commit(ctx context.Context, commit Commit) error { return f(ctx, commit) }

func testExecutor(t *testing.T, classifier Classifier, writer Writer) *Executor {
	return testExecutorWithConfig(t, classifier, writer, Config{Workers: 2, MaxAttempts: 1})
}

func testExecutorWithConfig(t *testing.T, classifier Classifier, writer Writer, config Config) *Executor {
	t.Helper()
	executor, err := NewExecutor(classifier, writer, config)
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}
	return executor
}

func testJob(t *testing.T, identity string) Job {
	t.Helper()
	job, err := NewJob(JobSpec{Kind: KindIndexBatch, IdentityKey: identity, Priority: PriorityIncremental})
	if err != nil {
		t.Fatalf("NewJob() error = %v", err)
	}
	return job
}
