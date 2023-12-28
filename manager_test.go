package keepalive_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ronelliott/keepalive"
)

func TestManager_Basic(t *testing.T) {
	manager := keepalive.NewManager()
	worker := newTestWorker()

	err := manager.AddWorker("test", worker)
	assert.NoError(t, err, "Adding a worker should succeed")

	err = manager.AddWorker("test", worker)
	assert.Error(t, err, "Adding a worker with the same name should fail")

	err = manager.StartWorker("test")
	assert.NoError(t, err, "Starting a worker should succeed")

	time.Sleep(time.Millisecond * 100)

	err = manager.StopWorker("test")
	assert.NoError(t, err, "Stopping a worker should succeed")

	assert.Equal(t, 1, worker.runCallCount, "Worker should have been run once")
	assert.Equal(t, 1, worker.stopCallCount, "Worker should have been stopped once")
}

func TestManager_StartWorker_Error(t *testing.T) {
	manager := keepalive.NewManager()

	worker := newTestWorker()
	worker.runError = assert.AnError

	err := manager.AddWorker("test", worker)
	assert.NoError(t, err, "Adding a worker should succeed")

	err = manager.StartWorker("test")
	assert.NoError(t, err, "Starting a worker should succeed")

	time.Sleep(time.Second * 10)

	err = manager.StopWorker("test")
	assert.NoError(t, err, "Stopping a worker should succeed")

	assert.Equal(t, 4, worker.runCallCount, "Worker should have been run 4 times")
	assert.Equal(t, 1, worker.stopCallCount, "Worker should have been stopped once")
}

func TestManager_StartWorker_AlreadyRunning(t *testing.T) {
	manager := keepalive.NewManager()

	worker := newTestWorker()

	err := manager.AddWorker("test", worker)
	assert.NoError(t, err, "Adding a worker should succeed")

	err = manager.StartWorker("test")
	assert.NoError(t, err, "Starting a worker should succeed")

	err = manager.StartWorker("test")
	assert.ErrorIs(t, err, keepalive.ErrorWorkerAlreadyRunning,
		"Starting a worker that's already started should return an error")

	time.Sleep(time.Millisecond * 100)

	err = manager.StopWorker("test")
	assert.NoError(t, err, "Stopping a worker should succeed")

	assert.Equal(t, 1, worker.runCallCount, "Worker should have been run once")
	assert.Equal(t, 1, worker.stopCallCount, "Worker should have been stopped once")
}

func TestManager_StartWorker_NotFound(t *testing.T) {
	manager := keepalive.NewManager()

	err := manager.StartWorker("test")
	assert.ErrorIs(t, err, keepalive.ErrorWorkerNotFound,
		"Starting a worker that doesn't exist should return an error")
}

func TestManager_StopWorker_Error(t *testing.T) {
	manager := keepalive.NewManager()

	worker := newTestWorker()
	worker.stopError = assert.AnError

	err := manager.AddWorker("test", worker)
	assert.NoError(t, err, "Adding a worker should succeed")

	err = manager.StartWorker("test")
	assert.NoError(t, err, "Starting a worker should succeed")

	time.Sleep(time.Millisecond * 100)

	err = manager.StopWorker("test")
	assert.ErrorIs(t, err, assert.AnError, "Stopping a worker should fail")

	assert.Equal(t, 1, worker.runCallCount, "Worker should have been run once")
	assert.Equal(t, 1, worker.stopCallCount, "Worker should have been stopped once")
}

func TestManager_StopWorker_NotRunning(t *testing.T) {
	manager := keepalive.NewManager()

	worker := newTestWorker()

	err := manager.AddWorker("test", worker)
	assert.NoError(t, err, "Adding a worker should succeed")

	err = manager.StopWorker("test")
	assert.ErrorIs(t, err, keepalive.ErrorWorkerNotRunning,
		"Stopping a worker that's not running should return an error")

	assert.Equal(t, 0, worker.runCallCount, "Worker should not have been run")
	assert.Equal(t, 0, worker.stopCallCount, "Worker should not have been stopped")
}

func TestManager_StopWorker_NotFound(t *testing.T) {
	manager := keepalive.NewManager()

	err := manager.StopWorker("test")
	assert.ErrorIs(t, err, keepalive.ErrorWorkerNotFound,
		"Stopping a worker that doesn't exist should return an error")
}

func TestManager_StopAllWorkers(t *testing.T) {
	manager := keepalive.NewManager()

	worker1 := newTestWorker()
	worker2 := newTestWorker()

	err := manager.AddWorker("test1", worker1)
	assert.NoError(t, err, "Adding worker 1 should succeed")

	err = manager.AddWorker("test2", worker2)
	assert.NoError(t, err, "Adding worker 2 should succeed")

	err = manager.StartWorker("test1")
	assert.NoError(t, err, "Starting worker 1 should succeed")

	err = manager.StartWorker("test2")
	assert.NoError(t, err, "Starting worker 2 should succeed")

	time.Sleep(time.Millisecond * 100)

	manager.StopAllWorkers()

	assert.Equal(t, 1, worker1.runCallCount, "Worker 1 should have been run once")
	assert.Equal(t, 1, worker1.stopCallCount, "Worker 1 should have been stopped once")

	assert.Equal(t, 1, worker2.runCallCount, "Worker 2 should have been run once")
	assert.Equal(t, 1, worker2.stopCallCount, "Worker 2 should have been stopped once")
}

type testWorker struct {
	runCallCount  int
	runError      error
	stop          chan bool
	stopCallCount int
	stopError     error
}

func newTestWorker() *testWorker {
	return &testWorker{
		stop: make(chan bool, 1),
	}
}

func (w *testWorker) Run() error {
	w.runCallCount++
	if w.runError != nil {
		return w.runError
	}

	for {
		select {
		case <-w.stop:
			return nil
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (w *testWorker) Stop() error {
	w.stopCallCount++
	w.stop <- true
	return w.stopError
}
