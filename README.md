# keepalive

![Build Status](https://github.com/ronelliott/keepalive/actions/workflows/master.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ronelliott/keepalive)](https://goreportcard.com/report/github.com/ronelliott/keepalive)
[![Coverage Status](https://coveralls.io/repos/github/ronelliott/keepalive/badge.svg?branch=master)](https://coveralls.io/github/ronelliott/keepalive?branch=master)
[![Go Reference](https://pkg.go.dev/badge/github.com/ronelliott/keepalive.svg)](https://pkg.go.dev/github.com/ronelliott/keepalive)

keepalive is a lightweight goroutine manager that keeps a goroutine running and restarts it after errors.

## Manager

A manager manages one or more workers. Workers can be registered with an identifier by calling the `AddWorker` method on a manager instance. Once registered, a worker can then be started and stopped using the `StartWorker` and `StopWorker` methods respectively.

If a worker encounters an error, it will be restarted after a delay that is increased exponentially.

## Worker

Workers must implement the `Worker` interface, which requires implementing the `Run` and `Stop` methods. When `Run` is called on a worker, it must block until `Stop` is called on that worker. If an error is encountered by a worker during normal operation, it should be returned from the `Run` method.

## Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/ronelliott/keepalive"
)

type myWorker struct {
	stopped bool
}

func (service *myWorker) Run() error {
	service.stopped = false
	for {
		if service.stopped {
			break
		}

		fmt.Println("cycle")
		time.Sleep(time.Second * 5)
	}

	return nil
}

func (service *myWorker) Stop() error {
	service.stopped = true
	return nil
}

func main() {
	manager := keepalive.NewManager()

	manager.AddWorker("worker1", &myWorker{})
	manager.AddWorker("worker2", &myWorker{})

	manager.StartWorker("worker1", "worker2")
}
```
