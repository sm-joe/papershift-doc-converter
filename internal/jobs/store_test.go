package jobs

import (
	"testing"
	"time"
)

func TestStoreSetAndGet(t *testing.T) {
	store := NewStore()

	job := Job{
		ID:     "test-job",
		Status: StatusCompleted,
	}

	store.Set(Result{
		Job:       job,
		Output:    []byte("converted"),
		Filename:  "output.pdf",
		ExpiresAt: time.Now().UTC().Add(time.Minute),
	})

	result, err := store.Get("test-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Job.ID != "test-job" {
		t.Fatalf("expected test-job, got %s", result.Job.ID)
	}

	if string(result.Output) != "converted" {
		t.Fatalf(
			"expected converted, got %s",
			string(result.Output),
		)
	}
}

func TestStoreMissingJob(t *testing.T) {
	store := NewStore()

	_, err := store.Get("missing")

	if err != ErrJobNotFound {
		t.Fatalf(
			"expected ErrJobNotFound, got %v",
			err,
		)
	}
}

func TestStoreExpiredJob(t *testing.T) {
	store := NewStore()

	store.Set(Result{
		Job: Job{
			ID: "expired-job",
		},
		Output:    []byte("converted"),
		Filename:  "output.pdf",
		ExpiresAt: time.Now().UTC().Add(-time.Minute),
	})

	_, err := store.Get("expired-job")

	if err != ErrJobNotFound {
		t.Fatalf(
			"expected expired job to be unavailable, got %v",
			err,
		)
	}
}
