package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	var errCount = int32(0)
	var wg sync.WaitGroup

	taskPool := make(chan Task)
	defer close(taskPool)
	for i := 0; i < n; i++ {
		go func() {
			for taskFromPool := range taskPool {
				wg.Add(1)
				if err := taskFromPool(); err != nil {
					atomic.AddInt32(&errCount, 1)
				}
				wg.Done()
			}
		}()
	}
	for _, task := range tasks {
		if atomic.LoadInt32(&errCount) >= int32(m) {
			return ErrErrorsLimitExceeded
		}
		taskPool <- task
	}
	wg.Wait()
	return nil
}
