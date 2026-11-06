package worker

import (
	"context"
	"fmt"

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

        if err := ctx.Err(); err != nil { // 每次开始工作前询问ctx情况，是否已经cancel！
            return err
        }

        
		
		job, err := w.repo.Claim(ctx, queue)
		if err != nil {
			log.Printf("worker: failed to claim job from queue %q: %v", queue, err)
			time.Sleep(pollInterval)
			continue
		} 
		if job == nil {
			time.Sleep(pollInterval)
			continue
		}

		var 

		executor, ok := w.executors[job.Kind]
		if !ok {
			log.Printf("worker: no executor registered for job kind %q (job_id=%d)", job.Kind, job.ID)
			if err := w.repo.Discard(ctx, job.ID); err != nil {
				log.Printf("worker: failed to mark job as discarded (job_id=%d): %v", job.ID, err)
			}
			continue
		}

		err = executor(ctx, job)
		if err != nil {
			log.Printf("worker: job execution failed (job_id=%d, kind=%q): %v", job.ID, job.Kind, err)

			// TODO: Retry / Discard
			continue
		}

		if err := w.repo.Complete(ctx, job.ID); err != nil {
			log.Printf("worker: failed to mark job as completed (job_id=%d): %v", job.ID, err)
			continue
		}

    }
}


