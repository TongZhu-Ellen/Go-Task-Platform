package job

import (
	"context"
	"time"
	"log"

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
		WHERE id = (SELECT id FROM claimed_job)
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

func (r *Repository) MarkComplete(ctx context.Context, id int64, attempt int32) error {
	result := r.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ? AND state = ? AND attempt = ?", id, JobRunning, attempt).
		Updates(map[string]interface{}{
			"state":        JobCompleted,
			"available_at": nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.Printf("repository: lease lost, skip mark complete (job_id=%d, attempt=%d)", id, attempt)
	}
	return nil
}

func (r *Repository) MarkAvail(ctx context.Context, id int64, attempt int32, availableAt time.Time, lastError error) error {
	result := r.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ? AND state = ? AND attempt = ?", id, JobRunning, attempt).
		Updates(map[string]interface{}{
			"state":        JobAvailable,
			"available_at": availableAt,
			"last_error":   lastError.Error(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.Printf("repository: lease lost, skip mark available (job_id=%d, attempt=%d)", id, attempt)
	}
	return nil
}

func (r *Repository) MarkDiscarded(ctx context.Context, id int64, attempt int32, lastError error) error {
	result := r.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ? AND state = ? AND attempt = ?", id, JobRunning, attempt).
		Updates(map[string]interface{}{
			"state":        JobDiscarded,
			"available_at": nil,
			"last_error":   lastError.Error(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.Printf("repository: lease lost, skip mark discarded (job_id=%d, attempt=%d)", id, attempt)
	}
	return nil
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