package worker

import (
	"context"

	"go-task-platform/job"
)

type Executor func(ctx context.Context, job *job.Job) error // 实际执行的函数类型！
