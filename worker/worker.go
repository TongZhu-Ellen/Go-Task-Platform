package worker

import (
	"context"
	"fmt"
	"time"
	"log"

	"go-task-platform/job"
)

type Worker struct {
	repo      *job.Repository
	executors map[string]Executor
}

func NewWorker(repo *job.Repository) *Worker {
	return &Worker{
		repo:      repo,
		executors: make(map[string]Executor),
	}
}

func (w *Worker) Register(kind string, executor Executor) {
	if _, exists := w.executors[kind]; exists {
		panic("worker: executor already registered for kind " + kind)
	}

	w.executors[kind] = executor
}

func (w *Worker) Get(kind string) (Executor, bool) {
	executor, ok := w.executors[kind]
	return executor, ok
}

func (w *Worker) Run(ctx context.Context, queue string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		// Claim job
		j, err := w.repo.Claim(ctx, queue)
		if err != nil {
			log.Printf("worker: failed to claim job from queue %q: %v", queue, err)
			time.Sleep(PollInterval)
			continue
		}

		if j == nil {
			time.Sleep(PollInterval)
			continue
		}

		// Find executor
		executor, ok := w.executors[j.Kind]
		if !ok {
			err := fmt.Errorf("executor %q not found", j.Kind)
			failToExecuteJob(j.ID, j.Kind, err)

			if markErr := w.repo.MarkDiscarded(ctx, j.ID, j.Attempt, err) ; markErr != nil {
				failToMarkJobAs("discarded", j, markErr)
			}
			continue
		}

		// Execute job
		err = executor(ctx, j)
		if err != nil {
			failToExecuteJob(j.ID, j.Kind, err)

			// Discard if max attempts reached
			if j.Attempt >= job.MaxAttempts {
				if markErr := w.repo.MarkDiscarded(ctx, j.ID, j.Attempt, err) ; markErr != nil {
					failToMarkJobAs("discarded", j, markErr)
				}
				continue
			}

			
			delay := time.Duration(j.Attempt*j.Attempt*j.Attempt*j.Attempt) * time.Second // 仿照river的进行attempt的四次方。
			availableAt := time.Now().Add(delay)

			if markErr := w.repo.MarkAvail(ctx, j.ID, j.Attempt, availableAt, err); markErr != nil {
				failToMarkJobAs("available", j, markErr)
			}
			continue
		}

		// Mark complete
		if markErr := w.repo.MarkComplete(ctx, j.ID, j.Attempt); markErr != nil {
			failToMarkJobAs("completed", j, markErr)
		}
	}
}





func failToMarkJobAs(action string, job *job.Job, err error) {
	log.Printf("worker: failed to mark job as %s (job_id=%d, kind=%q): %v",
		action, job.ID, job.Kind, err)
}

func failToExecuteJob(jobID int64, kind string, err error) {
	log.Printf("worker: failed to execute job (job_id=%d, kind=%q): %v",
		jobID, kind, err)
}


