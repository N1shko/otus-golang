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
	var errCount int64
	var wg sync.WaitGroup

	taskPool := make(chan Task)
	defer func() {
		close(taskPool)
		wg.Wait()
	}()
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			for taskFromPool := range taskPool {
				if err := taskFromPool(); err != nil {
					atomic.AddInt64(&errCount, 1)
				}
			}
			wg.Done()
		}()
	}
	for _, task := range tasks {
		if atomic.LoadInt64(&errCount) >= int64(m) {
			return ErrErrorsLimitExceeded
		}
		taskPool <- task
	}
	return nil
}
