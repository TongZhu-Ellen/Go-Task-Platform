package queue

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

func (r *Repository) create(ctx context.Context, name string, kinds []string, workerAmount int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qp := &Queue{}

		result := tx.Raw(`
			INSERT INTO queues (name, worker_amount, created_at)
			VALUES (?, ?, NOW())
			RETURNING *
		`, name, workerAmount).Scan(qp)

		if result.Error != nil {
			return result.Error
		}

		for _, kind := range kinds {
			if err := tx.Exec(`
				INSERT INTO queue_kinds (kind, queue_id)
				VALUES (?, ?)
			`, kind, qp.ID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
