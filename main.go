package main

import (
	"fmt"
	"log"
	"os"

	"go-task-platform/internal/database"
)

const databaseURL = "postgres://ellenchung@127.0.0.1:5432/go_task_platform?sslmode=disable"

func main() {
	db, sqlDB, err := database.NewDB(databaseURL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
		
	}

	defer sqlDB.Close()

	fmt.Println("database connected successfully")

	// 后续使用 db 执行 GORM 操作。
	_ = db
}