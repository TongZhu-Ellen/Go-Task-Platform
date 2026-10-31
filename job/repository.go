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

	var queueID int32

	result := r.db.WithContext(ctx).
		Table("queue_kinds").
		Select("queue_id").
		Where("kind = ?", jobp.Kind).
		Scan(&queueID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		queueID = 1
	}

	jobp.QueueID = queueID

	return r.db.WithContext(ctx).Create(jobp).Error
}

func (r *Repository) claim(ctx context.Context, queueID int32) (*Job, error) {
	query := `
		WITH claimed_job AS (
			SELECT id
			FROM jobs
			WHERE state = 'available'
			  AND available_at <= NOW()
			  AND queue_id = ?
			ORDER BY available_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE jobs
		SET
			state = 'running',
			attempt = attempt + 1,
			last_attempt = NOW()
		WHERE id = (SELECT id FROM claimed_job)
		RETURNING *
	`

	jobp := &Job{}
	result := r.db.WithContext(ctx).Raw(query, queueID).Scan(jobp)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return jobp, nil
}
