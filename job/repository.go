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


func (r *Repository) claim(ctx context.Context) (*Job, error) {
    query := `
        WITH claimed_job AS (
            SELECT id
            FROM jobs
            WHERE state = 'available'
              AND available_at <= NOW()
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
	// 我更倾向于用 = 而不是IN，因为我们设计上就是一次领取一个任务。

    jobp := &Job{}
    result := r.db.WithContext(ctx).Raw(query).Scan(jobp)

    if result.Error != nil {
        return nil, result.Error
    }

    if result.RowsAffected == 0 {
        return nil, nil
    }

    return jobp, nil
}