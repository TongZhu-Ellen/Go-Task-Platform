package main

import (
	"fmt"
	"log"
	"context"
	"errors"
	"time"

	"net/http"

	"go-task-platform/database"
	"go-task-platform/job"
	"go-task-platform/worker"
)

func main() {
	db, sqlDB, err := database.NewDB(DatabaseURL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)

	}

	defer sqlDB.Close()
	fmt.Println("database connected successfully")



	// job crud-in 启动。
	repository := job.NewRepository(db)
	service := job.NewService(repository)
	handler := job.NewHandler(service)
	
	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", handler.Create)

	

	

	// Worker
	w := worker.NewWorker(repository)
	w.Register("expected-success", func(ctx context.Context, j *job.Job) error {
		fmt.Printf("job %d executed successfully\n", j.ID)
		return nil
	})
	w.Register("expected-failure", func(ctx context.Context, j *job.Job) error {
		fmt.Printf("job %d execution failed\n", j.ID)
		return errors.New("exemplary failure")
	})

	ctx := context.Background()
	for i := 0; i < WorkerAmount; i++ {
		go w.Run(ctx, "default")
	}

	// Recovery
	go runRecovery(ctx, repository)



	// 主线程开启
	fmt.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))


	

}




func runRecovery(ctx context.Context, repo *job.Repository) {
    ticker := time.NewTicker(RecoveryInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := repo.RecoverExpiredJobs(ctx); err != nil {
                log.Printf("recovery failed: %v", err)
            }
        }
    }
}
