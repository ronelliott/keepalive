package keepalive

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// workerEntry is a helper struct that holds the state of a worker and is used
// internally by the default Manager implementation.
type workerEntry struct {
	stop    chan bool
	stopped chan bool
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
func (manager *managerImpl) keepalive(name string, stop <-chan bool, stopped chan<- bool, worker func() error) {
	defer func() {
		stopped <- true
		close(stopped)
		log.Debug().Str("name", name).Msg("Exited keepalive loop")
	}()
	backoff := NewExponentialBackoff(time.Second, time.Minute)
	for {
		select {
		case <-stop:
			log.Debug().Str("name", name).Msg("Worker stopped by request, not restarting")
			return
		default:
			log.Debug().Str("name", name).Msg("Calling worker function")
			if err := worker(); err != nil {
				log.Error().Err(err).Str("name", name).Msg("Worker failed")
			}
		}

		delay := backoff()
		log.Warn().Str("name", name).Dur("delay", delay).Msg("Worker stopped, restarting after delay")
		time.Sleep(delay)
	}
}

// AddWorker registers a new worker with the given name. If a worker with the
// given name already exists, ErrorWorkerAlreadyExists is returned. The worker
// is not started until StartWorker is called.
func (manager *managerImpl) AddWorker(name string, worker Worker) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if _, ok := manager.workers[name]; ok {
		log.Error().Str("name", name).Msg("Cannot add worker, worker already exists")
		return ErrorWorkerAlreadyExists
	}

	manager.workers[name] = &workerEntry{
		worker: worker,
	}

	log.Debug().Str("name", name).Msg("Worker added")
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
			log.Error().Str("name", name).Msg("Cannot start worker, worker not found")
			return ErrorWorkerNotFound
		}

		if definition.stop != nil {
			log.Debug().Str("name", name).Msg("Cannot start worker, worker already running")
			return ErrorWorkerAlreadyRunning
		}

		log.Debug().Str("name", name).Msg("Starting worker")
		definition.stop = make(chan bool, 1)
		definition.stopped = make(chan bool, 1)
		go manager.keepalive(name, definition.stop, definition.stopped, definition.worker.Run)
		log.Debug().Str("name", name).Msg("Started worker")
	}

	return nil
}

// StopWorker stops the worker with the given name and waits for it to stop. If
// the worker is not found, ErrorWorkerNotFound is returned. If multiple workers
// are given, they are stopped in the order given. If any worker fails to stop,
// the error is logged and returned and no further workers are stopped, however
// any workers already in the process of stopping will be waited for.
func (manager *managerImpl) StopWorker(names ...string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	wait := sync.WaitGroup{}
	defer wait.Wait()
	for _, name := range names {
		definition, ok := manager.workers[name]
		if !ok {
			log.Error().Str("name", name).Msg("Cannot stop worker, worker not found")
			return ErrorWorkerNotFound
		}

		if definition.stop == nil {
			log.Debug().Str("name", name).Msg("Cannot stop worker, worker not running")
			return ErrorWorkerNotRunning
		}

		log.Debug().Str("name", name).Msg("Sending stop request to worker")
		definition.stop <- true

		wait.Add(1)
		go func(name string, stopped <-chan bool) {
			defer wait.Done()
			log.Debug().Str("name", name).Msg("Waiting for worker to stop")
			<-stopped
			log.Debug().Str("name", name).Msg("Worker stopped")
		}(name, definition.stopped)

		log.Debug().Str("name", name).Msg("Stopping worker")
		if err := definition.worker.Stop(); err != nil {
			log.Error().Err(err).Str("name", name).Msg("Failed to stop worker")
			return err
		}
		log.Debug().Str("name", name).Msg("Worker stopped")

		close(definition.stop)
		definition.stop = nil
		definition.stopped = nil
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
