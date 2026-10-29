package job

import (
	"encoding/json"
	"time"
)

type Job struct {
	ID        int64           `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	Kind      string          `json:"kind"`
	Args      json.RawMessage `json:"args"`

	State       JobState   `json:"state"` 
	AvailableAt *time.Time `json:"available_at,omitempty"`

	Attempt     int32      `json:"attempt"`
	LastAttempt *time.Time `json:"last_attempt,omitempty"`
	LastError   *string    `json:"last_error,omitempty"`
}

type JobState string

const (
	JobAvailable JobState = "available"
	JobRunning   JobState = "running"
	JobRetryable JobState = "retryable"
	JobCompleted JobState = "completed"
	JobDiscarded JobState = "discarded"
)

type CreateJobRequest struct {
	Kind string          `json:"kind"`
	Args json.RawMessage `json:"args"`
}
