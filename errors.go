package keepalive

import "errors"

var (
	// ErrorWorkerAlreadyExists is returned when a worker is added with a name that already exists.
	ErrorWorkerAlreadyExists = errors.New("worker already exists")

	// ErrorWorkerNotFound is returned when a worker is stopped that doesn't exist.
	ErrorWorkerNotFound = errors.New("worker not found")
)
