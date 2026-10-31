package queue

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(
	ctx context.Context,
	name string,
	kinds []string,
	maxWorkers int,
) error {

	// TODO: workerpool!!!
	return s.repository.create(ctx, name, kinds, maxWorkers)
}
