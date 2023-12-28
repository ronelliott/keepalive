package keepalive_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ronelliott/keepalive"
)

func TestManager_Basic(t *testing.T) {
	manager := keepalive.NewManager()
	worker := &testWorker{}

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

func TestManager_Run_Error(t *testing.T) {
	manager := keepalive.NewManager()

	worker := &testWorker{runError: assert.AnError}

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

func TestManager_Stop_Error(t *testing.T) {
	manager := keepalive.NewManager()

	worker := &testWorker{stopError: assert.AnError}

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

func TestManager_StopAll(t *testing.T) {
	manager := keepalive.NewManager()

	worker1 := &testWorker{}
	worker2 := &testWorker{}

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
	stopCallCount int
	stopError     error
}

func (w *testWorker) Run() error {
	w.runCallCount++
	return w.runError
}

func (w *testWorker) Stop() error {
	w.stopCallCount++
	return w.stopError
}
