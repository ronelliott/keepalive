package keepalive

// Manager is an interface for managing workers. It provides methods for adding
// workers, starting and stopping workers, and enumerating registered workers.
type Manager interface {
	// AddWorker registers a new worker with the given name. If a worker with the
	// given name already exists, ErrorWorkerAlreadyExists is returned. The worker
	// is not started until StartWorker is called.
	AddWorker(string, Worker) error

	// Names returns a slice of all registered worker names.
	Names() []string

	// StartWorker starts the worker with the given name. If the worker is not found,
	// ErrorWorkerNotFound is returned. If multiple workers are given, they are
	// started in the order given. If any worker fails to start, the error is logged
	// and returned and no further workers are started. If any workers were already
	// running prior to the error being encountered, they are not stopped. If the
	// worker is already running, it is not restarted.
	StartWorker(...string) error

	// StopWorker stops the worker with the given name. If the worker is not found,
	// ErrorWorkerNotFound is returned. If multiple workers are given, they are
	// stopped in the order given. If any worker fails to stop, the error is logged
	// and returned and no further workers are stopped.
	StopWorker(...string) error

	// StopAllWorkers stops all registered workers. This is a convenience method that
	// calls StopWorker with all registered worker names. If any worker fails to stop,
	// the error is logged and iteration continues to the next worker.
	StopAllWorkers()
}

// Worker is an interface for a worker. It provides methods for starting and
// stopping the worker.
type Worker interface {
	// Run starts the worker. This method should block until the worker is stopped.
	Run() error

	// Stop stops the worker. This method should unblock the Run method.
	Stop() error
}
