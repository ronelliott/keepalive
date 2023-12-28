package keepalive

import "errors"

var (
	// ErrorWorkerAlreadyExists is returned when a worker is added with a name that already exists.
	ErrorWorkerAlreadyExists = errors.New("worker already exists")

	// ErrorWorkerAlreadyRunning is returned when a worker is started that is already running.
	ErrorWorkerAlreadyRunning = errors.New("worker already running")

	// ErrorWorkerNotRunning is returned when a worker is stopped that is not running.
	ErrorWorkerNotRunning = errors.New("worker not running")

	// ErrorWorkerNotFound is returned when a worker is stopped that doesn't exist.
	ErrorWorkerNotFound = errors.New("worker not found")
)
