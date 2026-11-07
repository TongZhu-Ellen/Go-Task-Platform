package job

import "time"

const JobTimeout = 5 * time.Minute 

const MaxAttempts = 3 