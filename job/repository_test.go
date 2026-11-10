package job


import (
    "testing"
    "context"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)


const testDBURL = "postgres://ellenchung@127.0.0.1:5432/go_task_platform_test?sslmode=disable"


func TestClaimConcurrent(t *testing.T) {
    db, err := gorm.Open(postgres.Open(testDBURL), &gorm.Config{})
    if err != nil {
        t.Fatal(err)
    }

    repo := NewRepository(db)
    ctx := context.Background()

    // 手动构造
    now := time.Now()
    testJob := &Job{
        Kind:        "test",
        Queue:       "default",
        State:       JobAvailable,
        AvailableAt:  &now,
    }

    // 插进去
    if err := repo.create(ctx, testJob); err != nil {
        t.Fatal(err)
    }
    defer db.Exec("DELETE FROM jobs WHERE id = ?", testJob.ID)

    // 制造worker抢！
    const workers = 10

    start := make(chan struct{})
    results := make(chan *Job, workers)
    errs := make(chan error, workers)

    for i := 0; i < workers; i++ {
        go func() {
            <-start

            claimed, err := repo.Claim(ctx, "default")
            if err != nil {
                errs <- err
                return
            }

            results <- claimed
        }()
    }

    close(start)



    var met bool

    for i := 0; i < workers; i++ {
        select {
        case err := <-errs:
            t.Fatal(err)
        case claimed := <-results:

            if claimed == nil {
                continue
            }

            if met {
                t.Fatal("job was claimed more than once")
                
            } 

            met = true
            if claimed.Attempt != 1 {
                t.Fatalf("expected attempt 1, got %d", claimed.Attempt)
            }

            
        }
    }


    if !met {
        t.Fatal("expected one successful claim, got none")
    }

    

}


func TestRecoverExpiredJobs(t *testing.T) {
    db, err := gorm.Open(postgres.Open(testDBURL), &gorm.Config{})
    if err != nil {
        t.Fatal(err)
    }

    repo := NewRepository(db)
    ctx := context.Background()

    lastAttempt := time.Now().Add(-2 * JobTimeout)
    testJob := &Job{
        Kind:        "test",
        Queue:       "default",
        State:       JobRunning,
        AvailableAt: nil,
        Attempt:     1,
        LastAttempt: &lastAttempt,
    }

    if err := repo.create(ctx, testJob); err != nil {
        t.Fatal(err)
    }
    defer db.Exec("DELETE FROM jobs WHERE id = ?", testJob.ID)

    if err := repo.RecoverExpiredJobs(ctx); err != nil {
        t.Fatal(err)
    }

    jp := &Job{}
    if err := db.First(jp, testJob.ID).Error; err != nil {
        t.Fatal(err)
    }

    if jp.State != JobAvailable {
        t.Fatalf("expected state %s, got %s", JobAvailable, jp.State)
    }

    if jp.AvailableAt == nil {
        t.Fatal("expected available_at to be set")
    }

}


func TestRecoverExpiredJobsDiscarded(t *testing.T) {
	db, err := gorm.Open(postgres.Open(testDBURL), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	ctx := context.Background()

	lastAttempt := time.Now().Add(-2 * JobTimeout)
	testJob := &Job{
		Kind:        "test",
		Queue:       "default",
		State:       JobRunning,
		AvailableAt: nil,
		Attempt:     MaxAttempts,
		LastAttempt: &lastAttempt,
	}

	if err := repo.create(ctx, testJob); err != nil {
		t.Fatal(err)
	}
	defer db.Exec("DELETE FROM jobs WHERE id = ?", testJob.ID)

	if err := repo.RecoverExpiredJobs(ctx); err != nil {
		t.Fatal(err)
	}

	jp := &Job{}
	if err := db.First(jp, testJob.ID).Error; err != nil {
		t.Fatal(err)
	}

	if jp.State != JobDiscarded {
		t.Fatalf("expected state %s, got %s", JobDiscarded, jp.State)
	}

	if jp.AvailableAt != nil {
		t.Fatal("expected available_at to be nil")
	}
}


