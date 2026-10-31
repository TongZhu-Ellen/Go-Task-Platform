package job

import (
	"context"
	"time"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) create(ctx context.Context, reqp *CreateJobRequest) (*Job, error) {

	now := time.Now()
	queue := reqp.Queue
	if queue == "" {
		queue = "default"
	}
	jobp := &Job{
		CreatedAt:   now,
		Queue:       queue,
		Kind:        reqp.Kind,
		Args:        reqp.Args,
		State:       JobAvailable,
		AvailableAt: &now,
		Attempt:     0,
	}

	if err := s.repository.create(ctx, jobp); err != nil {
		return nil, err
	}

	return jobp, nil
}




func (s *Service) Claim(ctx context.Context, queue string) (*Job, error) {
	return s.repository.claim(ctx, queue)
}