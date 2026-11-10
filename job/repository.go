package job

import (
	"context"
	"time"

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

func (r *Repository) Claim(ctx context.Context, queue string) (*Job, error) {
	query := `
		WITH claimed_job AS (
			SELECT id
			FROM jobs
			WHERE queue = ?
			AND state = 'available' 
			AND available_at <= NOW()
	
			ORDER BY available_at ASC, created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE jobs
		SET
			state = 'running',
			attempt = attempt + 1,
			last_attempt = NOW()
		WHERE id = (SELECT id FROM claimed_job) -- 这里用 = 不用 IN 因为【一个】！
		RETURNING *
	`
	

	jobp := &Job{}
	result := r.db.WithContext(ctx).Raw(query, queue).Scan(jobp)

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	
	return jobp, nil
}

func (r *Repository) MarkComplete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ?", id).
		Update("state", JobCompleted).
		Error
}

func (r *Repository) MarkAvail(ctx context.Context, id int64, availableAt time.Time, lastError error) error {
	return r.db.WithContext(ctx).
        Model(&Job{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "state":        JobAvailable,
            "available_at": availableAt,
            "last_error":   lastError.Error(),
        }).
        Error

}

func (r *Repository) MarkDiscarded(ctx context.Context, id int64, lastError error) error {
	return r.db.WithContext(ctx).
        Model(&Job{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "state":        JobDiscarded,
			"last_error":   lastError.Error(),
			"available_at": nil,
        }).
        Error
}

func (r *Repository) RecoverExpiredJobs(ctx context.Context) error {
	query := `
		UPDATE jobs
		SET
			state = CASE
				WHEN attempt >= ? THEN ?
				ELSE ?
			END,
			available_at = CASE
				WHEN attempt >= ? THEN NULL
				ELSE NOW()
			END
		WHERE state = ?
		AND last_attempt < NOW() - (? * INTERVAL '1 second')
	`

	return r.db.WithContext(ctx).Exec(
		query,
		MaxAttempts,
		JobDiscarded,
		JobAvailable,
		MaxAttempts,
		JobRunning,
		JobTimeout.Seconds(),
	).Error
}