package main

import (
	"fmt"
	"log"
	"context"
	"errors"

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


	// worker启动
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




	// HTTP 阻塞执行。

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", handler.Create)

	fmt.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))




	

}
