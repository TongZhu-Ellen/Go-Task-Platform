package main

import (
	"fmt"
	"log"
	
	"net/http"

    "go-task-platform/database"
    "go-task-platform/job"


)

const databaseURL = "postgres://ellenchung@127.0.0.1:5432/go_task_platform?sslmode=disable"

func main() {
	db, sqlDB, err := database.NewDB(databaseURL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
		
	}

	defer sqlDB.Close()

	fmt.Println("database connected successfully")

	repository := job.NewRepository(db)
	service := job.NewService(repository)
	handler := job.NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", handler.CreateJob)

	fmt.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}