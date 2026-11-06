package main

import (
	"fmt"
	"log"
	"context"

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
	w := worker.NewWorker(repository) // 现在啥也没有所以肯定找不到executor的。expected！

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
