package keepalive

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// workerEntry is a helper struct that holds the state of a worker and is used
// internally by the default Manager implementation.
type workerEntry struct {
	stopped bool
	worker  Worker
}

// managerImpl is the default implementation of the Manager interface.
type managerImpl struct {
	mutex   sync.Mutex
	workers map[string]*workerEntry
}

// NewManager creates a new Manager instance with no registered workers.
func NewManager() Manager {
	return &managerImpl{
		workers: map[string]*workerEntry{},
	}
}

// keepalive is a helper function that will restart a worker if it fails after a delay.
func (manager *managerImpl) keepalive(name string, stopped *bool, worker func() error) {
	backoff := NewExponentialBackoff(time.Second, time.Minute)
	for {
		if err := worker(); err != nil {
			log.Error().Err(err).Str("name", name).Msg("Worker failed")
		}

		if *stopped {
			log.Debug().Str("name", name).Msg("Worker stopped, not restarting")
			return
		} else {
			delay := backoff()
			log.Warn().Str("name", name).Dur("delay", delay).Msg("Worker stopped, restarting after delay")
			time.Sleep(delay)
		}
	}
}

// AddWorker registers a new worker with the given name. If a worker with the
// given name already exists, ErrorWorkerAlreadyExists is returned. The worker
// is not started until StartWorker is called.
func (manager *managerImpl) AddWorker(name string, worker Worker) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if _, ok := manager.workers[name]; ok {
		return ErrorWorkerAlreadyExists
	}

	manager.workers[name] = &workerEntry{
		worker: worker,
	}

	return nil
}

// Names returns a slice of all registered worker names.
func (manager *managerImpl) Names() []string {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	names := make([]string, 0, len(manager.workers))
	for name := range manager.workers {
		names = append(names, name)
	}

	return names
}

// StartWorker starts the worker with the given name. If the worker is not found,
// ErrorWorkerNotFound is returned. If multiple workers are given, they are
// started in the order given. If any worker fails to start, the error is logged
// and returned and no further workers are started. If any workers were already
// running prior to the error being encountered, they are not stopped. If the
// worker is already running, it is not restarted.
func (manager *managerImpl) StartWorker(names ...string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	for _, name := range names {
		definition, ok := manager.workers[name]
		if !ok {
			log.Error().Str("name", name).Msg("Worker not found")
			return ErrorWorkerNotFound
		}

		log.Debug().Str("name", name).Msg("Starting worker")
		definition.stopped = false
		go manager.keepalive(name, &definition.stopped, definition.worker.Run)
	}

	return nil
}

// StopWorker stops the worker with the given name. If the worker is not found,
// ErrorWorkerNotFound is returned. If multiple workers are given, they are
// stopped in the order given. If any worker fails to stop, the error is logged
// and returned and no further workers are stopped.
func (manager *managerImpl) StopWorker(names ...string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	for _, name := range names {
		definition, ok := manager.workers[name]
		if !ok {
			log.Error().Str("name", name).Msg("Worker not found")
			return ErrorWorkerNotFound
		}

		log.Debug().Str("name", name).Msg("Sending stop request to worker")
		definition.stopped = true

		log.Debug().Str("name", name).Msg("Stopping worker")
		if err := definition.worker.Stop(); err != nil {
			log.Error().Err(err).Str("name", name).Msg("Failed to stop worker")
			return err
		}
	}

	return nil
}

// StopAllWorkers stops all registered workers. This is a convenience method that
// calls StopWorker with all registered worker names. If any worker fails to stop,
// the error is logged and iteration continues to the next worker.
func (manager *managerImpl) StopAllWorkers() {
	names := manager.Names()

	for _, name := range names {
		if err := manager.StopWorker(name); err != nil {
			log.Error().Err(err).Str("name", name).Msg("Failed to stop worker")
		}
	}
}
