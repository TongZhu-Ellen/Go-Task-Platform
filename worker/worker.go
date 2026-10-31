package worker

import (
	"context"
	"fmt"
	"go-task-platform/job"
)

type Handler func(ctx context.Context, j *job.Job) error

type Worker struct {
	service  *job.Service
	handlers map[string]Handler
}

func New(service *job.Service) *Worker {
	return &Worker{
		service:  service,
		handlers: make(map[string]Handler),
	}
}

func (w *Worker) Register(kind string, handler Handler) {
	w.handlers[kind] = handler
}

// 找到对应handler并执行！
func (w *Worker) Execute(ctx context.Context, j *job.Job) error {
	handler, ok := w.handlers[j.Kind]
	if !ok {
		return fmt.Errorf("no handler registered for job kind: %s", j.Kind)
	}

	return handler(ctx, j)
}
