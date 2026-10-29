package job

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) create(ctx context.Context, jobp *Job) error {
	return r.db.WithContext(ctx).Create(jobp).Error
}