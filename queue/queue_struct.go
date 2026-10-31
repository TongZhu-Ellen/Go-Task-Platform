package queue

import "time"

type Queue struct {
	ID           int32
	Name         string
	WorkerAmount int
	CreatedAt    time.Time
}
