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

func (s *Service) createJob(ctx context.Context, reqp *CreateJobRequest) (*Job, error) {
	jobp := &Job{
		CreatedAt: time.Now(),
		Kind:      reqp.Kind,
		Args:      reqp.Args,
		State:     JobAvailable,
		Attempt:   0,
	}

	if err := s.repository.create(ctx, jobp); err != nil {
		return nil, err
	}

	return jobp, nil
}