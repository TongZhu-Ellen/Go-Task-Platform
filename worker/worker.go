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

        if err := ctx.Err(); err != nil { // 每次开始工作前询问ctx情况，是否已经cancel！
            return err
        }

        
		
		// claim job!
		job, err := w.repo.Claim(ctx, queue)
		if err != nil {
			log.Printf("worker: failed to claim job from queue %q: %v", queue, err)
			time.Sleep(PollInterval)
			continue
		} 
		if job == nil {
			time.Sleep(PollInterval)
			continue
		}

		
		// find	executor!
		executor, ok := w.executors[job.Kind]
		if !ok {
			err := fmt.Errorf("executor 404 %q", job.Kind)
			failToExecuteJob(job.ID, job.Kind, err)

			if markErr := w.repo.MarkDiscarded(ctx, job.ID, err); markErr != nil {
				failToMarkJobAs("discarded", job, markErr)
			}
			continue
		}


		// execute job! 
		err = executor(ctx, job)
		if err != nil {
			failToExecuteJob(job.ID, job.Kind, err)

			// 不能再重试了！扔掉！
			if job.Attempt >= MaxAttempts {
				if markErr := w.repo.MarkDiscarded(ctx, job.ID, err); markErr != nil {
					failToMarkJobAs("discarded", job, markErr)
				}
				continue
			}

			// 重试时间！
			delay := time.Duration(1<<job.Attempt) * time.Second // 指数退避！ 2^attempt seconds
			availableAt := time.Now().Add(delay)

			if markErr := w.repo.MarkRetry(ctx, job.ID, availableAt, err); markErr != nil {
				failToMarkJobAs("retryable", job, markErr)
			}
			continue
		}

		
		
		// 好了完成了！！！
		if markErr := w.repo.MarkComplete(ctx, job.ID); markErr != nil {
			failToMarkJobAs("completed", job, markErr)
	
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


