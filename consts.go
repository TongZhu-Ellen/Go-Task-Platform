package main

import (
	"time"
)

const DatabaseURL = "postgres://ellenchung@127.0.0.1:5432/go_task_platform?sslmode=disable"

// worker consts!
const WorkerAmount = 10
const PollInterval = time.Second
const MaxAttempts = 3 
